use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Tool annotations for AI clients
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct ToolAnnotations {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub title: Option<String>,
    #[serde(rename = "destructiveHint", skip_serializing_if = "Option::is_none")]
    pub destructive_hint: Option<bool>,
    #[serde(rename = "idempotentHint", skip_serializing_if = "Option::is_none")]
    pub idempotent_hint: Option<bool>,
    #[serde(rename = "readOnlyHint", skip_serializing_if = "Option::is_none")]
    pub read_only_hint: Option<bool>,
    #[serde(rename = "openWorldHint", skip_serializing_if = "Option::is_none")]
    pub open_world_hint: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tags: Option<Vec<String>>,
    #[serde(rename = "cacheTtlSeconds", skip_serializing_if = "Option::is_none")]
    pub cache_ttl_seconds: Option<i32>,
}

impl ToolAnnotations {
    /// Create analysis annotation
    pub fn analysis() -> Self {
        Self {
            title: None,
            read_only_hint: Some(true),
            idempotent_hint: Some(true),
            destructive_hint: Some(false),
            open_world_hint: Some(false),
            cache_ttl_seconds: Some(300),
            tags: Some(vec!["analysis".to_string(), "read-only".to_string()]),
        }
    }

    /// Create security annotation
    pub fn security() -> Self {
        Self {
            title: None,
            read_only_hint: Some(true),
            idempotent_hint: Some(true),
            destructive_hint: Some(false),
            open_world_hint: Some(true),
            cache_ttl_seconds: Some(60),
            tags: Some(vec!["security".to_string(), "audit".to_string()]),
        }
    }

    /// Create visualization annotation
    pub fn visualization() -> Self {
        Self {
            title: None,
            read_only_hint: Some(true),
            idempotent_hint: Some(true),
            destructive_hint: Some(false),
            open_world_hint: Some(false),
            cache_ttl_seconds: Some(300),
            tags: Some(vec!["visualization".to_string(), "graph".to_string()]),
        }
    }

    /// Create documentation annotation
    pub fn documentation() -> Self {
        Self {
            title: None,
            read_only_hint: Some(false),
            idempotent_hint: Some(true),
            destructive_hint: Some(false),
            open_world_hint: Some(false),
            cache_ttl_seconds: None,
            tags: Some(vec!["documentation".to_string(), "generate".to_string()]),
        }
    }

    /// Create multi-repo annotation
    pub fn multi_repo() -> Self {
        Self {
            title: None,
            read_only_hint: Some(true),
            idempotent_hint: Some(true),
            destructive_hint: Some(false),
            open_world_hint: Some(false),
            cache_ttl_seconds: Some(120),
            tags: Some(vec!["multi-repo".to_string(), "analysis".to_string()]),
        }
    }

    /// Set title
    pub fn with_title(mut self, title: &str) -> Self {
        self.title = Some(title.to_string());
        self
    }
}

/// Get annotation for a tool by name
pub fn get_tool_annotation(tool_name: &str) -> ToolAnnotations {
    match tool_name {
        "analyze_dependencies" => ToolAnnotations::analysis().with_title("Analyze Dependencies"),
        "analyze_psr4" => ToolAnnotations::analysis().with_title("Analyze PSR-4"),
        "detect_namespaces" => ToolAnnotations::analysis().with_title("Detect Namespaces"),
        "analyze_namespace_usage" => ToolAnnotations::analysis().with_title("Analyze Namespace Usage"),
        "audit_security" => ToolAnnotations::security().with_title("Audit Security"),
        "analyze_licenses" => ToolAnnotations::security().with_title("Analyze Licenses"),
        "generate_dependency_graph" => ToolAnnotations::visualization().with_title("Generate Dependency Graph"),
        "find_circular_dependencies" => ToolAnnotations::visualization().with_title("Find Circular Dependencies"),
        "analyze_multi_repo" => ToolAnnotations::multi_repo().with_title("Analyze Multi-Repo"),
        "generate_comprehensive_docs" => ToolAnnotations::documentation().with_title("Generate Documentation"),
        _ => ToolAnnotations::default(),
    }
}
