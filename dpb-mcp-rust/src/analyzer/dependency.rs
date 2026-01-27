use anyhow::Result;
use rayon::prelude::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::Path;
use std::sync::{Arc, Mutex};

use crate::composer::{filter_php_dependencies, read_composer_json, read_composer_lock};
use crate::types::{ComposerLock, DependencyNode};

#[derive(Debug, Serialize, Deserialize)]
pub struct DependencyAnalysisResult {
    pub production: HashMap<String, String>,
    pub development: HashMap<String, String>,
    pub tree: Vec<DependencyNode>,
    pub stats: DependencyStats,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DependencyStats {
    #[serde(rename = "totalProduction")]
    pub total_production: usize,
    #[serde(rename = "totalDevelopment")]
    pub total_development: usize,
    pub outdated: usize,
    #[serde(rename = "upToDate")]
    pub up_to_date: usize,
}

/// Analyze dependencies and return the raw struct
pub fn analyze_dependencies_raw<P: AsRef<Path>>(repo_path: P) -> Result<DependencyAnalysisResult> {
    let composer_json = read_composer_json(&repo_path)?;

    let lock = read_composer_lock(&repo_path).ok();

    let production = composer_json
        .require
        .map(|r| filter_php_dependencies(&r))
        .unwrap_or_default();

    let development = composer_json.require_dev.unwrap_or_default();

    let tree = if let Some(lock) = lock {
        build_dependency_tree(&lock)
    } else {
        Vec::new()
    };

    Ok(DependencyAnalysisResult {
        production: production.clone(),
        development: development.clone(),
        tree,
        stats: DependencyStats {
            total_production: production.len(),
            total_development: development.len(),
            outdated: 0,
            up_to_date: 0,
        },
    })
}

/// Analyze dependencies and return JSON string
pub fn analyze_dependencies<P: AsRef<Path>>(repo_path: P) -> Result<String> {
    let result = analyze_dependencies_raw(repo_path)?;
    Ok(serde_json::to_string_pretty(&result)?)
}

fn build_dependency_tree(lock: &ComposerLock) -> Vec<DependencyNode> {
    let mut all_packages = lock.packages.clone();
    if let Some(dev_packages) = &lock.packages_dev {
        all_packages.extend(dev_packages.clone());
    }

    let production_count = lock.packages.len();

    // Build tree in parallel using rayon
    let tree: Vec<DependencyNode> = all_packages
        .par_iter()
        .enumerate()
        .map(|(index, pkg)| {
            let node_type = if index < production_count {
                "production"
            } else {
                "development"
            };

            let deps: Vec<String> = pkg
                .require
                .as_ref()
                .map(|r| {
                    r.keys()
                        .filter(|name| !name.starts_with("php") && !name.starts_with("ext-"))
                        .cloned()
                        .collect()
                })
                .unwrap_or_default();

            let license = pkg
                .license
                .as_ref()
                .and_then(|l| l.first())
                .cloned();

            DependencyNode {
                name: pkg.name.clone(),
                version: pkg.version.clone(),
                node_type: node_type.to_string(),
                dependencies: deps,
                used_by: Vec::new(), // Will be filled in next step
                license,
            }
        })
        .collect();

    // Calculate reverse dependencies (used_by)
    let used_by_map: Arc<Mutex<HashMap<String, Vec<String>>>> =
        Arc::new(Mutex::new(HashMap::new()));

    tree.par_iter().for_each(|node| {
        for dep in &node.dependencies {
            let mut map = used_by_map.lock().unwrap();
            map.entry(dep.clone())
                .or_insert_with(Vec::new)
                .push(node.name.clone());
        }
    });

    let used_by_map = used_by_map.lock().unwrap();

    tree.into_iter()
        .map(|mut node| {
            node.used_by = used_by_map.get(&node.name).cloned().unwrap_or_default();
            node
        })
        .collect()
}

#[derive(Debug, Serialize)]
pub struct CircularDependenciesResult {
    pub cycles: Vec<Vec<String>>,
    pub count: usize,
}

pub fn find_circular_dependencies<P: AsRef<Path>>(repo_path: P) -> Result<String> {
    let lock = read_composer_lock(&repo_path)?;
    let tree = build_dependency_tree(&lock);

    let cycles = detect_cycles(&tree);

    let result = CircularDependenciesResult {
        count: cycles.len(),
        cycles,
    };

    Ok(serde_json::to_string_pretty(&result)?)
}

fn detect_cycles(tree: &[DependencyNode]) -> Vec<Vec<String>> {
    let mut cycles = Vec::new();
    let mut visited = HashMap::new();
    let mut rec_stack = HashMap::new();

    for node in tree {
        if !visited.contains_key(&node.name) {
            dfs(&node.name, tree, &mut visited, &mut rec_stack, &mut Vec::new(), &mut cycles);
        }
    }

    cycles
}

fn dfs(
    pkg_name: &str,
    tree: &[DependencyNode],
    visited: &mut HashMap<String, bool>,
    rec_stack: &mut HashMap<String, bool>,
    path: &mut Vec<String>,
    cycles: &mut Vec<Vec<String>>,
) {
    visited.insert(pkg_name.to_string(), true);
    rec_stack.insert(pkg_name.to_string(), true);
    path.push(pkg_name.to_string());

    if let Some(node) = tree.iter().find(|n| n.name == pkg_name) {
        for dep in &node.dependencies {
            if !visited.get(dep).unwrap_or(&false) {
                dfs(dep, tree, visited, rec_stack, path, cycles);
            } else if *rec_stack.get(dep).unwrap_or(&false) {
                // Found a cycle
                if let Some(start) = path.iter().position(|p| p == dep) {
                    let mut cycle = path[start..].to_vec();
                    cycle.push(dep.clone());
                    cycles.push(cycle);
                }
            }
        }
    }

    rec_stack.insert(pkg_name.to_string(), false);
    path.pop();
}
