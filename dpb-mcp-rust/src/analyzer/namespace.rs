use anyhow::Result;
use lazy_static::lazy_static;
use rayon::prelude::*;
use regex::Regex;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::path::Path;
use std::sync::{Arc, Mutex};

use crate::analyzer::psr4::find_php_files;
use crate::types::NamespaceInfo;

lazy_static! {
    static ref NAMESPACE_RE: Regex = Regex::new(r"namespace\s+([\w\\]+)\s*;").unwrap();
    static ref CLASS_RE: Regex = Regex::new(r"(?:abstract\s+)?class\s+(\w+)").unwrap();
    static ref INTERFACE_RE: Regex = Regex::new(r"interface\s+(\w+)").unwrap();
    static ref TRAIT_RE: Regex = Regex::new(r"trait\s+(\w+)").unwrap();
    static ref USE_RE: Regex = Regex::new(r"use\s+([\w\\]+)(?:\s+as\s+\w+)?;").unwrap();
}

#[derive(Debug, Serialize, Deserialize)]
pub struct NamespaceDetectionResult {
    pub namespaces: Vec<NamespaceInfo>,
    #[serde(rename = "totalFiles")]
    pub total_files: usize,
    #[serde(rename = "filesWithoutNamespace")]
    pub files_without_namespace: Vec<String>,
}

#[derive(Debug, Default)]
struct FileInfo {
    namespace: String,
    classes: Vec<String>,
    interfaces: Vec<String>,
    traits: Vec<String>,
    uses: Vec<String>,
}

pub fn detect_namespaces<P: AsRef<Path> + Sync>(repo_path: P) -> Result<String> {
    let php_files = find_php_files(repo_path.as_ref())?;

    let namespace_map: Arc<Mutex<HashMap<String, NamespaceInfo>>> =
        Arc::new(Mutex::new(HashMap::new()));
    let files_without: Arc<Mutex<Vec<String>>> = Arc::new(Mutex::new(Vec::new()));

    // Process files in parallel
    php_files.par_iter().for_each(|file| {
        if let Ok(info) = analyze_file(file) {
            if let Ok(relative_path) = file.strip_prefix(&repo_path) {
                let rel_str = relative_path.to_string_lossy().to_string();

                if !info.namespace.is_empty() {
                    let mut map = namespace_map.lock().unwrap();
                    let ns_info = map.entry(info.namespace.clone()).or_insert_with(|| {
                        NamespaceInfo {
                            namespace: info.namespace.clone(),
                            files: Vec::new(),
                            classes: Vec::new(),
                            interfaces: Vec::new(),
                            traits: Vec::new(),
                        }
                    });

                    ns_info.files.push(rel_str);
                    ns_info.classes.extend(info.classes);
                    ns_info.interfaces.extend(info.interfaces);
                    ns_info.traits.extend(info.traits);
                } else {
                    let mut files = files_without.lock().unwrap();
                    files.push(rel_str);
                }
            }
        }
    });

    let namespace_map = Arc::try_unwrap(namespace_map).unwrap().into_inner().unwrap();
    let namespaces: Vec<NamespaceInfo> = namespace_map.into_values().collect();
    let files_without_namespace = Arc::try_unwrap(files_without).unwrap().into_inner().unwrap();

    let result = NamespaceDetectionResult {
        namespaces,
        total_files: php_files.len(),
        files_without_namespace,
    };

    Ok(serde_json::to_string_pretty(&result)?)
}

fn analyze_file(file_path: &Path) -> Result<FileInfo> {
    let contents = fs::read_to_string(file_path)?;

    let mut info = FileInfo::default();

    for line in contents.lines() {
        // Extract namespace
        if let Some(captures) = NAMESPACE_RE.captures(line) {
            info.namespace = captures[1].to_string();
        }

        // Extract classes
        if let Some(captures) = CLASS_RE.captures(line) {
            info.classes.push(captures[1].to_string());
        }

        // Extract interfaces
        if let Some(captures) = INTERFACE_RE.captures(line) {
            info.interfaces.push(captures[1].to_string());
        }

        // Extract traits
        if let Some(captures) = TRAIT_RE.captures(line) {
            info.traits.push(captures[1].to_string());
        }

        // Extract use statements
        if let Some(captures) = USE_RE.captures(line) {
            info.uses.push(captures[1].to_string());
        }
    }

    Ok(info)
}

#[derive(Debug, Serialize)]
pub struct NamespaceUsageResult {
    #[serde(rename = "definedIn")]
    pub defined_in: Vec<String>,
    #[serde(rename = "importedBy")]
    pub imported_by: Vec<ImportInfo>,
    #[serde(rename = "totalUsages")]
    pub total_usages: usize,
}

#[derive(Debug, Serialize)]
pub struct ImportInfo {
    pub file: String,
    pub imports: Vec<String>,
}

pub fn analyze_namespace_usage<P: AsRef<Path> + Sync>(
    repo_path: P,
    target_namespace: &str,
) -> Result<String> {
    let php_files = find_php_files(repo_path.as_ref())?;

    let defined_in: Arc<Mutex<Vec<String>>> = Arc::new(Mutex::new(Vec::new()));
    let imported_by: Arc<Mutex<Vec<ImportInfo>>> = Arc::new(Mutex::new(Vec::new()));

    php_files.par_iter().for_each(|file| {
        if let Ok(info) = analyze_file(file) {
            if let Ok(relative_path) = file.strip_prefix(&repo_path) {
                let rel_str = relative_path.to_string_lossy().to_string();

                if info.namespace == target_namespace {
                    let mut defined = defined_in.lock().unwrap();
                    defined.push(rel_str.clone());
                }

                let relevant_imports: Vec<String> = info
                    .uses
                    .into_iter()
                    .filter(|u| u.starts_with(target_namespace))
                    .collect();

                if !relevant_imports.is_empty() {
                    let mut imported = imported_by.lock().unwrap();
                    imported.push(ImportInfo {
                        file: rel_str,
                        imports: relevant_imports,
                    });
                }
            }
        }
    });

    let defined_in = Arc::try_unwrap(defined_in).unwrap().into_inner().unwrap();
    let imported_by = Arc::try_unwrap(imported_by).unwrap().into_inner().unwrap();

    let result = NamespaceUsageResult {
        total_usages: defined_in.len() + imported_by.len(),
        defined_in,
        imported_by,
    };

    Ok(serde_json::to_string_pretty(&result)?)
}
