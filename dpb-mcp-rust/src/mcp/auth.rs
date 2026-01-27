use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::RwLock;
use std::time::{SystemTime, UNIX_EPOCH};
use sha2::{Sha256, Digest};

use super::errors::authentication_error;
use super::McpError;

/// Authentication configuration
#[derive(Debug, Clone)]
pub struct AuthConfig {
    pub enabled: bool,
    pub static_tokens: Vec<String>,
    pub token_env_var: String,
    pub public_methods: Vec<String>,
}

impl Default for AuthConfig {
    fn default() -> Self {
        Self {
            enabled: false,
            static_tokens: Vec::new(),
            token_env_var: "MCP_TOKEN".to_string(),
            public_methods: vec!["initialize".to_string(), "tools/list".to_string()],
        }
    }
}

/// Credentials representing authenticated context
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Credentials {
    #[serde(rename = "type")]
    pub cred_type: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub subject: Option<String>,
    #[serde(rename = "tokenHash", skip_serializing_if = "Option::is_none")]
    pub token_hash: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub context: Option<HashMap<String, serde_json::Value>>,
}

impl Credentials {
    pub fn anonymous() -> Self {
        Self {
            cred_type: "anonymous".to_string(),
            subject: None,
            token_hash: None,
            context: None,
        }
    }

    pub fn static_token(subject: &str, token_hash: &str) -> Self {
        Self {
            cred_type: "static_token".to_string(),
            subject: Some(subject.to_string()),
            token_hash: Some(token_hash.to_string()),
            context: None,
        }
    }
}

/// Request context with credentials and metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RequestContext {
    pub credentials: Credentials,
    #[serde(rename = "requestId")]
    pub request_id: String,
    pub timestamp: String,
    #[serde(rename = "clientInfo", skip_serializing_if = "Option::is_none")]
    pub client_info: Option<HashMap<String, String>>,
}

impl RequestContext {
    pub fn new(credentials: Credentials) -> Self {
        let timestamp = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_millis();
        
        Self {
            credentials,
            request_id: format!("req_{}_{:x}", timestamp, rand::random::<u32>()),
            timestamp: chrono::Utc::now().to_rfc3339(),
            client_info: None,
        }
    }
}

lazy_static::lazy_static! {
    static ref AUTH_CONFIG: RwLock<AuthConfig> = RwLock::new(AuthConfig::default());
}

/// Configure authentication
pub fn configure_auth(config: AuthConfig) {
    let mut cfg = AUTH_CONFIG.write().unwrap();
    *cfg = config;

    // Load token from environment
    if let Ok(token) = std::env::var(&cfg.token_env_var) {
        if !token.is_empty() && !cfg.static_tokens.contains(&token) {
            cfg.static_tokens.push(token);
        }
    }
}

/// Check if auth is enabled
pub fn is_auth_enabled() -> bool {
    AUTH_CONFIG.read().unwrap().enabled
}

/// Hash a token for logging
fn hash_token(token: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(token.as_bytes());
    let result = hasher.finalize();
    hex::encode(&result[..8])
}

/// Validate authentication
pub fn validate_auth(method: &str, headers: &HashMap<String, String>) -> Result<Credentials, McpError> {
    let config = AUTH_CONFIG.read().unwrap();

    // If auth disabled, return anonymous
    if !config.enabled {
        return Ok(Credentials::anonymous());
    }

    // Check public methods
    if config.public_methods.contains(&method.to_string()) {
        return Ok(Credentials::anonymous());
    }

    // Get authorization header
    let auth_header = headers.get("Authorization")
        .or_else(|| headers.get("authorization"));

    let auth_header = match auth_header {
        Some(h) => h,
        None => return Err(authentication_error("Authentication required: No authorization header")),
    };

    // Extract token
    let token = if auth_header.starts_with("Bearer ") {
        &auth_header[7..]
    } else {
        auth_header
    };

    // Validate token
    if config.static_tokens.contains(&token.to_string()) {
        return Ok(Credentials::static_token("mcp-client", &hash_token(token)));
    }

    Err(authentication_error("Invalid token"))
}

/// Get auth info for reporting
pub fn get_auth_info() -> HashMap<String, serde_json::Value> {
    let config = AUTH_CONFIG.read().unwrap();
    let mut info = HashMap::new();
    info.insert("enabled".to_string(), serde_json::json!(config.enabled));
    info.insert("methods".to_string(), serde_json::json!(["static_token"]));
    info
}

/// Generate a random token
pub fn generate_token() -> String {
    use rand::Rng;
    let mut rng = rand::thread_rng();
    let bytes: [u8; 24] = rng.gen();
    base64::encode(&bytes)
}
