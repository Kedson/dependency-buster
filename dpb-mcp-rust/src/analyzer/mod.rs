pub mod dependency;
pub mod namespace;
pub mod psr4;
pub mod security;
pub mod tracker;
pub mod suggestions;

use anyhow::Result;
use serde::Serialize;
use std::collections::HashMap;
use std::fs;
use std::path::Path;

use crate::composer::{get_licenses, read_composer_json, read_composer_lock};
use crate::types::{RepoConfig, RepoVersion, VersionConflict};

pub fn generate_dependency_graph<P: AsRef<Path>>(
    repo_path: P,
    max_depth: usize,
    include_dev: bool,
    focus_package: Option<String>,
) -> Result<String> {
    let lock = match read_composer_lock(&repo_path) {
        Ok(l) => l,
        Err(_) => return Ok("graph TD\n  NoLock[composer.lock not found]".to_string()),
    };

    let max_depth = if max_depth == 0 { 2 } else { max_depth };

    let mut mermaid = String::from("graph TD\n");

    let mut packages = lock.packages.clone();
    if include_dev {
        if let Some(dev_packages) = &lock.packages_dev {
            packages.extend(dev_packages.clone());
        }
    }

    if let Some(focus) = focus_package {
        mermaid.push_str(&generate_focused_graph(&packages, &focus, max_depth));
    } else {
        mermaid.push_str(&generate_full_graph(&packages, max_depth));
    }

    Ok(mermaid)
}

fn generate_full_graph(packages: &[crate::types::PackageInfo], max_depth: usize) -> String {
    let mut result = String::from("  Root[Your Application]\n");

    let limit = packages.len().min(15);

    for pkg in &packages[..limit] {
        let sanitized = sanitize_for_mermaid(&pkg.name);
        result.push_str(&format!(
            "  Root --> {}[\"{}...
{}\"]
",
            sanitized, pkg.name, pkg.version
        ));

        if max_depth > 1 {
            if let Some(require) = &pkg.require {
                let mut dep_count = 0;
                for (dep, version) in require {
                    if !dep.starts_with("php") && !dep.starts_with("ext-") && dep_count < 3 {
                        let dep_sanitized = sanitize_for_mermaid(dep);
                        result.push_str(&format!(
                            "  {} --> {}[\"{}...
{}\"]
",
                            sanitized, dep_sanitized, dep, version
                        ));
                        dep_count += 1;
                    }
                }
            }
        }
    }

    result
}

fn generate_focused_graph(
    packages: &[crate::types::PackageInfo],
    focus_package: &str,
    _max_depth: usize,
) -> String {
    let focus_sanitized = sanitize_for_mermaid(focus_package);
    let mut result = format!("  {}[{}]\n", focus_sanitized, focus_package);

    if let Some(pkg) = packages.iter().find(|p| p.name == focus_package) {
        if let Some(require) = &pkg.require {
            for (dep, version) in require {
                if !dep.starts_with("php") && !dep.starts_with("ext-") {
                    let dep_sanitized = sanitize_for_mermaid(dep);
                    result.push_str(&format!(
                        "  {} --> {}[\"{}...
{}\"]
",
                        focus_sanitized, dep_sanitized, dep, version
                    ));
                }
            }
        }
    }

    result
}

fn sanitize_for_mermaid(name: &str) -> String {
    name.replace('/', "_")
        .replace('-', "_")
        .replace('.', "_")
        .replace('@', "_")
}

#[allow(dead_code)]
#[derive(Debug, Serialize)]
pub struct MultiRepoAnalysisResult {
    pub repositories: Vec<RepoConfig>,
    #[serde(rename = "sharedDependencies")]
    pub shared_dependencies: HashMap<String, Vec<String>>,
    #[serde(rename = "versionConflicts")]
    pub version_conflicts: Vec<VersionConflict>,
    #[serde(rename = "totalPackages")]
    pub total_packages: usize,
    #[serde(rename = "commonLicenses")]
    pub common_licenses: HashMap<String, usize>,
}

pub fn analyze_multiple_repositories<P: AsRef<Path>>(config_path: P) -> Result<String> {
    let contents = fs::read_to_string(&config_path)?;
    let repos: Vec<RepoConfig> = serde_json::from_str(&contents)?;

    let mut package_usage: HashMap<String, Vec<String>> = HashMap::new();
    let mut all_packages = std::collections::HashSet::new();
    let mut license_count: HashMap<String, usize> = HashMap::new();

    for repo in &repos {
        if let Ok(composer) = read_composer_json(&repo.path) {
            // Collect dependencies
            if let Some(require) = &composer.require {
                for (pkg, _) in require {
                    if pkg != "php" {
                        all_packages.insert(pkg.clone());
                        package_usage
                            .entry(pkg.clone())
                            .or_insert_with(Vec::new)
                            .push(repo.name.clone());
                    }
                }
            }

            if let Some(require_dev) = &composer.require_dev {
                for (pkg, _) in require_dev {
                    all_packages.insert(pkg.clone());
                    package_usage
                        .entry(pkg.clone())
                        .or_insert_with(Vec::new)
                        .push(repo.name.clone());
                }
            }

            // Collect licenses
            let licenses = get_licenses(&composer);
            for license in licenses {
                *license_count.entry(license).or_insert(0) += 1;
            }
        }
    }

    // Find shared dependencies
    let shared_dependencies: HashMap<String, Vec<String>> = package_usage
        .iter()
        .filter(|(_, repos)| repos.len() > 1)
        .map(|(k, v)| (k.clone(), v.clone()))
        .collect();

    // Find version conflicts
    let mut version_conflicts = Vec::new();
    for (pkg, used_by_repos) in &package_usage {
        let mut versions: HashMap<String, Vec<String>> = HashMap::new();

        for repo in used_by_repos {
            if let Some(repo_config) = repos.iter().find(|r| &r.name == repo) {
                if let Ok(composer) = read_composer_json(&repo_config.path) {
                    let version = composer
                        .require
                        .as_ref()
                        .and_then(|r| r.get(pkg))
                        .or_else(|| composer.require_dev.as_ref().and_then(|r| r.get(pkg)));

                    if let Some(v) = version {
                        versions
                            .entry(v.clone())
                            .or_insert_with(Vec::new)
                            .push(repo.clone());
                    }
                }
            }
        }

        if versions.len() > 1 {
            let mut conflict_versions = Vec::new();
            for (version, repos) in versions {
                for repo in repos {
                    conflict_versions.push(RepoVersion { repo, version: version.clone() });
                }
            }

            version_conflicts.push(VersionConflict {
                package: pkg.clone(),
                versions: conflict_versions,
            });
        }
    }

    // Generate markdown report
    let report = generate_multi_repo_report(
        &repos,
        &shared_dependencies,
        &version_conflicts,
        all_packages.len(),
        &license_count,
    );

    Ok(report)
}

