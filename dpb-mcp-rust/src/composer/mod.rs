use anyhow::{Context, Result};
use std::collections::HashMap;
use std::fs;
use std::path::Path;

use crate::types::{ComposerJson, ComposerLock, LicenseField, Psr4Mapping, Psr4Path};

pub fn read_composer_json<P: AsRef<Path>>(repo_path: P) -> Result<ComposerJson> {
    let composer_path = repo_path.as_ref().join("composer.json");
    let contents = fs::read_to_string(&composer_path)
        .with_context(|| format!("Failed to read composer.json at {:?}", composer_path))?;

    let composer: ComposerJson = serde_json::from_str(&contents)
        .with_context(|| "Failed to parse composer.json")?;

    Ok(composer)
}

pub fn read_composer_lock<P: AsRef<Path>>(repo_path: P) -> Result<ComposerLock> {
    let lock_path = repo_path.as_ref().join("composer.lock");
    let contents = fs::read_to_string(&lock_path)
        .with_context(|| format!("Failed to read composer.lock at {:?}", lock_path))?;

    let lock: ComposerLock = serde_json::from_str(&contents)
        .with_context(|| "Failed to parse composer.lock")?;

    Ok(lock)
}

pub fn get_psr4_mappings(composer: &ComposerJson) -> Vec<Psr4Mapping> {
    let mut mappings = Vec::new();

    // Production autoload
    if let Some(autoload) = &composer.autoload {
        if let Some(psr4) = &autoload.psr4 {
            for (namespace, paths) in psr4 {
                mappings.push(Psr4Mapping {
                    namespace: namespace.clone(),
                    paths: normalize_paths(paths),
                    mapping_type: "psr-4".to_string(),
                    is_dev: false,
                });
            }
        }
    }

    // Dev autoload
    if let Some(autoload_dev) = &composer.autoload_dev {
        if let Some(psr4) = &autoload_dev.psr4 {
            for (namespace, paths) in psr4 {
                mappings.push(Psr4Mapping {
                    namespace: namespace.clone(),
                    paths: normalize_paths(paths),
                    mapping_type: "psr-4".to_string(),
                    is_dev: true,
                });
            }
        }
    }

    mappings
}

fn normalize_paths(paths: &Psr4Path) -> Vec<String> {
    match paths {
        Psr4Path::Single(s) => vec![s.clone()],
        Psr4Path::Multiple(v) => v.clone(),
    }
}

pub fn get_licenses(composer: &ComposerJson) -> Vec<String> {
    match &composer.license {
        Some(LicenseField::Single(s)) => vec![s.clone()],
        Some(LicenseField::Multiple(v)) => v.clone(),
        None => vec![],
    }
}

pub fn calculate_expected_namespace(base_namespace: &str, relative_file_path: &str) -> String {
    // Remove .php extension
    let without_ext = relative_file_path.trim_end_matches(".php");

    // Split by directory separator
    let parts: Vec<&str> = without_ext.split('/').collect();

    // Remove the filename (last part) to get directory structure
    let dir_parts = if parts.len() > 1 {
        &parts[..parts.len() - 1]
    } else {
        &[]
    };

    // Build expected namespace
    let namespace = base_namespace.trim_end_matches('\\');

    if dir_parts.is_empty() {
        namespace.to_string()
    } else {
        let sub_namespace = dir_parts.join("\\");
        format!("{}\\{}", namespace, sub_namespace)
    }
}

pub fn filter_php_dependencies(deps: &HashMap<String, String>) -> HashMap<String, String> {
    deps.iter()
        .filter(|(name, _)| !name.starts_with("php") && !name.starts_with("ext-"))
        .map(|(k, v)| (k.clone(), v.clone()))
        .collect()
}
