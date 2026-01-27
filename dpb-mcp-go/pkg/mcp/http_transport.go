package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HttpConfig configures the HTTP/SSE transport
type HttpConfig struct {
	Port        int      `json:"port"`
	Host        string   `json:"host"`
	BasePath    string   `json:"basePath"`
	CorsOrigins []string `json:"corsOrigins"`
}

// SseClient represents a connected SSE client
type SseClient struct {
	ID       string
	Response http.ResponseWriter
	Context  RequestContext
	Flusher  http.Flusher
}

// JsonRpcRequest represents a JSON-RPC 2.0 request
type JsonRpcHttpRequest struct {
	Jsonrpc string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// JsonRpcResponse represents a JSON-RPC 2.0 response
type JsonRpcHttpResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// HttpTransport provides HTTP/SSE transport for MCP
type HttpTransport struct {
	config      HttpConfig
	server      *http.Server
	handler     func(method string, params map[string]interface{}, ctx *RequestContext) (interface{}, error)
	sseClients  map[string]*SseClient
	clientMu    sync.RWMutex
	clientCount int
}

// DefaultHttpConfig returns default configuration
func DefaultHttpConfig() HttpConfig {
	return HttpConfig{
		Port:        3000,
		Host:        "127.0.0.1",
		BasePath:    "/api/mcp",
		CorsOrigins: []string{"*"},
	}
}

// NewHttpTransport creates a new HTTP/SSE transport
func NewHttpTransport(
	handler func(method string, params map[string]interface{}, ctx *RequestContext) (interface{}, error),
	config *HttpConfig,
) *HttpTransport {
	cfg := DefaultHttpConfig()
	if config != nil {
		if config.Port > 0 {
			cfg.Port = config.Port
		}
		if config.Host != "" {
			cfg.Host = config.Host
		}
		if config.BasePath != "" {
			cfg.BasePath = config.BasePath
		}
		if len(config.CorsOrigins) > 0 {
			cfg.CorsOrigins = config.CorsOrigins
		}
	}

	return &HttpTransport{
		config:     cfg,
		handler:    handler,
		sseClients: make(map[string]*SseClient),
	}
}

// Start begins listening for HTTP connections
func (t *HttpTransport) Start() error {
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc(t.config.BasePath+"/v1", t.handleStreamableHttp)
	mux.HandleFunc(t.config.BasePath+"/v1/", t.handleStreamableHttp)
	mux.HandleFunc(t.config.BasePath+"/v1/sse", t.handleSse)
	mux.HandleFunc(t.config.BasePath+"/health", t.handleHealth)
	mux.HandleFunc(t.config.BasePath+"/info", t.handleInfo)

	addr := fmt.Sprintf("%s:%d", t.config.Host, t.config.Port)
	t.server = &http.Server{
		Addr:    addr,
		Handler: t.corsMiddleware(mux),
	}

	fmt.Printf("HTTP/SSE transport listening on http://%s%s/v1\n", addr, t.config.BasePath)
	fmt.Printf("SSE endpoint: http://%s%s/v1/sse\n", addr, t.config.BasePath)

	return t.server.ListenAndServe()
}

// Stop closes the HTTP server
func (t *HttpTransport) Stop() error {
	// Close all SSE clients
	t.clientMu.Lock()
	for id := range t.sseClients {
		delete(t.sseClients, id)
	}
	t.clientMu.Unlock()

	if t.server != nil {
		return t.server.Close()
	}
	return nil
}

// Broadcast sends an event to all SSE clients
func (t *HttpTransport) Broadcast(event string, data interface{}) {
	t.clientMu.RLock()
	defer t.clientMu.RUnlock()

	for _, client := range t.sseClients {
		t.sendSseEvent(client, event, data)
	}
}

// CORS middleware
func (t *HttpTransport) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		// Check if origin is allowed
		allowed := false
		for _, o := range t.config.CorsOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Handle Streamable HTTP (JSON-RPC over HTTP POST)
func (t *HttpTransport) handleStreamableHttp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		t.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.jsonError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var request JsonRpcHttpRequest
	if err := json.Unmarshal(body, &request); err != nil {
		t.jsonError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate authentication
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	credentials, _ := ValidateAuth(request.Method, headers)
	ctx := CreateRequestContext(credentials, nil)

	// Handle the request
	result, err := t.handler(request.Method, request.Params, &ctx)
	if err != nil {
		response := JsonRpcHttpResponse{
			Jsonrpc: "2.0",
			ID:      request.ID,
			Error: map[string]interface{}{
				"code":    -32603,
				"message": err.Error(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := JsonRpcHttpResponse{
		Jsonrpc: "2.0",
		ID:      request.ID,
		Result:  result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Handle SSE connections
func (t *HttpTransport) handleSse(w http.ResponseWriter, r *http.Request) {
	// Check if response supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		t.jsonError(w, http.StatusInternalServerError, "SSE not supported")
		return
	}

	// Validate authentication
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	credentials, err := ValidateAuth("sse/connect", headers)
	if err != nil && isAuthEnabled() {
		t.jsonError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	ctx := CreateRequestContext(credentials, nil)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Generate client ID
	t.clientMu.Lock()
	t.clientCount++
	clientID := fmt.Sprintf("client_%d_%d", t.clientCount, time.Now().Unix())
	client := &SseClient{
		ID:       clientID,
		Response: w,
		Context:  ctx,
		Flusher:  flusher,
	}
	t.sseClients[clientID] = client
	t.clientMu.Unlock()

	// Send connected event
	t.sendSseEvent(client, "connected", map[string]string{"clientId": clientID})

	// Keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Wait for client disconnect
	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			t.clientMu.Lock()
			delete(t.sseClients, clientID)
			t.clientMu.Unlock()
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": keep-alive\n\n")
			flusher.Flush()
		}
	}
}

func (t *HttpTransport) sendSseEvent(client *SseClient, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(client.Response, "event: %s\n", event)
	fmt.Fprintf(client.Response, "data: %s\n\n", jsonData)
	client.Flusher.Flush()
}

func (t *HttpTransport) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (t *HttpTransport) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":      "dpb-mcp",
		"version":   "1.0.0",
		"protocols": []string{"stdio", "http", "sse"},
		"endpoints": map[string]string{
			"http":   t.config.BasePath + "/v1",
			"sse":    t.config.BasePath + "/v1/sse",
			"health": t.config.BasePath + "/health",
		},
	})
}

func (t *HttpTransport) jsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Helper to check if auth is enabled
func isAuthEnabled() bool {
	return false // TODO: Implement proper auth check
}
