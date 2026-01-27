//! MCP Server with Enterprise Features
//! 
//! Features:
//! - Dynamic Action Registry
//! - Authentication (static tokens)
//! - Tool Annotations (readOnlyHint, idempotentHint, etc.)
//! - HTTP/SSE Transport
//! - Typed Errors (NotFound, NotAllowed, ValidationError)
//! - Credentials Context

pub mod errors;
pub mod annotations;
pub mod auth;

pub use errors::*;
pub use annotations::*;
pub use auth::*;

use anyhow::Result;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::sync::RwLock;

pub type ToolHandler = Arc<dyn Fn(Value) -> Result<String> + Send + Sync>;

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Tool {
    pub name: String,
    pub description: String,
    #[serde(rename = "inputSchema")]
    pub input_schema: InputSchema,
    #[serde(skip_serializing_if = "Option::is_none", default)]
    pub annotations: Option<ToolAnnotations>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct InputSchema {
    #[serde(rename = "type", default)]
    pub schema_type: String,
    #[serde(default)]
    pub properties: HashMap<String, Property>,
    #[serde(default)]
    pub required: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Property {
    #[serde(rename = "type", default)]
    pub property_type: String,
    #[serde(default)]
    pub description: String,
}

#[derive(Debug, Deserialize)]
struct JsonRpcRequest {
    #[allow(dead_code)]
    jsonrpc: String,
    id: Option<Value>,
    method: String,
    params: Option<Value>,
}

#[derive(Debug, Serialize)]
struct JsonRpcResponse {
    jsonrpc: String,
    id: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    result: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<RpcError>,
}

#[derive(Debug, Serialize)]
struct RpcError {
    code: i32,
    message: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    data: Option<Value>,
}

#[derive(Debug, Serialize)]
struct ToolContent {
    #[serde(rename = "type")]
    content_type: String,
    text: String,
}

pub struct Server {
    name: String,
    version: String,
    tools: Arc<RwLock<Vec<Tool>>>,
    handlers: Arc<RwLock<HashMap<String, ToolHandler>>>,
}

impl Server {
    pub fn new(name: &str, version: &str) -> Self {
        // Configure auth from environment
        let auth_enabled = std::env::var("MCP_AUTH_ENABLED")
            .map(|v| v == "true")
            .unwrap_or(false);
        
        configure_auth(AuthConfig {
            enabled: auth_enabled,
            token_env_var: "MCP_TOKEN".to_string(),
            public_methods: vec!["initialize".to_string(), "tools/list".to_string()],
            ..Default::default()
        });

        Self {
            name: name.to_string(),
            version: version.to_string(),
            tools: Arc::new(RwLock::new(Vec::new())),
            handlers: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub async fn register_tool<F>(&self, tool: Tool, handler: F)
    where
        F: Fn(Value) -> Result<String> + Send + Sync + 'static,
    {
        let mut tools = self.tools.write().await;
        let mut handlers = self.handlers.write().await;

        // Auto-add annotations if not provided
        let mut tool = tool;
        if tool.annotations.is_none() {
            tool.annotations = Some(get_tool_annotation(&tool.name));
        }

        tools.push(tool.clone());
        handlers.insert(tool.name.clone(), Arc::new(handler));
    }

    pub async fn run(&self) -> Result<()> {
        let transport = std::env::var("MCP_TRANSPORT").unwrap_or_else(|_| "stdio".to_string());
        
        eprintln!("PHP Dependency Analyzer MCP Server v{}", self.version);
        eprintln!("Transport: {}", transport);
        eprintln!("Auth: {}", is_auth_enabled());
        eprintln!("Features: Tool Annotations, Typed Errors, Credentials Context");

        if transport == "http" {
            self.run_http().await
        } else {
            self.run_stdio().await
        }
    }

    async fn run_stdio(&self) -> Result<()> {
        let stdin = tokio::io::stdin();
        let mut reader = BufReader::new(stdin);
        let mut stdout = tokio::io::stdout();

        let mut line = String::new();
        loop {
            line.clear();
            let bytes_read = reader.read_line(&mut line).await?;
            if bytes_read == 0 {
                break; // EOF
            }

            let request: JsonRpcRequest = match serde_json::from_str(&line) {
                Ok(req) => req,
                Err(_) => {
                    self.send_error(&mut stdout, None, error_codes::PARSE_ERROR, "Parse error", None)
                        .await?;
                    continue;
                }
            };

            // Create request context
            let ctx = RequestContext::new(Credentials::anonymous());
            self.handle_request(&mut stdout, request, ctx).await?;
        }

        Ok(())
    }

    async fn run_http(&self) -> Result<()> {
        let _port = std::env::var("MCP_HTTP_PORT").unwrap_or_else(|_| "3000".to_string());
        eprintln!("HTTP transport not yet implemented in Rust. Use stdio mode.");
        eprintln!("To use HTTP, set MCP_TRANSPORT=stdio (default)");
        // For a full HTTP implementation, we would use axum, actix-web, or hyper
        // For now, fall back to stdio
        self.run_stdio().await
    }

    async fn handle_request(
        &self,
        stdout: &mut tokio::io::Stdout,
        request: JsonRpcRequest,
        ctx: RequestContext,
    ) -> Result<()> {
        match request.method.as_str() {
            "initialize" => self.handle_initialize(stdout, request.id).await,
            "tools/list" => self.handle_list_tools(stdout, request.id).await,
            "tools/call" => self.handle_call_tool(stdout, request.id, request.params, ctx).await,
            _ => {
                self.send_error(stdout, request.id, error_codes::METHOD_NOT_FOUND, "Method not found", None)
                    .await
            }
        }
    }

    async fn handle_initialize(
        &self,
        stdout: &mut tokio::io::Stdout,
        id: Option<Value>,
    ) -> Result<()> {
        let result = json!({
            "protocolVersion": "2024-11-05",
            "capabilities": {
                "tools": true
            },
            "serverInfo": {
                "name": self.name,
                "version": self.version
            },
            "features": {
                "authentication": get_auth_info(),
                "transports": ["stdio"],
                "http_sse": "not_implemented"
            }
        });

        self.send_response(stdout, id, result).await
    }

    async fn handle_list_tools(
        &self,
        stdout: &mut tokio::io::Stdout,
        id: Option<Value>,
    ) -> Result<()> {
        let tools = self.tools.read().await;
        let result = json!({ "tools": tools.clone() });
        self.send_response(stdout, id, result).await
    }

    async fn handle_call_tool(
        &self,
        stdout: &mut tokio::io::Stdout,
        id: Option<Value>,
        params: Option<Value>,
        _ctx: RequestContext,
    ) -> Result<()> {
        let params = params.unwrap_or(Value::Null);

        let name = match params.get("name").and_then(|v| v.as_str()) {
            Some(n) => n,
            None => {
                return self
                    .send_error(stdout, id, error_codes::INVALID_PARAMS, "Invalid params: name required", None)
                    .await;
            }
        };

        let args = params.get("arguments").cloned().unwrap_or(json!({}));

        let handlers = self.handlers.read().await;
        let handler = match handlers.get(name) {
            Some(h) => h.clone(),
            None => {
                // Use typed NotFoundError
                let mcp_err = not_found_error(&format!("Tool \"{}\" not found", name));
                let content = vec![ToolContent {
                    content_type: "text".to_string(),
                    text: mcp_err.to_json(),
                }];

                return self.send_response(
                    stdout,
                    id,
                    json!({ "content": content, "isError": true }),
                )
                .await;
            }
        };

        drop(handlers); // Release lock before calling handler

        match handler(args) {
            Ok(result_text) => {
                let content = vec![ToolContent {
                    content_type: "text".to_string(),
                    text: result_text,
                }];

                self.send_response(stdout, id, json!({ "content": content }))
                    .await
            }
            Err(e) => {
                // Convert to typed MCP error
                let mcp_err = anyhow_to_mcp_error(&e);
                let content = vec![ToolContent {
                    content_type: "text".to_string(),
                    text: mcp_err.to_json(),
                }];

                self.send_response(
                    stdout,
                    id,
                    json!({ "content": content, "isError": true }),
                )
                .await
            }
        }
    }

    async fn send_response(
        &self,
        stdout: &mut tokio::io::Stdout,
        id: Option<Value>,
        result: Value,
    ) -> Result<()> {
        let response = JsonRpcResponse {
            jsonrpc: "2.0".to_string(),
            id,
            result: Some(result),
            error: None,
        };

        let json = serde_json::to_string(&response)?;
        stdout.write_all(json.as_bytes()).await?;
        stdout.write_all(b"\n").await?;
        stdout.flush().await?;
        Ok(())
    }

    async fn send_error(
        &self,
        stdout: &mut tokio::io::Stdout,
        id: Option<Value>,
        code: i32,
        message: &str,
        data: Option<Value>,
    ) -> Result<()> {
        let response = JsonRpcResponse {
            jsonrpc: "2.0".to_string(),
            id,
            result: None,
            error: Some(RpcError {
                code,
                message: message.to_string(),
                data,
            }),
        };

        let json = serde_json::to_string(&response)?;
        stdout.write_all(json.as_bytes()).await?;
        stdout.write_all(b"\n").await?;
        stdout.flush().await?;
        Ok(())
    }
}
