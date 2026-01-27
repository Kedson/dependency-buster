//! HTTP/SSE Transport for MCP
//! Provides HTTP endpoints alongside stdio transport

#![allow(dead_code)]

use anyhow::Result;
use hyper::{body::Bytes, Request, Response, StatusCode, Method};
use http_body_util::{BodyExt, Full};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::sync::{broadcast, RwLock};

use super::auth::{validate_auth, Credentials, RequestContext};

/// HTTP/SSE configuration
#[derive(Debug, Clone)]
pub struct HttpConfig {
    pub port: u16,
    pub host: String,
    pub base_path: String,
    pub cors_origins: Vec<String>,
}

impl Default for HttpConfig {
    fn default() -> Self {
        Self {
            port: 3000,
            host: "127.0.0.1".to_string(),
            base_path: "/api/mcp".to_string(),
            cors_origins: vec!["*".to_string()],
        }
    }
}

/// JSON-RPC request
#[derive(Debug, Deserialize)]
struct JsonRpcRequest {
    jsonrpc: String,
    id: Option<Value>,
    method: String,
    params: Option<Value>,
}

/// JSON-RPC response
#[derive(Debug, Serialize)]
struct JsonRpcResponse {
    jsonrpc: String,
    id: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    result: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<Value>,
}

/// SSE event
#[derive(Debug, Clone, Serialize)]
pub struct SseEvent {
    pub event: String,
    pub data: Value,
}

/// Request handler type
pub type RequestHandler = Arc<
    dyn Fn(String, Value, RequestContext) -> std::pin::Pin<Box<dyn std::future::Future<Output = Result<Value>> + Send>>
        + Send
        + Sync,
>;

/// HTTP Transport state
pub struct HttpTransport {
    config: HttpConfig,
    handler: RequestHandler,
    sse_tx: broadcast::Sender<SseEvent>,
    client_count: Arc<RwLock<u64>>,
}

impl HttpTransport {
    /// Create new HTTP transport
    pub fn new<F, Fut>(handler: F, config: Option<HttpConfig>) -> Self
    where
        F: Fn(String, Value, RequestContext) -> Fut + Send + Sync + 'static,
        Fut: std::future::Future<Output = Result<Value>> + Send + 'static,
    {
        let (sse_tx, _) = broadcast::channel(100);
        
        let handler: RequestHandler = Arc::new(move |method, params, ctx| {
            let fut = handler(method, params, ctx);
            Box::pin(fut)
        });

        Self {
            config: config.unwrap_or_default(),
            handler,
            sse_tx,
            client_count: Arc::new(RwLock::new(0)),
        }
    }

    /// Broadcast event to all SSE clients
    pub fn broadcast(&self, event: &str, data: Value) {
        let _ = self.sse_tx.send(SseEvent {
            event: event.to_string(),
            data,
        });
    }

    /// Start HTTP server (simplified version using hyper directly)
    pub async fn start(self: Arc<Self>) -> Result<()> {
        use hyper::{
            body::Bytes,
            server::conn::http1,
            service::service_fn,
            Method, Request, Response, StatusCode,
        };
        use hyper_util::rt::TokioIo;
        use http_body_util::{BodyExt, Full};
        use tokio::net::TcpListener;

        let addr: SocketAddr = format!("{}:{}", self.config.host, self.config.port)
            .parse()
            .unwrap();

        let listener = TcpListener::bind(addr).await?;
        
        eprintln!(
            "HTTP/SSE transport listening on http://{}{}",
            addr, self.config.base_path
        );
        eprintln!(
            "SSE endpoint: http://{}{}/sse",
            addr, self.config.base_path
        );

        loop {
            let (stream, _) = listener.accept().await?;
            let io = TokioIo::new(stream);
            let transport = Arc::clone(&self);

            tokio::spawn(async move {
                let service = service_fn(move |req| {
                    let transport = Arc::clone(&transport);
                    async move { transport.handle_request(req).await }
                });

                if let Err(e) = http1::Builder::new()
                    .serve_connection(io, service)
                    .await
                {
                    eprintln!("HTTP error: {}", e);
                }
            });
        }
    }

