use anyhow::Result;
use lazy_static::lazy_static;
use rayon::prelude::*;
use regex::Regex;
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::{Path, PathBuf};
use std::sync::{Arc, Mutex};
use walkdir::WalkDir;

use crate::composer::{calculate_expected_namespace, get_psr4_mappings, read_composer_json};
use crate::types::{Psr4Mapping, Psr4Violation};

lazy_static! {
    static ref NAMESPACE_RE: Regex = Regex::new(r"namespace\s+([\w\\]+)\s*;").unwrap();
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Psr4AnalysisResult {
    pub mappings: Vec<Psr4Mapping>,
    pub violations: Vec<Psr4Violation>,
    pub stats: Psr4Stats,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Psr4Stats {
    #[serde(rename = "totalMappings")]
    pub total_mappings: usize,
    #[serde(rename = "totalFiles")]
    pub total_files: usize,
    #[serde(rename = "validFiles")]
    pub valid_files: usize,
    #[serde(rename = "violationCount")]
    pub violation_count: usize,
}

pub fn analyze_psr4_autoloading<P: AsRef<Path> + Sync>(repo_path: P) -> Result<String> {
    let composer_json = read_composer_json(&repo_path)?;
    let mappings = get_psr4_mappings(&composer_json);

    let violations = Arc::new(Mutex::new(Vec::new()));
    let total_files = Arc::new(Mutex::new(0usize));
    let valid_files = Arc::new(Mutex::new(0usize));

    // Process each mapping in parallel
    mappings.par_iter().for_each(|mapping| {
        for relative_path in &mapping.paths {
            let abs_path = repo_path.as_ref().join(relative_path);

            if let Ok(php_files) = find_php_files(&abs_path) {
                // Process files in parallel
                php_files.par_iter().for_each(|file| {
                    {
                        let mut count = total_files.lock().unwrap();
                        *count += 1;
                    }

                    if let Ok(namespace) = extract_namespace(file) {
                        if let Ok(rel_to_root) = file.strip_prefix(&abs_path) {
                            let expected_ns = calculate_expected_namespace(
                                &mapping.namespace,
                                &rel_to_root.to_string_lossy(),
                            );

                            if namespace == expected_ns {
                                let mut count = valid_files.lock().unwrap();
                                *count += 1;
                            } else {
                                let issue = if namespace.is_empty() {
                                    "Missing namespace declaration"
                                } else {
                                    "Namespace mismatch"
                                };

                                let mut viols = violations.lock().unwrap();
                                viols.push(Psr4Violation {
                                    file: PathBuf::from(relative_path)
                                        .join(rel_to_root)
                                        .to_string_lossy()
                                        .to_string(),
                                    expected_namespace: expected_ns,
                                    actual_namespace: Some(namespace),
                                    issue: issue.to_string(),
                                });
                            }
                        }
                    }
                });
            }
        }
    });

    let violations = Arc::try_unwrap(violations).unwrap().into_inner().unwrap();
    let total_files = *total_files.lock().unwrap();
    let valid_files = *valid_files.lock().unwrap();
    let total_mappings = mappings.len();
    let violation_count = violations.len();

    let result = Psr4AnalysisResult {
        mappings,
        violations,
        stats: Psr4Stats {
            total_mappings,
            total_files,
            valid_files,
            violation_count,
        },
    };

    Ok(serde_json::to_string_pretty(&result)?)
}

pub fn find_php_files(dir: &Path) -> Result<Vec<PathBuf>> {
    let files: Vec<PathBuf> = WalkDir::new(dir)
        .into_iter()
        .filter_entry(|e| {
            let name = e.file_name().to_string_lossy();
            !name.starts_with('.') && name != "vendor" && name != "node_modules"
        })
        .filter_map(|e| e.ok())
        .filter(|e| e.path().extension().and_then(|s| s.to_str()) == Some("php"))
        .map(|e| e.path().to_path_buf())
        .collect();

    Ok(files)
}

fn extract_namespace(file_path: &Path) -> Result<String> {
    let contents = fs::read_to_string(file_path)?;

    for line in contents.lines() {
        if let Some(captures) = NAMESPACE_RE.captures(line) {
            return Ok(captures[1].to_string());
        }
    }

    Ok(String::new())
}
