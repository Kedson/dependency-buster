package mcp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strings"
	"time"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled       bool
	StaticTokens  []string
	TokenEnvVar   string
	PublicMethods []string
}

// Credentials represents authenticated user context
type Credentials struct {
	Type      string                 `json:"type"`
	Subject   string                 `json:"subject,omitempty"`
	TokenHash string                 `json:"tokenHash,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// RequestContext holds request metadata and credentials
type RequestContext struct {
	Credentials Credentials            `json:"credentials"`
	RequestID   string                 `json:"requestId"`
	Timestamp   time.Time              `json:"timestamp"`
	ClientInfo  map[string]interface{} `json:"clientInfo,omitempty"`
}

// Global auth configuration
var authConfig = AuthConfig{
	Enabled:       false,
	TokenEnvVar:   "MCP_TOKEN",
	PublicMethods: []string{"initialize", "tools/list"},
}

// ConfigureAuth sets up authentication
func ConfigureAuth(config AuthConfig) {
	authConfig = config

	// Load token from environment
	if authConfig.TokenEnvVar != "" {
		if token := os.Getenv(authConfig.TokenEnvVar); token != "" {
			authConfig.StaticTokens = append(authConfig.StaticTokens, token)
		}
	}
}

// IsAuthEnabled returns whether authentication is enabled
func IsAuthEnabled() bool {
	return authConfig.Enabled
}

// hashToken creates a hash of a token for logging
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])[:16]
}

// ValidateAuth validates authentication for a request
func ValidateAuth(method string, headers map[string]string) (Credentials, error) {
	// If auth disabled, return anonymous
	if !authConfig.Enabled {
		return Credentials{Type: "anonymous"}, nil
	}

	// Check public methods
	for _, pm := range authConfig.PublicMethods {
		if method == pm {
			return Credentials{Type: "anonymous"}, nil
		}
	}

	// Get authorization header
	authHeader := headers["Authorization"]
	if authHeader == "" {
		authHeader = headers["authorization"]
	}
	if authHeader == "" {
		return Credentials{}, AuthenticationError("Authentication required: No authorization header")
	}

	// Extract token
	var token string
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		token = authHeader
	}

	// Validate token
	for _, validToken := range authConfig.StaticTokens {
		if token == validToken {
			return Credentials{
				Type:      "static_token",
				Subject:   "mcp-client",
				TokenHash: hashToken(token),
			}, nil
		}
	}

	return Credentials{}, AuthenticationError("Invalid token")
}

// CreateRequestContext creates a new request context
func CreateRequestContext(creds Credentials, clientInfo map[string]interface{}) RequestContext {
	return RequestContext{
		Credentials: creds,
		RequestID:   generateRequestID(),
		Timestamp:   time.Now(),
		ClientInfo:  clientInfo,
	}
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	timestamp := time.Now().UnixMilli()
	random := make([]byte, 4)
	rand.Read(random)
	return "req_" + base64.RawURLEncoding.EncodeToString([]byte{
		byte(timestamp >> 24), byte(timestamp >> 16),
		byte(timestamp >> 8), byte(timestamp),
	}) + "_" + hex.EncodeToString(random)
}

// GenerateToken creates a secure random token
func GenerateToken() string {
	token := make([]byte, 24)
	rand.Read(token)
	return base64.StdEncoding.EncodeToString(token)
}

// GetAuthInfo returns authentication configuration info
func GetAuthInfo() map[string]interface{} {
	return map[string]interface{}{
		"enabled": authConfig.Enabled,
		"methods": []string{"static_token"},
	}
}
