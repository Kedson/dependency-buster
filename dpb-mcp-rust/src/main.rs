mod analyzer;
mod composer;
mod mcp;
mod types;

use anyhow::Result;
use std::collections::HashMap;

use analyzer::{dependency, namespace, psr4, security};
use mcp::{InputSchema, Property, Server, Tool};

#[tokio::main]
async fn main() -> Result<()> {
    let server = Server::new("php-dependency-analyzer", "2.0.0");

    register_tools(&server).await;

    server.run().await
}

/// Helper to create a tool with standard repo_path input
fn repo_path_tool(name: &str, description: &str) -> Tool {
    Tool {
        name: name.to_string(),
        description: description.to_string(),
        input_schema: InputSchema {
            schema_type: "object".to_string(),
            properties: HashMap::from([
                ("repo_path".to_string(), Property {
                    property_type: "string".to_string(),
                    description: "Absolute path to PHP repository".to_string(),
                }),
            ]),
            required: vec!["repo_path".to_string()],
        },
        annotations: None, // Auto-filled by register_tool
    }
}

async fn register_tools(server: &Server) {
    // Tool 1: Analyze Dependencies
    server
        .register_tool(
            repo_path_tool(
                "analyze_dependencies",
                "Comprehensive dependency analysis including production, dev, and dependency tree"
            ),
            |args| {
                let repo_path = args.get("repo_path")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("repo_path required"))?;
                dependency::analyze_dependencies(repo_path)
            },
        )
        .await;

    // Tool 2: Analyze PSR-4
    server
        .register_tool(
            repo_path_tool(
                "analyze_psr4",
                "Analyze PSR-4 autoloading configuration and validate namespace compliance"
            ),
            |args| {
                let repo_path = args.get("repo_path")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("repo_path required"))?;
                psr4::analyze_psr4_autoloading(repo_path)
            },
        )
        .await;

    // Tool 3: Detect Namespaces
    server
        .register_tool(
            repo_path_tool(
                "detect_namespaces",
                "Detect all namespaces used in the codebase"
            ),
            |args| {
                let repo_path = args.get("repo_path")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("repo_path required"))?;
                namespace::detect_namespaces(repo_path)
            },
        )
        .await;

    // Tool 4: Analyze Namespace Usage
    server
        .register_tool(
            Tool {
                name: "analyze_namespace_usage".to_string(),
                description: "Analyze usage of a specific namespace across the codebase".to_string(),
                input_schema: InputSchema {
                    schema_type: "object".to_string(),
                    properties: HashMap::from([
                        ("repo_path".to_string(), Property {
                            property_type: "string".to_string(),
                            description: "Absolute path to PHP repository".to_string(),
                        }),
                        ("namespace".to_string(), Property {
                            property_type: "string".to_string(),
                            description: "Target namespace to analyze".to_string(),
                        }),
                    ]),
                    required: vec!["repo_path".to_string(), "namespace".to_string()],
                },
                annotations: None,
            },
            |args| {
                let repo_path = args.get("repo_path")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("repo_path required"))?;
                let namespace = args.get("namespace")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("namespace required"))?;
                namespace::analyze_namespace_usage(repo_path, namespace)
            },
        )
        .await;

    // Tool 5: Generate Dependency Graph
    server
        .register_tool(
            Tool {
                name: "generate_dependency_graph".to_string(),
                description: "Generate Mermaid diagram of dependency relationships".to_string(),
                input_schema: InputSchema {
                    schema_type: "object".to_string(),
                    properties: HashMap::from([
                        ("repo_path".to_string(), Property {
                            property_type: "string".to_string(),
                            description: "Absolute path to PHP repository".to_string(),
                        }),
                        ("max_depth".to_string(), Property {
                            property_type: "number".to_string(),
                            description: "Maximum depth for dependency tree (default: 2)".to_string(),
                        }),
                        ("include_dev".to_string(), Property {
                            property_type: "boolean".to_string(),
                            description: "Include development dependencies".to_string(),
                        }),
                        ("focus_package".to_string(), Property {
                            property_type: "string".to_string(),
                            description: "Focus on specific package and its dependencies".to_string(),
                        }),
                    ]),
                    required: vec!["repo_path".to_string()],
                },
                annotations: None,
            },
            |args| {
                let repo_path = args.get("repo_path")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("repo_path required"))?;
                let max_depth = args.get("max_depth")
                    .and_then(|v| v.as_f64())
                    .map(|v| v as usize)
                    .unwrap_or(2);
                let include_dev = args.get("include_dev")
                    .and_then(|v| v.as_bool())
                    .unwrap_or(false);
                let focus_package = args.get("focus_package")
                    .and_then(|v| v.as_str())
                    .map(|s| s.to_string());
                analyzer::generate_dependency_graph(repo_path, max_depth, include_dev, focus_package)
            },
        )
        .await;

    // Tool 6: Audit Security
    server
        .register_tool(
            repo_path_tool(
                "audit_security",
                "Audit dependencies for security vulnerabilities and outdated packages"
            ),
            |args| {
                let repo_path = args.get("repo_path")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("repo_path required"))?;
                security::audit_security(repo_path)
            },
        )
        .await;

    // Tool 7: Analyze Licenses
    server
        .register_tool(
            repo_path_tool(
                "analyze_licenses",
                "Analyze license distribution and compatibility across dependencies"
            ),
            |args| {
                let repo_path = args.get("repo_path")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("repo_path required"))?;
                security::analyze_licenses(repo_path)
            },
        )
        .await;

    // Tool 8: Find Circular Dependencies
    server
        .register_tool(
            repo_path_tool(
                "find_circular_dependencies",
                "Find circular dependency chains in the package graph"
            ),
            |args| {
                let repo_path = args.get("repo_path")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("repo_path required"))?;
                dependency::find_circular_dependencies(repo_path)
            },
        )
        .await;

    // Tool 9: Analyze Multi Repo
    server
        .register_tool(
            Tool {
                name: "analyze_multi_repo".to_string(),
                description: "Analyze dependencies across multiple repositories (Dependency Buster platform)".to_string(),
                input_schema: InputSchema {
                    schema_type: "object".to_string(),
                    properties: HashMap::from([
                        ("config_path".to_string(), Property {
                            property_type: "string".to_string(),
                            description: "Path to repository configuration JSON file".to_string(),
                        }),
                    ]),
                    required: vec!["config_path".to_string()],
                },
                annotations: None,
            },
            |args| {
                let config_path = args.get("config_path")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("config_path required"))?;
                analyzer::analyze_multiple_repositories(config_path)
            },
        )
        .await;

    // Tool 10: Generate Comprehensive Docs
    server
        .register_tool(
            Tool {
                name: "generate_comprehensive_docs".to_string(),
                description: "Generate comprehensive markdown documentation for a repository".to_string(),
                input_schema: InputSchema {
                    schema_type: "object".to_string(),
                    properties: HashMap::from([
                        ("repo_path".to_string(), Property {
                            property_type: "string".to_string(),
                            description: "Absolute path to PHP repository".to_string(),
                        }),
                        ("output_path".to_string(), Property {
                            property_type: "string".to_string(),
                            description: "Where to save the documentation file".to_string(),
                        }),
                    ]),
                    required: vec!["repo_path".to_string()],
                },
                annotations: None,
            },
            |args| {
                let repo_path = args.get("repo_path")
                    .and_then(|v| v.as_str())
                    .ok_or_else(|| anyhow::anyhow!("repo_path required"))?;
                let output_path = args.get("output_path")
                    .and_then(|v| v.as_str());
                analyzer::generate_comprehensive_docs(repo_path, output_path)
            },
        )
        .await;
}
