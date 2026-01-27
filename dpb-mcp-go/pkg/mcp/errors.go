package mcp

import (
	"encoding/json"
	"strings"
)

// Standard MCP error codes (JSON-RPC compatible)
const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603

	// Custom MCP errors
	ErrCodeNotFound       = -32000
	ErrCodeNotAllowed     = -32001
	ErrCodeValidation     = -32002
	ErrCodeAuthentication = -32003
	ErrCodeRateLimited    = -32004
	ErrCodeTimeout        = -32005
)

// McpError represents a typed MCP error
type McpError struct {
	Type    string      `json:"type"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *McpError) Error() string {
	return e.Message
}

func (e *McpError) ToJSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// NotFoundError represents a resource not found error
func NotFoundError(message string, data ...interface{}) *McpError {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	return &McpError{
		Type:    "NotFoundError",
		Code:    ErrCodeNotFound,
		Message: message,
		Data:    d,
	}
}

// NotAllowedError represents a permission denied error
func NotAllowedError(message string, data ...interface{}) *McpError {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	return &McpError{
		Type:    "NotAllowedError",
		Code:    ErrCodeNotAllowed,
		Message: message,
		Data:    d,
	}
}

// ValidationError represents an input validation error
func ValidationError(message string, data ...interface{}) *McpError {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	return &McpError{
		Type:    "ValidationError",
		Code:    ErrCodeValidation,
		Message: message,
		Data:    d,
	}
}

// AuthenticationError represents an authentication failure
func AuthenticationError(message string, data ...interface{}) *McpError {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	return &McpError{
		Type:    "AuthenticationError",
		Code:    ErrCodeAuthentication,
		Message: message,
		Data:    d,
	}
}

// TimeoutError represents a timeout error
func TimeoutError(message string, data ...interface{}) *McpError {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	return &McpError{
		Type:    "TimeoutError",
		Code:    ErrCodeTimeout,
		Message: message,
		Data:    d,
	}
}

// ToMcpError converts any error to an McpError
func ToMcpError(err error) *McpError {
	if mcpErr, ok := err.(*McpError); ok {
		return mcpErr
	}

	msg := err.Error()
	msgLower := strings.ToLower(msg)

	// Check for common error patterns
	if strings.Contains(msgLower, "not found") || strings.Contains(msgLower, "no such file") {
		return NotFoundError(msg)
	}
	if strings.Contains(msgLower, "permission denied") || strings.Contains(msgLower, "access denied") {
		return NotAllowedError(msg)
	}
	if strings.Contains(msgLower, "invalid") || strings.Contains(msgLower, "required") {
		return ValidationError(msg)
	}
	if strings.Contains(msgLower, "unauthorized") || strings.Contains(msgLower, "authentication") {
		return AuthenticationError(msg)
	}
	if strings.Contains(msgLower, "timeout") {
		return TimeoutError(msg)
	}

	return &McpError{
		Type:    "InternalError",
		Code:    ErrCodeInternalError,
		Message: msg,
	}
}
