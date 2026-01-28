//! MkDocs Documentation Generator
//! Generates MkDocs-compatible documentation structure with multi-file layout

use anyhow::Result;
use chrono::Utc;
use std::fs;

use crate::composer::read_composer_json;
use super::dependency::analyze_dependencies;
use super::psr4::analyze_psr4_autoloading;
use super::namespace::detect_namespaces;
use super::security::{audit_security, analyze_licenses};
use super::generate_dependency_graph;
use super::tracker::{create_dependency_snapshot, load_tracker, compare_snapshots};

pub struct MkDocsOptions {
    pub repo_path: String,
    pub output_dir: Option<String>,
    pub include_changelog: bool,
    pub format: String, // "mkdocs", "html", "markdown"
    pub site_name: Option<String>,
    pub site_description: Option<String>,
}

/// Generate MkDocs-compatible documentation structure
pub fn generate_mkdocs_docs(options: MkDocsOptions) -> Result<String> {
    let output_dir = options.output_dir
        .unwrap_or_else(|| format!("{}/docs", options.repo_path));
    
    let format = if options.format.is_empty() {
        "mkdocs".to_string()
    } else {
        options.format
    };

    // Ensure output directory exists
    fs::create_dir_all(&output_dir)?;

    // Gather all analysis data
    let composer = read_composer_json(&options.repo_path)?;
    let deps_json = analyze_dependencies(&options.repo_path)?;
    let psr4_json = analyze_psr4_autoloading(&options.repo_path)?;
    let namespaces_json = detect_namespaces(&options.repo_path)?;
    let security_json = audit_security(&options.repo_path)?;
    let licenses_json = analyze_licenses(&options.repo_path)?;
    let dep_graph = generate_dependency_graph(&options.repo_path, 2, false, None)?;
    
    // Parse JSON strings (simplified - in production would properly handle errors)
    let deps: serde_json::Value = serde_json::from_str(&deps_json).unwrap_or(serde_json::json!({}));
    let psr4: serde_json::Value = serde_json::from_str(&psr4_json).unwrap_or(serde_json::json!({}));
    let namespaces: serde_json::Value = serde_json::from_str(&namespaces_json).unwrap_or(serde_json::json!({}));
    let security: serde_json::Value = serde_json::from_str(&security_json).unwrap_or(serde_json::json!({}));
    let licenses: serde_json::Value = serde_json::from_str(&licenses_json).unwrap_or(serde_json::json!({}));

    // Get project info
    let project_name = options.site_name.unwrap_or_else(|| {
        composer.name.clone()
            .unwrap_or_else(|| std::path::PathBuf::from(&options.repo_path)
                .file_name()
                .and_then(|n| n.to_str())
                .unwrap_or("Project")
                .to_string())
    });

    let project_desc = options.site_description.unwrap_or_else(|| {
        composer.description.clone().unwrap_or_else(|| "Dependency Analysis Documentation".to_string())
    });

    // Generate changelog if requested
    let changelog_content = if options.include_changelog {
        generate_changelog(&options.repo_path).unwrap_or_default()
    } else {
        String::new()
    };

    // Generate individual markdown files
    let index_content = generate_index(&project_name, &project_desc, &composer, &deps, options.include_changelog);
    let dependencies_content = generate_dependencies_doc(&deps, &dep_graph);
    let security_content = generate_security_doc(&security);
    let licenses_content = generate_licenses_doc(&licenses);
    let architecture_content = generate_architecture_doc(&psr4, &namespaces);

    // Write markdown files
    fs::write(format!("{}/index.md", output_dir), index_content)?;
    fs::write(format!("{}/dependencies.md", output_dir), dependencies_content)?;
    fs::write(format!("{}/security.md", output_dir), security_content)?;
    fs::write(format!("{}/licenses.md", output_dir), licenses_content)?;
    fs::write(format!("{}/architecture.md", output_dir), architecture_content)?;
    
    if !changelog_content.is_empty() {
        fs::write(format!("{}/changelog.md", output_dir), changelog_content)?;
    }

    // Generate mkdocs.yml if format is mkdocs
    if format == "mkdocs" {
        let mkdocs_config = generate_mkdocs_config(&project_name, &project_desc, options.include_changelog);
        fs::write(format!("{}/mkdocs.yml", output_dir), mkdocs_config)?;
    }

    Ok(format!("Documentation generated successfully in {}", output_dir))
}

fn generate_index(
    project_name: &str,
    description: &str,
    composer: &crate::types::ComposerJson,
    _deps: &serde_json::Value,
    include_changelog: bool,
) -> String {
    let now = Utc::now().to_rfc3339();
    let project_type = composer.package_type.as_deref().unwrap_or("library");
    let licenses = crate::composer::get_licenses(composer);
    let license_str = if licenses.is_empty() {
        "Not specified".to_string()
    } else {
        licenses.join(", ")
    };

    let mut content = format!("# {}\n\n", project_name);
    content.push_str(&format!("{}\n\n", description));
    content.push_str(&format!("**Generated:** {}\n\n", now));
    content.push_str("## Quick Overview\n\n");
    content.push_str(&format!("- **Project Type:** {}\n", project_type));
    content.push_str(&format!("- **License:** {}\n", license_str));
    content.push_str("- **Production Dependencies:** See dependencies.md\n");
    content.push_str("- **Development Dependencies:** See dependencies.md\n\n");
    content.push_str("## Documentation Sections\n\n");
    content.push_str("- [Dependencies](./dependencies.md) - Complete dependency analysis and tree\n");
    content.push_str("- [Security](./security.md) - Security audit and vulnerability report\n");
    content.push_str("- [Licenses](./licenses.md) - License compliance and distribution\n");
    content.push_str("- [Architecture](./architecture.md) - Namespace structure and PSR-4 compliance\n");
    if include_changelog {
        content.push_str("- [Changelog](./changelog.md) - Dependency change history\n");
    }
    content.push_str("\n## Getting Started\n\n");
    content.push_str("This documentation was automatically generated by dependency-buster MCP.\n\n");
    content.push_str("To view with MkDocs:\n");
    content.push_str("```bash\n");
    content.push_str("cd docs\n");
    content.push_str("mkdocs serve\n");
    content.push_str("```\n");

    content
}

