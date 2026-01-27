//! Dependency Tracker - Timestamps and versioning for dependency changes
//! Enables reverting or replacing non-compliant dependencies

use anyhow::Result;
use chrono::{DateTime, Duration, Utc};
use serde::{Deserialize, Serialize};
use sha2::{Sha256, Digest};
use std::collections::HashMap;
use std::fs;
use std::path::Path;

const TRACKER_FILE: &str = ".dpb-dependency-tracker.json";

/// Snapshot of all dependencies at a point in time
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DependencySnapshot {
    pub timestamp: String,
    pub checksum: String,
    pub dependencies: Vec<TrackedDependency>,
    pub metadata: SnapshotMetadata,
}

/// A single tracked dependency with timestamps
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrackedDependency {
    pub name: String,
    pub version: String,
    #[serde(rename = "type")]
    pub dep_type: String, // "production" or "development"
    #[serde(skip_serializing_if = "Option::is_none")]
    pub added_at: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub updated_at: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub license: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub security_status: Option<String>,
}

/// Metadata about the snapshot
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SnapshotMetadata {
    pub repo_path: String,
    pub package_manager: String,
    pub total_count: usize,
}

/// A change between two snapshots
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DependencyChange {
    #[serde(rename = "type")]
    pub change_type: String, // "added", "removed", "updated"
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub old_version: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub new_version: Option<String>,
    pub timestamp: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub reason: Option<String>,
}

/// A compliance issue with a dependency
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ComplianceIssue {
    pub dependency: String,
    pub version: String,
    pub issue: String, // "license", "security", "outdated", "deprecated"
    pub severity: String, // "critical", "high", "medium", "low"
    pub description: String,
    pub recommendation: String,
    pub auto_fix_available: bool,
}

/// Dependency history with categorization
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DependencyHistory {
    pub current_snapshot: DependencySnapshot,
    pub recently_added: Vec<TrackedDependency>,
    pub recently_updated: Vec<TrackedDependency>,
    pub stale: Vec<TrackedDependency>,
}

/// Create a snapshot of current dependencies
pub fn create_dependency_snapshot(repo_path: &str) -> Result<DependencySnapshot> {
    let deps = super::analyze_dependencies(repo_path)?;
    let now = Utc::now().to_rfc3339();
    
    // Load existing tracker to preserve timestamps
    let existing = load_tracker(repo_path).ok();
    let existing_deps: HashMap<String, TrackedDependency> = existing
        .as_ref()
        .map(|s| s.dependencies.iter().map(|d| (d.name.clone(), d.clone())).collect())
        .unwrap_or_default();
    
    let mut tracked: Vec<TrackedDependency> = Vec::new();
    
    for pkg in &deps.tree {
        let existing_dep = existing_deps.get(&pkg.name);
        
        let added_at = existing_dep
            .and_then(|e| e.added_at.clone())
            .unwrap_or_else(|| now.clone());
        
        let updated_at = if existing_dep.map(|e| &e.version) != Some(&pkg.version) {
            now.clone()
        } else {
            existing_dep
                .and_then(|e| e.updated_at.clone())
                .unwrap_or_else(|| now.clone())
        };
        
        tracked.push(TrackedDependency {
            name: pkg.name.clone(),
            version: pkg.version.clone(),
            dep_type: pkg.dep_type.clone().unwrap_or_else(|| "production".to_string()),
            added_at: Some(added_at),
            updated_at: Some(updated_at),
            license: pkg.license.clone(),
            security_status: Some("unknown".to_string()),
        });
    }
    
    // Calculate checksum
    let mut names: Vec<String> = tracked
        .iter()
        .map(|d| format!("{}@{}", d.name, d.version))
        .collect();
    names.sort();
    
    let mut hasher = Sha256::new();
    hasher.update(names.join("|"));
    let hash = hasher.finalize();
    let checksum = hex::encode(&hash[..8]);
    
    Ok(DependencySnapshot {
        timestamp: now,
        checksum,
        dependencies: tracked.clone(),
        metadata: SnapshotMetadata {
            repo_path: repo_path.to_string(),
            package_manager: "composer".to_string(),
            total_count: tracked.len(),
        },
    })
}

/// Load existing tracker from file
pub fn load_tracker(repo_path: &str) -> Result<DependencySnapshot> {
    let tracker_path = Path::new(repo_path).join(TRACKER_FILE);
    let content = fs::read_to_string(tracker_path)?;
    let snapshot: DependencySnapshot = serde_json::from_str(&content)?;
    Ok(snapshot)
}

/// Save snapshot to tracker file
pub fn save_snapshot(repo_path: &str, snapshot: &DependencySnapshot) -> Result<()> {
    let tracker_path = Path::new(repo_path).join(TRACKER_FILE);
    let content = serde_json::to_string_pretty(snapshot)?;
    fs::write(tracker_path, content)?;
    Ok(())
}

