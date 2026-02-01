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

    // Generate HTML if format is html (before writing markdown files, so we can reuse the strings)
    if format == "html" {
        let html_content = generate_html_site(
            &project_name,
            &project_desc,
            &index_content,
            &dependencies_content,
            &security_content,
            &licenses_content,
            &architecture_content,
            &changelog_content,
        );
        fs::write(format!("{}/index.html", output_dir), html_content)?;
    }

    // Write markdown files
    fs::write(format!("{}/index.md", output_dir), &index_content)?;
    fs::write(format!("{}/dependencies.md", output_dir), &dependencies_content)?;
    fs::write(format!("{}/security.md", output_dir), &security_content)?;
    fs::write(format!("{}/licenses.md", output_dir), &licenses_content)?;
    fs::write(format!("{}/architecture.md", output_dir), &architecture_content)?;
    
    if !changelog_content.is_empty() {
        fs::write(format!("{}/changelog.md", output_dir), &changelog_content)?;
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
    deps: &serde_json::Value,
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

    // Extract dependency counts from JSON
    let prod_count = deps.get("stats")
        .and_then(|s| s.get("totalProduction"))
        .and_then(|v| v.as_u64())
        .unwrap_or(0);
    let dev_count = deps.get("stats")
        .and_then(|s| s.get("totalDevelopment"))
        .and_then(|v| v.as_u64())
        .unwrap_or(0);

    let mut content = format!("# {}\n\n", project_name);
    content.push_str(&format!("{}\n\n", description));
    content.push_str(&format!("**Generated:** {}\n\n", now));
    content.push_str("## Quick Overview\n\n");
    content.push_str(&format!("- **Project Type:** {}\n", project_type));
    content.push_str(&format!("- **License:** {}\n", license_str));
    content.push_str(&format!("- **Production Dependencies:** {}\n", prod_count));
    content.push_str(&format!("- **Development Dependencies:** {}\n\n", dev_count));
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

fn generate_dependencies_doc(deps: &serde_json::Value, graph: &str) -> String {
    let mut content = String::from("# Dependencies\n\n");
    
    // Extract stats
    if let Some(stats) = deps.get("stats") {
        let prod = stats.get("totalProduction").and_then(|v| v.as_u64()).unwrap_or(0);
        let dev = stats.get("totalDevelopment").and_then(|v| v.as_u64()).unwrap_or(0);
        content.push_str("## Summary\n\n");
        content.push_str(&format!("- **Production:** {} packages\n", prod));
        content.push_str(&format!("- **Development:** {} packages\n", dev));
        content.push_str(&format!("- **Total:** {} packages\n\n", prod + dev));
    }
    
    // Extract production dependencies
    if let Some(production) = deps.get("production").and_then(|v| v.as_object()) {
        if !production.is_empty() {
            content.push_str("## Production Dependencies\n\n");
            content.push_str("| Package | Version |\n");
            content.push_str("|---------|----------|\n");
            for (name, version) in production.iter().take(50) {
                let ver_str = version.as_str().unwrap_or("");
                content.push_str(&format!("| `{}` | `{}` |\n", name, ver_str));
            }
            if production.len() > 50 {
                content.push_str(&format!("\n*... and {} more*\n\n", production.len() - 50));
            } else {
                content.push_str("\n");
            }
        }
    }
    
    // Extract development dependencies
    if let Some(development) = deps.get("development").and_then(|v| v.as_object()) {
        if !development.is_empty() {
            content.push_str("## Development Dependencies\n\n");
            content.push_str("| Package | Version |\n");
            content.push_str("|---------|----------|\n");
            for (name, version) in development.iter().take(50) {
                let ver_str = version.as_str().unwrap_or("");
                content.push_str(&format!("| `{}` | `{}` |\n", name, ver_str));
            }
            if development.len() > 50 {
                content.push_str(&format!("\n*... and {} more*\n\n", development.len() - 50));
            } else {
                content.push_str("\n");
            }
        }
    }
    
    content.push_str("## Dependency Graph\n\n");
    content.push_str("```mermaid\n");
    content.push_str(graph);
    content.push_str("\n```\n\n");
    content.push_str("*For detailed dependency information, use the `analyze_dependencies` tool.*\n");
    content
}

fn generate_security_doc(security: &serde_json::Value) -> String {
    let mut content = String::from("# Security Audit\n\n");
    
    if let Some(risk_level) = security.get("riskLevel").and_then(|v| v.as_str()) {
        content.push_str(&format!("## Risk Level: {}\n\n", risk_level.to_uppercase()));
    }
    
    if let Some(summary) = security.get("summary") {
        let critical = summary.get("critical").and_then(|v| v.as_u64()).unwrap_or(0);
        let high = summary.get("high").and_then(|v| v.as_u64()).unwrap_or(0);
        let medium = summary.get("medium").and_then(|v| v.as_u64()).unwrap_or(0);
        let low = summary.get("low").and_then(|v| v.as_u64()).unwrap_or(0);
        
        content.push_str("## Summary\n\n");
        content.push_str(&format!("- **Critical:** {}\n", critical));
        content.push_str(&format!("- **High:** {}\n", high));
        content.push_str(&format!("- **Medium:** {}\n", medium));
        content.push_str(&format!("- **Low:** {}\n", low));
        
        if let Some(vulns) = security.get("vulnerabilities").and_then(|v| v.as_array()) {
            let total = vulns.len();
            content.push_str(&format!("- **Total Issues:** {}\n\n", total));
            
            if !vulns.is_empty() {
                content.push_str("## Vulnerabilities\n\n");
                content.push_str("| Package | Version | Severity | Description |\n");
                content.push_str("|---------|---------|----------|-------------|\n");
                for vuln in vulns.iter().take(100) {
                    let pkg = vuln.get("package").and_then(|v| v.as_str()).unwrap_or("");
                    let ver = vuln.get("version").and_then(|v| v.as_str()).unwrap_or("");
                    let sev = vuln.get("severity").and_then(|v| v.as_str()).unwrap_or("");
                    let desc = vuln.get("description").and_then(|v| v.as_str()).unwrap_or("");
                    content.push_str(&format!("| `{}` | `{}` | {} | {} |\n", pkg, ver, sev, desc));
                }
                if vulns.len() > 100 {
                    content.push_str(&format!("\n*... and {} more vulnerabilities*\n", vulns.len() - 100));
                }
            } else {
                content.push_str("## Status\n\n✅ No known vulnerabilities found.\n");
            }
        }
    } else {
        content.push_str("*For detailed security information, use the `audit_security` tool.*\n");
    }
    
    content
}

fn generate_licenses_doc(licenses: &serde_json::Value) -> String {
    let mut content = String::from("# License Compliance\n\n");
    
    if let Some(summary) = licenses.get("summary") {
        let total = summary.get("totalPackages").and_then(|v| v.as_u64()).unwrap_or(0);
        let unique = summary.get("uniqueLicenses").and_then(|v| v.as_u64()).unwrap_or(0);
        let unknown = summary.get("unknownLicenses").and_then(|v| v.as_u64()).unwrap_or(0);
        
        content.push_str("## Summary\n\n");
        content.push_str(&format!("- **Total Packages:** {}\n", total));
        content.push_str(&format!("- **Unique Licenses:** {}\n", unique));
        content.push_str(&format!("- **Unknown Licenses:** {}\n\n", unknown));
    }
    
    if let Some(dist) = licenses.get("distribution").and_then(|v| v.as_array()) {
        if !dist.is_empty() {
            content.push_str("## License Distribution\n\n");
            content.push_str("| License | Count | Percentage |\n");
            content.push_str("|---------|-------|------------|\n");
            let total = licenses.get("summary")
                .and_then(|s| s.get("totalPackages"))
                .and_then(|v| v.as_u64())
                .unwrap_or(1) as f64;
            for item in dist {
                let license = item.get("license").and_then(|v| v.as_str()).unwrap_or("");
                let count = item.get("count").and_then(|v| v.as_u64()).unwrap_or(0);
                let pct = if total > 0.0 { (count as f64 / total) * 100.0 } else { 0.0 };
                content.push_str(&format!("| {} | {} | {:.1}% |\n", license, count, pct));
            }
            content.push_str("\n");
        }
    }
    
    if let Some(issues) = licenses.get("compatibilityIssues").and_then(|v| v.as_array()) {
        if !issues.is_empty() {
            content.push_str("## Compatibility Issues\n\n");
            for issue in issues {
                if let Some(issue_str) = issue.as_str() {
                    content.push_str(&format!("- {}\n", issue_str));
                }
            }
            content.push_str("\n");
        }
    }
    
    if content == "# License Compliance\n\n" {
        content.push_str("*For detailed license information, use the `analyze_licenses` tool.*\n");
    }
    
    content
}

fn generate_architecture_doc(psr4: &serde_json::Value, namespaces: &serde_json::Value) -> String {
    let mut content = String::from("# Architecture\n\n");
    
    // Parse PSR-4 data
    if let Some(stats) = psr4.get("stats") {
        let mappings = stats.get("totalMappings").and_then(|v| v.as_u64()).unwrap_or(0);
        let files = stats.get("totalFiles").and_then(|v| v.as_u64()).unwrap_or(0);
        let valid = stats.get("validFiles").and_then(|v| v.as_u64()).unwrap_or(0);
        let violations = stats.get("violationCount").and_then(|v| v.as_u64()).unwrap_or(0);
        
        content.push_str("## PSR-4 Autoloading\n\n");
        content.push_str("### Summary\n\n");
        content.push_str(&format!("- **Total Mappings:** {}\n", mappings));
        content.push_str(&format!("- **Files Analyzed:** {}\n", files));
        content.push_str(&format!("- **PSR-4 Compliant:** {}\n", valid));
        content.push_str(&format!("- **Violations:** {}\n\n", violations));
        
        if let Some(mapping_list) = psr4.get("mappings").and_then(|v| v.as_array()) {
            if !mapping_list.is_empty() {
                content.push_str("### Mappings\n\n");
                content.push_str("| Namespace Prefix | Directory |\n");
                content.push_str("|------------------|-----------|\n");
                for mapping in mapping_list.iter().take(20) {
                    let ns = mapping.get("namespace").and_then(|v| v.as_str()).unwrap_or("");
                    let paths = mapping.get("paths")
                        .and_then(|v| v.as_array())
                        .map(|arr| arr.iter()
                            .filter_map(|p| p.as_str())
                            .collect::<Vec<_>>()
                            .join(", "))
                        .unwrap_or_default();
                    content.push_str(&format!("| `{}` | `{}` |\n", ns, paths));
                }
                if mapping_list.len() > 20 {
                    content.push_str(&format!("\n*... and {} more mappings*\n", mapping_list.len() - 20));
                }
                content.push_str("\n");
            }
        }
    } else {
        content.push_str("## PSR-4 Autoloading\n\n");
        content.push_str("*For detailed PSR-4 information, use the `analyze_psr4` tool.*\n\n");
    }
    
    // Parse namespaces data
    if let Some(ns_list) = namespaces.get("namespaces").and_then(|v| v.as_array()) {
        content.push_str("## Namespaces\n\n");
        if !ns_list.is_empty() {
            content.push_str(&format!("Found **{}** namespaces:\n\n", ns_list.len()));
            for ns in ns_list.iter().take(30) {
                let ns_name = ns.get("namespace").and_then(|v| v.as_str()).unwrap_or("");
                let files = ns.get("files").and_then(|v| v.as_array()).map(|a| a.len()).unwrap_or(0);
                content.push_str(&format!("- `{}` ({} files)\n", ns_name, files));
            }
            if ns_list.len() > 30 {
                content.push_str(&format!("\n*... and {} more namespaces*\n", ns_list.len() - 30));
            }
        } else {
            content.push_str("*No namespaces detected.*\n");
        }
    } else {
        content.push_str("## Namespaces\n\n");
        content.push_str("*For detailed namespace information, use the `detect_namespaces` tool.*\n");
    }
    
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

fn generate_html_site(
    site_name: &str,
    site_description: &str,
    index: &str,
    dependencies: &str,
    security: &str,
    licenses: &str,
    architecture: &str,
    changelog: &str,
) -> String {
    // Escape markdown content for JavaScript strings
    let escape_js = |s: &str| -> String {
        s.replace('\\', "\\\\")
         .replace('`', "\\`")
         .replace('$', "\\$")
         .replace('\n', "\\n")
         .replace('\r', "")
    };
    
    let index_escaped = escape_js(index);
    let deps_escaped = escape_js(dependencies);
    let sec_escaped = escape_js(security);
    let lic_escaped = escape_js(licenses);
    let arch_escaped = escape_js(architecture);
    
    let changelog_nav = if !changelog.is_empty() {
        "\n    <a href=\"#changelog\">Changelog</a>"
    } else {
        ""
    };
    
    let changelog_section = if !changelog.is_empty() {
        "\n  <div id=\"changelog\" class=\"section\">\n    <h2>Changelog</h2>\n    <div id=\"changelog-content\"></div>\n  </div>"
    } else {
        ""
    };
    
    // Build HTML string piece by piece to avoid format! macro issues with nested {}
    let mut html = String::new();
    html.push_str("<!DOCTYPE html>\n");
    html.push_str("<html lang=\"en\">\n");
    html.push_str("<head>\n");
    html.push_str("  <meta charset=\"UTF-8\">\n");
    html.push_str("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n");
    html.push_str(&format!("  <title>{}</title>\n", site_name));
    html.push_str("  <style>\n");
    html.push_str("    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; line-height: 1.6; }\n");
    html.push_str("    nav { background: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px; }\n");
    html.push_str("    nav a { margin-right: 20px; text-decoration: none; color: #0066cc; font-weight: 500; }\n");
    html.push_str("    nav a:hover { text-decoration: underline; }\n");
    html.push_str("    h1 { color: #333; border-bottom: 2px solid #0066cc; padding-bottom: 10px; }\n");
    html.push_str("    h2 { color: #555; margin-top: 30px; border-bottom: 1px solid #ddd; padding-bottom: 5px; }\n");
    html.push_str("    h3 { color: #666; margin-top: 20px; }\n");
    html.push_str("    code { background: #f5f5f5; padding: 2px 6px; border-radius: 3px; font-family: 'Courier New', monospace; }\n");
    html.push_str("    pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; border-left: 3px solid #0066cc; }\n");
    html.push_str("    table { border-collapse: collapse; width: 100%; margin: 20px 0; }\n");
    html.push_str("    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }\n");
    html.push_str("    th { background: #f5f5f5; font-weight: 600; }\n");
    html.push_str("    tr:nth-child(even) { background: #fafafa; }\n");
    html.push_str("    a { color: #0066cc; }\n");
    html.push_str("    .section { margin-bottom: 40px; }\n");
    html.push_str("    .meta { color: #666; font-size: 0.9em; margin-bottom: 20px; }\n");
    html.push_str("  </style>\n");
    html.push_str("  <script src=\"https://cdn.jsdelivr.net/npm/marked/marked.min.js\"></script>\n");
    html.push_str("</head>\n");
    html.push_str("<body>\n");
    html.push_str("  <nav>\n");
    html.push_str("    <a href=\"#index\">Home</a>\n");
    html.push_str("    <a href=\"#dependencies\">Dependencies</a>\n");
    html.push_str("    <a href=\"#security\">Security</a>\n");
    html.push_str("    <a href=\"#licenses\">Licenses</a>\n");
    html.push_str("    <a href=\"#architecture\">Architecture</a>");
    html.push_str(changelog_nav);
    html.push_str("\n  </nav>\n");
    html.push_str("  \n");
    html.push_str(&format!("  <div id=\"index\" class=\"section\">\n    <h1>{}</h1>\n    <p class=\"meta\">{}</p>\n    <div id=\"index-content\"></div>\n  </div>\n", site_name, site_description));
    html.push_str("  \n");
    html.push_str("  <div id=\"dependencies\" class=\"section\">\n    <h2>Dependencies</h2>\n    <div id=\"dependencies-content\"></div>\n  </div>\n");
    html.push_str("  \n");
    html.push_str("  <div id=\"security\" class=\"section\">\n    <h2>Security</h2>\n    <div id=\"security-content\"></div>\n  </div>\n");
    html.push_str("  \n");
    html.push_str("  <div id=\"licenses\" class=\"section\">\n    <h2>Licenses</h2>\n    <div id=\"licenses-content\"></div>\n  </div>\n");
    html.push_str("  \n");
    html.push_str("  <div id=\"architecture\" class=\"section\">\n    <h2>Architecture</h2>\n    <div id=\"architecture-content\"></div>\n  </div>");
    html.push_str(changelog_section);
    html.push_str("\n  \n");
    html.push_str("  <script>\n");
    html.push_str("    function markdownToHTML(md) {\n");
    html.push_str("      if (typeof marked !== 'undefined') {\n");
    html.push_str("        return marked.parse(md);\n");
    html.push_str("      }\n");
    html.push_str("      return md\n");
    html.push_str("        .replace(/^# (.*$)/gim, '<h1>$1</h1>')\n");
    html.push_str("        .replace(/^## (.*$)/gim, '<h2>$1</h2>')\n");
    html.push_str("        .replace(/^### (.*$)/gim, '<h3>$1</h3>')\n");
    html.push_str("        .replace(/\\*\\*(.*?)\\*\\*/gim, '<strong>$1</strong>')\n");
    html.push_str("        .replace(/\\*(.*?)\\*/gim, '<em>$1</em>')\n");
    html.push_str("        .replace(/`([^`]+)`/gim, '<code>$1</code>')\n");
    html.push_str("        .replace(/\\n/gim, '<br>');\n");
    html.push_str("    }\n");
    html.push_str("    \n");
    html.push_str(&format!("    const indexMD = \"{}\";\n", index_escaped.replace('"', "\\\"")));
    html.push_str(&format!("    const depsMD = \"{}\";\n", deps_escaped.replace('"', "\\\"")));
    html.push_str(&format!("    const secMD = \"{}\";\n", sec_escaped.replace('"', "\\\"")));
    html.push_str(&format!("    const licMD = \"{}\";\n", lic_escaped.replace('"', "\\\"")));
    html.push_str(&format!("    const archMD = \"{}\";\n", arch_escaped.replace('"', "\\\"")));
    
    if !changelog.is_empty() {
        let changelog_escaped = escape_js(changelog).replace('"', "\\\"");
        html.push_str(&format!(
            r#"
    const changelogMD = "{}";
    
    document.getElementById('index-content').innerHTML = markdownToHTML(indexMD);
    document.getElementById('dependencies-content').innerHTML = markdownToHTML(depsMD);
    document.getElementById('security-content').innerHTML = markdownToHTML(secMD);
    document.getElementById('licenses-content').innerHTML = markdownToHTML(licMD);
    document.getElementById('architecture-content').innerHTML = markdownToHTML(archMD);
    document.getElementById('changelog-content').innerHTML = markdownToHTML(changelogMD);
  </script>
</body>
</html>"#,
            changelog_escaped
        ));
    } else {
        html.push_str(
            r#"
    document.getElementById('index-content').innerHTML = markdownToHTML(indexMD);
    document.getElementById('dependencies-content').innerHTML = markdownToHTML(depsMD);
    document.getElementById('security-content').innerHTML = markdownToHTML(secMD);
    document.getElementById('licenses-content').innerHTML = markdownToHTML(licMD);
    document.getElementById('architecture-content').innerHTML = markdownToHTML(archMD);
  </script>
</body>
</html>"#,
        );
    }
    
    html
}
