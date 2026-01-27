package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Server represents an MCP server with enterprise features
type Server struct {
	name         string
	version      string
	capabilities Capabilities
	tools        []Tool
	handlers     map[string]ToolHandler
	httpServer   *http.Server
}

// Capabilities represents server capabilities
type Capabilities struct {
	Tools bool `json:"tools"`
}

// Tool represents an MCP tool definition with annotations
type Tool struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	InputSchema InputSchema      `json:"inputSchema"`
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

// InputSchema represents tool input schema
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property represents a schema property
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ToolHandler is a function that handles tool execution
type ToolHandler func(args map[string]interface{}) (interface{}, error)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolContent represents tool response content
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewServer creates a new MCP server with enterprise features
func NewServer(name, version string) *Server {
	// Configure auth from environment
	ConfigureAuth(AuthConfig{
		Enabled:       os.Getenv("MCP_AUTH_ENABLED") == "true",
		TokenEnvVar:   "MCP_TOKEN",
		PublicMethods: []string{"initialize", "tools/list"},
	})

	return &Server{
		name:    name,
		version: version,
		capabilities: Capabilities{
			Tools: true,
		},
		tools:    make([]Tool, 0),
		handlers: make(map[string]ToolHandler),
	}
}

// RegisterTool registers a new tool with annotations
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	// Auto-add annotations if not provided
	if tool.Annotations == nil {
		ann := GetToolAnnotation(tool.Name)
		tool.Annotations = &ann
	}
	s.tools = append(s.tools, tool)
	s.handlers[tool.Name] = handler
}

// Run starts the MCP server (stdio or HTTP based on environment)
func (s *Server) Run() error {
	transportMode := os.Getenv("MCP_TRANSPORT")
	if transportMode == "" {
		transportMode = "stdio"
	}

	log.SetOutput(os.Stderr)
	log.Printf("PHP Dependency Analyzer MCP Server v%s", s.version)
	log.Printf("Transport: %s", transportMode)
	log.Printf("Auth: %v", IsAuthEnabled())
	log.Println("Features: Tool Annotations, Typed Errors, Credentials Context")

	if transportMode == "http" {
		return s.runHTTP()
	}
	return s.runStdio()
}

// runStdio runs the server in stdio mode
func (s *Server) runStdio() error {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(writer, req.ID, -32700, "Parse error", nil)
			continue
		}

		// Create request context
		ctx := CreateRequestContext(Credentials{Type: "anonymous"}, nil)
		s.handleRequest(writer, &req, ctx)
	}

	return nil
}

// runHTTP runs the server in HTTP mode with SSE support
func (s *Server) runHTTP() error {
	port := os.Getenv("MCP_HTTP_PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	// JSON-RPC endpoint
	mux.HandleFunc("/api/mcp/v1", s.handleHTTPRequest)
	mux.HandleFunc("/api/mcp/v1/", s.handleHTTPRequest)

	// SSE endpoint
	mux.HandleFunc("/api/mcp/v1/sse", s.handleSSE)

	// Health check
	mux.HandleFunc("/api/mcp/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Info endpoint
	mux.HandleFunc("/api/mcp/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":       s.name,
			"version":    s.version,
			"protocols":  []string{"stdio", "http", "sse"},
			"auth":       GetAuthInfo(),
		})
	})

	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("HTTP/SSE transport listening on http://127.0.0.1:%s/api/mcp/v1", port)
	return s.httpServer.ListenAndServe()
}

// handleHTTPRequest handles HTTP JSON-RPC requests
func (s *Server) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendHTTPError(w, nil, ErrCodeParseError, "Parse error")
		return
	}

	// Validate authentication
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	creds, err := ValidateAuth(req.Method, headers)
	if err != nil {
		s.sendHTTPError(w, req.ID, ErrCodeAuthentication, err.Error())
		return
	}

	ctx := CreateRequestContext(creds, nil)
	result := s.executeRequest(&req, ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleSSE handles SSE connections
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Validate auth
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	_, err := ValidateAuth("sse/connect", headers)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Send connected event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	// Keep alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": keep-alive\n\n")
			flusher.Flush()
		}
	}
}

// executeRequest executes a JSON-RPC request and returns the response
func (s *Server) executeRequest(req *JSONRPCRequest, ctx RequestContext) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":    s.capabilities,
				"serverInfo": map[string]string{
					"name":    s.name,
					"version": s.version,
				},
				"features": map[string]interface{}{
					"authentication": GetAuthInfo(),
					"transports":     []string{"stdio", "http", "sse"},
				},
			},
		}
	case "tools/list":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]interface{}{"tools": s.tools},
		}
	case "tools/call":
		return s.executeToolCall(req, ctx)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: ErrCodeMethodNotFound, Message: "Method not found"},
		}
	}
}

// executeToolCall executes a tool and returns structured response
func (s *Server) executeToolCall(req *JSONRPCRequest, _ RequestContext) *JSONRPCResponse {
	name, ok := req.Params["name"].(string)
	if !ok {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: ErrCodeInvalidParams, Message: "Invalid params: name required"},
		}
	}

	args, ok := req.Params["arguments"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	handler, exists := s.handlers[name]
	if !exists {
		mcpErr := NotFoundError(fmt.Sprintf("Tool \"%s\" not found", name))
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []ToolContent{{Type: "text", Text: mcpErr.ToJSON()}},
				"isError": true,
			},
		}
	}

	result, err := handler(args)
	if err != nil {
		mcpErr := ToMcpError(err)
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []ToolContent{{Type: "text", Text: mcpErr.ToJSON()}},
				"isError": true,
			},
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []ToolContent{{Type: "text", Text: fmt.Sprintf("%v", result)}},
		},
	}
}

// sendHTTPError sends an error response via HTTP
func (s *Server) sendHTTPError(w http.ResponseWriter, id interface{}, code int, message string) {
	statusCode := http.StatusInternalServerError
	if code == ErrCodeAuthentication {
		statusCode = http.StatusUnauthorized
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	})
}

// handleRequest processes a JSON-RPC request with context
func (s *Server) handleRequest(writer *bufio.Writer, req *JSONRPCRequest, ctx RequestContext) {
	resp := s.executeRequest(req, ctx)
	s.writeResponse(writer, resp)
}

// sendError sends a JSON-RPC error response
func (s *Server) sendError(writer *bufio.Writer, id interface{}, code int, message string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.writeResponse(writer, &resp)
}

// writeResponse writes a JSON-RPC response to stdout
func (s *Server) writeResponse(writer *bufio.Writer, resp *JSONRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v\n", err)
		return
	}

	writer.Write(data)
	writer.WriteByte('\n')
	writer.Flush()
}