    async fn handle_request(
        &self,
        req: Request<hyper::body::Incoming>,
    ) -> Result<Response<Full<Bytes>>, hyper::Error> {
        let path = req.uri().path().to_string();
        let method = req.method().clone();

        // CORS headers
        let mut response_headers = vec![
            ("Access-Control-Allow-Origin", "*"),
            ("Access-Control-Allow-Methods", "GET, POST, OPTIONS"),
            ("Access-Control-Allow-Headers", "Content-Type, Authorization"),
        ];

        // Handle preflight
        if method == Method::OPTIONS {
            let mut response = Response::new(Full::new(Bytes::new()));
            *response.status_mut() = StatusCode::NO_CONTENT;
            for (key, value) in response_headers {
                response.headers_mut().insert(
                    hyper::header::HeaderName::from_static(key),
                    hyper::header::HeaderValue::from_static(value),
                );
            }
            return Ok(response);
        }

        // Route handling
        let (status, body) = if path == format!("{}/v1", self.config.base_path)
            || path == format!("{}/v1/", self.config.base_path)
        {
            if method == Method::POST {
                self.handle_json_rpc(req).await
            } else {
                (StatusCode::METHOD_NOT_ALLOWED, json!({"error": "Method not allowed"}))
            }
        } else if path == format!("{}/health", self.config.base_path) {
            self.handle_health()
        } else if path == format!("{}/info", self.config.base_path) {
            self.handle_info()
        } else {
            (StatusCode::NOT_FOUND, json!({"error": "Not found"}))
        };

        let body_str = serde_json::to_string(&body).unwrap_or_default();
        let mut response = Response::new(Full::new(Bytes::from(body_str)));
        *response.status_mut() = status;
        response.headers_mut().insert(
            hyper::header::CONTENT_TYPE,
            hyper::header::HeaderValue::from_static("application/json"),
        );
        for (key, value) in response_headers {
            response.headers_mut().insert(
                hyper::header::HeaderName::from_static(key),
                hyper::header::HeaderValue::from_static(value),
            );
        }

        Ok(response)
    }

    async fn handle_json_rpc(
        &self,
        req: Request<hyper::body::Incoming>,
    ) -> (StatusCode, Value) {
        // Extract headers before consuming body
        let headers: HashMap<String, String> = req
            .headers()
            .iter()
            .filter_map(|(k, v)| {
                v.to_str().ok().map(|v| (k.to_string(), v.to_string()))
            })
            .collect();

        // Read body
        let body_bytes = match req.collect().await {
            Ok(collected) => collected.to_bytes(),
            Err(_) => return (StatusCode::BAD_REQUEST, json!({"error": "Failed to read body"})),
        };

        // Parse JSON-RPC request
        let request: JsonRpcRequest = match serde_json::from_slice(&body_bytes) {
            Ok(r) => r,
            Err(_) => return (StatusCode::BAD_REQUEST, json!({"error": "Invalid JSON"})),
        };

        // Validate auth
        let credentials = validate_auth(&request.method, &headers)
            .unwrap_or_else(|_| Credentials::anonymous());
        let ctx = RequestContext::new(credentials);

        // Call handler
        let params = request.params.unwrap_or(json!({}));
        let result = (self.handler)(request.method, params, ctx).await;

        match result {
            Ok(value) => (
                StatusCode::OK,
                json!({
                    "jsonrpc": "2.0",
                    "id": request.id,
                    "result": value
                }),
            ),
            Err(e) => (
                StatusCode::OK,
                json!({
                    "jsonrpc": "2.0",
                    "id": request.id,
                    "error": {
                        "code": -32603,
                        "message": e.to_string()
                    }
                }),
            ),
        }
    }

    fn handle_health(&self) -> (StatusCode, Value) {
        (
            StatusCode::OK,
            json!({
                "status": "healthy",
                "timestamp": chrono::Utc::now().to_rfc3339()
            }),
        )
    }

    fn handle_info(&self) -> (StatusCode, Value) {
        (
            StatusCode::OK,
            json!({
                "name": "dpb-mcp",
                "version": "1.0.0",
                "protocols": ["stdio", "http", "sse"],
                "endpoints": {
                    "http": format!("{}/v1", self.config.base_path),
                    "sse": format!("{}/v1/sse", self.config.base_path),
                    "health": format!("{}/health", self.config.base_path)
                }
            }),
        )
    }
}
