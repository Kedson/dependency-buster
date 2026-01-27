package mcp

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ActionSchema defines input/output validation for an action
type ActionSchema struct {
	Input  interface{} `json:"input"`
	Output interface{} `json:"output"`
}

// ActionDefinition defines a dynamic action
type ActionDefinition struct {
	Name        string                                                      `json:"name"`
	Title       string                                                      `json:"title"`
	Description string                                                      `json:"description"`
	Schema      ActionSchema                                                `json:"schema"`
	Annotations *ToolAnnotations                                            `json:"annotations,omitempty"`
	Handler     func(input map[string]interface{}, ctx *RequestContext) (interface{}, error) `json:"-"`
	PluginID    string                                                      `json:"pluginId,omitempty"`
}

// RegisteredAction is an action with registration metadata
type RegisteredAction struct {
	ActionDefinition
	ID           string    `json:"id"`
	RegisteredAt time.Time `json:"registeredAt"`
}

// ActionRegistry provides dynamic action registration
type ActionRegistry struct {
	actions map[string]*RegisteredAction
	counter int
	mu      sync.RWMutex
}

// NewActionRegistry creates a new registry
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string]*RegisteredAction),
	}
}

// Register adds a new action to the registry
func (r *ActionRegistry) Register(def ActionDefinition) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[def.Name]; exists {
		return "", fmt.Errorf("action %q is already registered", def.Name)
	}

	r.counter++
	id := fmt.Sprintf("action_%d", r.counter)

	registered := &RegisteredAction{
		ActionDefinition: def,
		ID:               id,
		RegisteredAt:     time.Now(),
	}

	// Apply default annotations if not provided
	if registered.Annotations == nil {
		ann := GetToolAnnotation(def.Name)
		registered.Annotations = &ann
	}

	r.actions[def.Name] = registered
	fmt.Printf("[Registry] Registered action: %s (%s)\n", def.Name, id)

	return id, nil
}

// Unregister removes an action from the registry
func (r *ActionRegistry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[name]; exists {
		delete(r.actions, name)
		fmt.Printf("[Registry] Unregistered action: %s\n", name)
		return true
	}
	return false
}

// Get retrieves an action by name
func (r *ActionRegistry) Get(name string) (*RegisteredAction, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	action, exists := r.actions[name]
	return action, exists
}

// List returns all registered actions
func (r *ActionRegistry) List(pluginID string) []*RegisteredAction {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*RegisteredAction
	for _, action := range r.actions {
		if pluginID == "" || action.PluginID == pluginID {
			result = append(result, action)
		}
	}
	return result
}

// Invoke executes an action by name
func (r *ActionRegistry) Invoke(name string, input map[string]interface{}, ctx *RequestContext) (interface{}, error) {
	action, exists := r.Get(name)
	if !exists {
		return nil, fmt.Errorf("action %q not found", name)
	}

	if action.Handler == nil {
		return nil, fmt.Errorf("action %q has no handler", name)
	}

	return action.Handler(input, ctx)
}

// ToMcpTools converts registry actions to MCP tool format
func (r *ActionRegistry) ToMcpTools() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tools []Tool
	for _, action := range r.actions {
		tool := Tool{
			Name:        action.Name,
			Description: action.Description,
			InputSchema: InputSchema{
				Type:       "object",
				Properties: make(map[string]Property),
				Required:   []string{},
			},
			Annotations: action.Annotations,
		}

		// Convert schema to InputSchema if it's a map
		if schemaMap, ok := action.Schema.Input.(map[string]interface{}); ok {
			if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
				for key, val := range props {
					if propMap, ok := val.(map[string]interface{}); ok {
						tool.InputSchema.Properties[key] = Property{
							Type:        getString(propMap, "type"),
							Description: getString(propMap, "description"),
						}
					}
				}
			}
			if required, ok := schemaMap["required"].([]interface{}); ok {
				for _, req := range required {
					if reqStr, ok := req.(string); ok {
						tool.InputSchema.Required = append(tool.InputSchema.Required, reqStr)
					}
				}
			}
		}

		tools = append(tools, tool)
	}
	return tools
}

// ToJSON returns JSON representation of the registry
func (r *ActionRegistry) ToJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return json.MarshalIndent(map[string]interface{}{
		"count":   len(r.actions),
		"actions": r.List(""),
	}, "", "  ")
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// Global registry instance
var Registry = NewActionRegistry()