fn generate_dependencies_doc(_deps: &serde_json::Value, graph: &str) -> String {
    let mut content = String::from("# Dependencies\n\n");
    content.push_str("## Summary\n\n");
    content.push_str("See full dependency analysis below.\n\n");
    content.push_str("## Dependency Graph\n\n");
    content.push_str("```mermaid\n");
    content.push_str(graph);
    content.push_str("\n```\n\n");
    content.push_str("*For detailed dependency information, use the `analyze_dependencies` tool.*\n");
    content
}

fn generate_security_doc(_security: &serde_json::Value) -> String {
    String::from("# Security Audit\n\n*For detailed security information, use the `audit_security` tool.*\n")
}

fn generate_licenses_doc(_licenses: &serde_json::Value) -> String {
    String::from("# License Compliance\n\n*For detailed license information, use the `analyze_licenses` tool.*\n")
}

fn generate_architecture_doc(_psr4: &serde_json::Value, _namespaces: &serde_json::Value) -> String {
    let mut content = String::from("# Architecture\n\n");
    content.push_str("## PSR-4 Autoloading\n\n");
    content.push_str("*For detailed PSR-4 information, use the `analyze_psr4` tool.*\n\n");
    content.push_str("## Namespaces\n\n");
    content.push_str("*For detailed namespace information, use the `detect_namespaces` tool.*\n");
    content
}

fn generate_changelog(repo_path: &str) -> Result<String> {
    let current_snapshot = create_dependency_snapshot(repo_path)?;
    let now = Utc::now().format("%Y-%m-%d").to_string();

    let old_snapshot = load_tracker(repo_path).ok();
    
    if old_snapshot.is_none() {
        return Ok(format!(
            "# Dependency Changelog\n\n## {}\n\nInitial snapshot created.\n\n**Total Dependencies:** {}\n",
            now,
            current_snapshot.metadata.total_count
        ));
    }

    let changes = compare_snapshots(&old_snapshot.unwrap(), &current_snapshot);
    
    if changes.is_empty() {
        return Ok(format!(
            "# Dependency Changelog\n\n## {}\n\nNo changes detected since last snapshot.\n\n**Total Dependencies:** {}\n",
            now,
            current_snapshot.metadata.total_count
        ));
    }

    let added: Vec<_> = changes.iter().filter(|c| c.change_type == "added").collect();
    let updated: Vec<_> = changes.iter().filter(|c| c.change_type == "updated").collect();
    let removed: Vec<_> = changes.iter().filter(|c| c.change_type == "removed").collect();

    let mut content = format!("# Dependency Changelog\n\n## {}\n\n### Summary\n\n", now);
    content.push_str(&format!("- **Added:** {}\n", added.len()));
    content.push_str(&format!("- **Updated:** {}\n", updated.len()));
    content.push_str(&format!("- **Removed:** {}\n\n", removed.len()));

    if !added.is_empty() {
        content.push_str("### Added\n\n");
        for change in added {
            content.push_str(&format!("- `{}` `{}`\n", change.name, change.new_version.as_ref().unwrap_or(&"".to_string())));
        }
        content.push_str("\n");
    }

    if !updated.is_empty() {
        content.push_str("### Updated\n\n");
        for change in updated {
            content.push_str(&format!(
                "- `{}`: `{}` → `{}`\n",
                change.name,
                change.old_version.as_ref().unwrap_or(&"".to_string()),
                change.new_version.as_ref().unwrap_or(&"".to_string())
            ));
        }
        content.push_str("\n");
    }

    if !removed.is_empty() {
        content.push_str("### Removed\n\n");
        for change in removed {
            content.push_str(&format!("- `{}` `{}`\n", change.name, change.old_version.as_ref().unwrap_or(&"".to_string())));
        }
        content.push_str("\n");
    }

    Ok(content)
}

fn generate_mkdocs_config(site_name: &str, site_description: &str, include_changelog: bool) -> String {
    let mut config = format!("site_name: {}\n", site_name);
    config.push_str(&format!("site_description: {}\n", site_description));
    config.push_str("site_url: https://example.com\n\n");
    config.push_str("theme:\n");
    config.push_str("  name: material\n");
    config.push_str("  palette:\n");
    config.push_str("    primary: blue\n");
    config.push_str("    accent: blue\n\n");
    config.push_str("markdown_extensions:\n");
    config.push_str("  - pymdownx.highlight:\n");
    config.push_str("      anchor_linenums: true\n");
    config.push_str("  - pymdownx.inlinehilite\n");
    config.push_str("  - pymdownx.snippets\n");
    config.push_str("  - pymdownx.superfences:\n");
    config.push_str("      custom_fences:\n");
    config.push_str("        - name: mermaid\n");
    config.push_str("          class: mermaid\n");
    config.push_str("          format: !!python/name:pymdownx.superfences.fence_code_format\n\n");
    config.push_str("nav:\n");
    config.push_str("  - Home: index.md\n");
    config.push_str("  - Dependencies: dependencies.md\n");
    config.push_str("  - Security: security.md\n");
    config.push_str("  - Licenses: licenses.md\n");
    config.push_str("  - Architecture: architecture.md\n");
    if include_changelog {
        config.push_str("  - Changelog: changelog.md\n");
    }

    config
}
