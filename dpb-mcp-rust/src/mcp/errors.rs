use serde::{Deserialize, Serialize};

/// Standard MCP error codes (JSON-RPC compatible)
pub mod error_codes {
    pub const PARSE_ERROR: i32 = -32700;
    pub const INVALID_REQUEST: i32 = -32600;
    pub const METHOD_NOT_FOUND: i32 = -32601;
    pub const INVALID_PARAMS: i32 = -32602;
    pub const INTERNAL_ERROR: i32 = -32603;

    // Custom MCP errors
    pub const NOT_FOUND: i32 = -32000;
    pub const NOT_ALLOWED: i32 = -32001;
    pub const VALIDATION: i32 = -32002;
    pub const AUTHENTICATION: i32 = -32003;
    pub const RATE_LIMITED: i32 = -32004;
    pub const TIMEOUT: i32 = -32005;
}

/// Typed MCP error
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct McpError {
    #[serde(rename = "type")]
    pub error_type: String,
    pub code: i32,
    pub message: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub data: Option<serde_json::Value>,
}

impl std::fmt::Display for McpError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.message)
    }
}

impl std::error::Error for McpError {}

impl McpError {
    /// Create a new MCP error
    pub fn new(error_type: &str, code: i32, message: &str) -> Self {
        Self {
            error_type: error_type.to_string(),
            code,
            message: message.to_string(),
            data: None,
        }
    }

    /// Create with additional data
    pub fn with_data(mut self, data: serde_json::Value) -> Self {
        self.data = Some(data);
        self
    }

    /// Convert to JSON string
    pub fn to_json(&self) -> String {
        serde_json::to_string(self).unwrap_or_else(|_| format!("{{\"error\":\"{}\"}}", self.message))
    }
}

/// Resource not found error
pub fn not_found_error(message: &str) -> McpError {
    McpError::new("NotFoundError", error_codes::NOT_FOUND, message)
}

/// Permission denied error
pub fn not_allowed_error(message: &str) -> McpError {
    McpError::new("NotAllowedError", error_codes::NOT_ALLOWED, message)
}

/// Validation error
pub fn validation_error(message: &str) -> McpError {
    McpError::new("ValidationError", error_codes::VALIDATION, message)
}

/// Authentication error
pub fn authentication_error(message: &str) -> McpError {
    McpError::new("AuthenticationError", error_codes::AUTHENTICATION, message)
}

/// Timeout error
pub fn timeout_error(message: &str) -> McpError {
    McpError::new("TimeoutError", error_codes::TIMEOUT, message)
}

/// Convert any error to MCP error
pub fn to_mcp_error(err: &dyn std::error::Error) -> McpError {
    let msg = err.to_string();
    let msg_lower = msg.to_lowercase();

    if msg_lower.contains("not found") || msg_lower.contains("no such file") {
        return not_found_error(&msg);
    }
    if msg_lower.contains("permission denied") || msg_lower.contains("access denied") {
        return not_allowed_error(&msg);
    }
    if msg_lower.contains("invalid") || msg_lower.contains("required") {
        return validation_error(&msg);
    }
    if msg_lower.contains("unauthorized") || msg_lower.contains("authentication") {
        return authentication_error(&msg);
    }
    if msg_lower.contains("timeout") {
        return timeout_error(&msg);
    }

    McpError::new("InternalError", error_codes::INTERNAL_ERROR, &msg)
}

/// Convert anyhow::Error to McpError
pub fn anyhow_to_mcp_error(err: &anyhow::Error) -> McpError {
    let msg = err.to_string();
    let msg_lower = msg.to_lowercase();

    if msg_lower.contains("not found") || msg_lower.contains("no such file") {
        return not_found_error(&msg);
    }
    if msg_lower.contains("permission denied") || msg_lower.contains("access denied") {
        return not_allowed_error(&msg);
    }
    if msg_lower.contains("invalid") || msg_lower.contains("required") {
        return validation_error(&msg);
    }

    McpError::new("InternalError", error_codes::INTERNAL_ERROR, &msg)
}