fn generate_multi_repo_report(
    repos: &[RepoConfig],
    shared_deps: &HashMap<String, Vec<String>>,
    conflicts: &[VersionConflict],
    total_pkgs: usize,
    licenses: &HashMap<String, usize>,
) -> String {
    let mut report = String::from("# Multi-Repository Dependency Analysis\n\n");
    report.push_str(&format!("**Generated:** {}\n\n", chrono::Utc::now().to_rfc3339()));

    report.push_str("## Repositories Analyzed\n\n");
    for repo in repos {
        report.push_str(&format!("- **{}** ({})", repo.name, repo.repo_type));
        if let Some(team) = &repo.team {
            report.push_str(&format!(" - Team: {}", team));
        }
        if let Some(desc) = &repo.description {
            report.push_str(&format!("\n  {}", desc));
        }
        report.push('\n');
    }

    report.push_str("\n## Summary\n\n");
    report.push_str(&format!("- Total unique packages: {}\n", total_pkgs));
    report.push_str(&format!("- Shared dependencies: {}\n", shared_deps.len()));
    report.push_str(&format!("- Version conflicts: {}\n\n", conflicts.len()));

    if !shared_deps.is_empty() {
        report.push_str("## Shared Dependencies\n\n");
        report.push_str("| Package | Used By |\n");
        report.push_str("|---------|----------|\n");
        for (pkg, repos) in shared_deps {
            report.push_str(&format!("| {} | {} |\n", pkg, repos.join(", ")));
        }
        report.push('\n');
    }

    if !conflicts.is_empty() {
        report.push_str("## ⚠️ Version Conflicts\n\n");
        for conflict in conflicts {
            report.push_str(&format!("### {}\n\n", conflict.package));
            for version in &conflict.versions {
                report.push_str(&format!("- **{}**: {}\n", version.repo, version.version));
            }
            report.push('\n');
        }
    }

    if !licenses.is_empty() {
        report.push_str("## License Distribution\n\n");
        report.push_str("| License | Count |\n");
        report.push_str("|---------|-------|\n");
        for (license, count) in licenses {
            report.push_str(&format!("| {} | {} |\n", license, count));
        }
    }

    report
}

pub fn generate_comprehensive_docs<P: AsRef<Path>>(
    repo_path: P,
    output_path: Option<P>,
) -> Result<String> {
    let composer = read_composer_json(&repo_path)?;

    let mut report = String::from("# PHP Dependency Documentation\n\n");
    report.push_str(&format!("**Generated:** {}\n\n", chrono::Utc::now().to_rfc3339()));

    report.push_str("## Project Information\n\n");
    report.push_str(&format!(
        "- **Name:** {}\n",
        composer.name.as_deref().unwrap_or("Unknown")
    ));
    report.push_str(&format!(
        "- **Description:** {}\n",
        composer.description.as_deref().unwrap_or("N/A")
    ));
    report.push_str(&format!(
        "- **Type:** {}\n",
        composer.package_type.as_deref().unwrap_or("library")
    ));

    let licenses = get_licenses(&composer);
    let license_str = if licenses.is_empty() {
        "Not specified".to_string()
    } else {
        licenses.join(", ")
    };
    report.push_str(&format!("- **License:** {}\n\n", license_str));

    report.push_str("## Dependency Summary\n\n");
    let prod_count = composer.require.as_ref().map(|r| r.len()).unwrap_or(0);
    let dev_count = composer.require_dev.as_ref().map(|r| r.len()).unwrap_or(0);
    report.push_str(&format!("- **Production Dependencies:** {}\n", prod_count));
    report.push_str(&format!("- **Development Dependencies:** {}\n\n", dev_count));

    report.push_str("For detailed analysis, use the individual tools:\n");
    report.push_str("- `analyze_dependencies`\n");
    report.push_str("- `analyze_psr4`\n");
    report.push_str("- `audit_security`\n");
    report.push_str("- `analyze_licenses`\n");

    if let Some(output) = output_path {
        fs::write(output.as_ref(), &report)?;
        return Ok(format!("Documentation saved to: {}", output.as_ref().display()));
    }

    Ok(report)
}
