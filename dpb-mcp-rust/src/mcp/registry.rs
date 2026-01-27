//! Dynamic Action Registry for MCP
//! Allows runtime registration and discovery of actions

#![allow(dead_code)]

use anyhow::Result;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::collections::HashMap;
use std::sync::{Arc, RwLock};
use chrono::{DateTime, Utc};

use super::annotations::ToolAnnotations;
use super::auth::RequestContext;

/// Action schema for input/output validation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ActionSchema {
    pub input: Value,
    pub output: Value,
}

/// Action definition
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ActionDefinition {
    pub name: String,
    pub title: String,
    pub description: String,
    pub schema: ActionSchema,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub annotations: Option<ToolAnnotations>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub plugin_id: Option<String>,
}

/// Registered action with metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegisteredAction {
    #[serde(flatten)]
    pub definition: ActionDefinition,
    pub id: String,
    pub registered_at: DateTime<Utc>,
}

/// Action handler type
pub type ActionHandler = Arc<dyn Fn(Value, &RequestContext) -> Result<Value> + Send + Sync>;

/// Internal action with handler
struct InternalAction {
    registered: RegisteredAction,
    handler: ActionHandler,
}

/// Dynamic Action Registry
pub struct ActionRegistry {
    actions: RwLock<HashMap<String, InternalAction>>,
    counter: RwLock<u64>,
}

impl Default for ActionRegistry {
    fn default() -> Self {
        Self::new()
    }
}

impl ActionRegistry {
    /// Create a new registry
    pub fn new() -> Self {
        Self {
            actions: RwLock::new(HashMap::new()),
            counter: RwLock::new(0),
        }
    }

    /// Register a new action
    pub fn register<F>(&self, definition: ActionDefinition, handler: F) -> Result<String>
    where
        F: Fn(Value, &RequestContext) -> Result<Value> + Send + Sync + 'static,
    {
        let mut actions = self.actions.write().unwrap();
        
        if actions.contains_key(&definition.name) {
            return Err(anyhow::anyhow!("Action '{}' is already registered", definition.name));
        }

        let mut counter = self.counter.write().unwrap();
        *counter += 1;
        let id = format!("action_{}", *counter);

        let registered = RegisteredAction {
            definition: definition.clone(),
            id: id.clone(),
            registered_at: Utc::now(),
        };

        let internal = InternalAction {
            registered,
            handler: Arc::new(handler),
        };

        eprintln!("[Registry] Registered action: {} ({})", definition.name, id);
        actions.insert(definition.name, internal);

        Ok(id)
    }

    /// Unregister an action
    pub fn unregister(&self, name: &str) -> bool {
        let mut actions = self.actions.write().unwrap();
        if actions.remove(name).is_some() {
            eprintln!("[Registry] Unregistered action: {}", name);
            true
        } else {
            false
        }
    }

    /// Get action by name
    pub fn get(&self, name: &str) -> Option<RegisteredAction> {
        let actions = self.actions.read().unwrap();
        actions.get(name).map(|a| a.registered.clone())
    }

    /// List all registered actions
    pub fn list(&self, plugin_id: Option<&str>) -> Vec<RegisteredAction> {
        let actions = self.actions.read().unwrap();
        actions
            .values()
            .filter(|a| {
                plugin_id.map_or(true, |pid| {
                    a.registered.definition.plugin_id.as_deref() == Some(pid)
                })
            })
            .map(|a| a.registered.clone())
            .collect()
    }

    /// Invoke an action by name
    pub fn invoke(&self, name: &str, input: Value, ctx: &RequestContext) -> Result<Value> {
        let actions = self.actions.read().unwrap();
        
        let action = actions
            .get(name)
            .ok_or_else(|| anyhow::anyhow!("Action '{}' not found", name))?;

        (action.handler)(input, ctx)
    }

    /// Convert to MCP tools format
    pub fn to_mcp_tools(&self) -> Vec<Value> {
        let actions = self.actions.read().unwrap();
        
        actions
            .values()
            .map(|action| {
                let def = &action.registered.definition;
                serde_json::json!({
                    "name": def.name,
                    "description": def.description,
                    "inputSchema": def.schema.input,
                    "annotations": def.annotations,
                })
            })
            .collect()
    }

    /// Get count of registered actions
    pub fn count(&self) -> usize {
        self.actions.read().unwrap().len()
    }
}

lazy_static::lazy_static! {
    /// Global registry instance
    pub static ref REGISTRY: ActionRegistry = ActionRegistry::new();
}

/// Convenience function to register an action
pub fn register_action<F>(definition: ActionDefinition, handler: F) -> Result<String>
where
    F: Fn(Value, &RequestContext) -> Result<Value> + Send + Sync + 'static,
{
    REGISTRY.register(definition, handler)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_registry() {
        let registry = ActionRegistry::new();
        
        let def = ActionDefinition {
            name: "test_action".to_string(),
            title: "Test Action".to_string(),
            description: "A test action".to_string(),
            schema: ActionSchema {
                input: serde_json::json!({"type": "object"}),
                output: serde_json::json!({"type": "string"}),
            },
            annotations: None,
            plugin_id: None,
        };

        let id = registry.register(def, |_input, _ctx| {
            Ok(serde_json::json!("result"))
        }).unwrap();

        assert!(id.starts_with("action_"));
        assert_eq!(registry.count(), 1);
        
        let action = registry.get("test_action").unwrap();
        assert_eq!(action.definition.name, "test_action");
        
        assert!(registry.unregister("test_action"));
        assert_eq!(registry.count(), 0);
    }
}