/// Compare two snapshots and return changes
pub fn compare_snapshots(old: &DependencySnapshot, new: &DependencySnapshot) -> Vec<DependencyChange> {
    let mut changes = Vec::new();
    
    let old_deps: HashMap<&str, &TrackedDependency> = old
        .dependencies
        .iter()
        .map(|d| (d.name.as_str(), d))
        .collect();
    
    let new_deps: HashMap<&str, &TrackedDependency> = new
        .dependencies
        .iter()
        .map(|d| (d.name.as_str(), d))
        .collect();
    
    // Find added and updated
    for (name, new_dep) in &new_deps {
        if let Some(old_dep) = old_deps.get(name) {
            if old_dep.version != new_dep.version {
                changes.push(DependencyChange {
                    change_type: "updated".to_string(),
                    name: name.to_string(),
                    old_version: Some(old_dep.version.clone()),
                    new_version: Some(new_dep.version.clone()),
                    timestamp: new.timestamp.clone(),
                    reason: None,
                });
            }
        } else {
            changes.push(DependencyChange {
                change_type: "added".to_string(),
                name: name.to_string(),
                old_version: None,
                new_version: Some(new_dep.version.clone()),
                timestamp: new.timestamp.clone(),
                reason: None,
            });
        }
    }
    
    // Find removed
    for (name, old_dep) in &old_deps {
        if !new_deps.contains_key(name) {
            changes.push(DependencyChange {
                change_type: "removed".to_string(),
                name: name.to_string(),
                old_version: Some(old_dep.version.clone()),
                new_version: None,
                timestamp: new.timestamp.clone(),
                reason: None,
            });
        }
    }
    
    changes
}

/// Get dependency history with categorization
pub fn get_dependency_history(repo_path: &str) -> Result<DependencyHistory> {
    let snapshot = create_dependency_snapshot(repo_path)?;
    let now = Utc::now();
    let thirty_days_ago = now - Duration::days(30);
    let one_year_ago = now - Duration::days(365);
    
    let mut recently_added = Vec::new();
    let mut recently_updated = Vec::new();
    let mut stale = Vec::new();
    
    for dep in &snapshot.dependencies {
        if let Some(ref added_at) = dep.added_at {
            if let Ok(added_time) = DateTime::parse_from_rfc3339(added_at) {
                if added_time.with_timezone(&Utc) > thirty_days_ago {
                    recently_added.push(dep.clone());
                }
            }
        }
        
        if let Some(ref updated_at) = dep.updated_at {
            if let Ok(updated_time) = DateTime::parse_from_rfc3339(updated_at) {
                let updated_utc = updated_time.with_timezone(&Utc);
                if updated_utc > thirty_days_ago && dep.updated_at != dep.added_at {
                    recently_updated.push(dep.clone());
                }
                if updated_utc < one_year_ago {
                    stale.push(dep.clone());
                }
            }
        }
    }
    
    Ok(DependencyHistory {
        current_snapshot: snapshot,
        recently_added,
        recently_updated,
        stale,
    })
}

/// Check dependencies for compliance issues
pub fn check_compliance(repo_path: &str) -> Result<Vec<ComplianceIssue>> {
    let snapshot = create_dependency_snapshot(repo_path)?;
    let mut issues = Vec::new();
    
    let restrictive_licenses = ["GPL-3.0", "AGPL-3.0", "GPL-2.0", "SSPL"];
    
    for dep in &snapshot.dependencies {
        // Check for restrictive licenses
        if dep.dep_type == "production" {
            if let Some(ref license) = dep.license {
                for restricted in &restrictive_licenses {
                    if license.to_uppercase().contains(&restricted.to_uppercase()) {
                        issues.push(ComplianceIssue {
                            dependency: dep.name.clone(),
                            version: dep.version.clone(),
                            issue: "license".to_string(),
                            severity: "high".to_string(),
                            description: format!("Uses restrictive license: {}", license),
                            recommendation: "Consider replacing with an MIT/Apache-2.0 licensed alternative".to_string(),
                            auto_fix_available: false,
                        });
                    }
                }
            }
        }
        
        // Check for stale dependencies
        if let Some(ref updated_at) = dep.updated_at {
            if let Ok(updated_time) = DateTime::parse_from_rfc3339(updated_at) {
                let two_years_ago = Utc::now() - Duration::days(730);
                if updated_time.with_timezone(&Utc) < two_years_ago {
                    issues.push(ComplianceIssue {
                        dependency: dep.name.clone(),
                        version: dep.version.clone(),
                        issue: "outdated".to_string(),
                        severity: "low".to_string(),
                        description: "Not updated in over 2 years".to_string(),
                        recommendation: "Check if a newer version is available".to_string(),
                        auto_fix_available: true,
                    });
                }
            }
        }
    }
    
    Ok(issues)
}

/// Generate command to revert a dependency change
pub fn generate_revert_command(change: &DependencyChange) -> String {
    match change.change_type.as_str() {
        "added" => format!("composer remove {}", change.name),
        "removed" => format!("composer require {}:{}", change.name, change.old_version.as_deref().unwrap_or("*")),
        "updated" => format!("composer require {}:{}", change.name, change.old_version.as_deref().unwrap_or("*")),
        _ => String::new(),
    }
}
