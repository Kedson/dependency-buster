use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ComposerJson {
    pub name: Option<String>,
    pub description: Option<String>,
    #[serde(rename = "type")]
    pub package_type: Option<String>,
    pub license: Option<LicenseField>,
    pub require: Option<HashMap<String, String>>,
    #[serde(rename = "require-dev")]
    pub require_dev: Option<HashMap<String, String>>,
    pub autoload: Option<AutoloadConfig>,
    #[serde(rename = "autoload-dev")]
    pub autoload_dev: Option<AutoloadConfig>,
    pub scripts: Option<HashMap<String, serde_json::Value>>,
    pub config: Option<HashMap<String, serde_json::Value>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum LicenseField {
    Single(String),
    Multiple(Vec<String>),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AutoloadConfig {
    #[serde(rename = "psr-4")]
    pub psr4: Option<HashMap<String, Psr4Path>>,
    #[serde(rename = "psr-0")]
    pub psr0: Option<HashMap<String, Psr4Path>>,
    pub files: Option<Vec<String>>,
    pub classmap: Option<Vec<String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Psr4Path {
    Single(String),
    Multiple(Vec<String>),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ComposerLock {
    pub packages: Vec<PackageInfo>,
    #[serde(rename = "packages-dev")]
    pub packages_dev: Option<Vec<PackageInfo>>,
    #[serde(rename = "content-hash")]
    pub content_hash: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PackageInfo {
    pub name: String,
    pub version: String,
    pub description: Option<String>,
    #[serde(rename = "type")]
    pub package_type: Option<String>,
    pub license: Option<Vec<String>>,
    pub authors: Option<Vec<Author>>,
    pub require: Option<HashMap<String, String>>,
    #[serde(rename = "require-dev")]
    pub require_dev: Option<HashMap<String, String>>,
    pub autoload: Option<AutoloadConfig>,
    pub homepage: Option<String>,
    pub source: Option<SourceInfo>,
    pub dist: Option<DistInfo>,
    pub time: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Author {
    pub name: String,
    pub email: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SourceInfo {
    #[serde(rename = "type")]
    pub source_type: String,
    pub url: String,
    pub reference: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DistInfo {
    #[serde(rename = "type")]
    pub dist_type: String,
    pub url: String,
    pub reference: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Psr4Mapping {
    pub namespace: String,
    pub paths: Vec<String>,
    #[serde(rename = "type")]
    pub mapping_type: String,
    #[serde(rename = "isDev")]
    pub is_dev: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Psr4Violation {
    pub file: String,
    #[serde(rename = "expectedNamespace")]
    pub expected_namespace: String,
    #[serde(rename = "actualNamespace")]
    pub actual_namespace: Option<String>,
    pub issue: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DependencyNode {
    pub name: String,
    pub version: String,
    #[serde(rename = "type")]
    pub node_type: String,
    pub dependencies: Vec<String>,
    #[serde(rename = "usedBy")]
    pub used_by: Vec<String>,
    pub license: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SecurityVulnerability {
    pub package: String,
    pub version: String,
    pub severity: String,
    pub cve: Option<String>,
    pub description: String,
    pub recommendation: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NamespaceInfo {
    pub namespace: String,
    pub files: Vec<String>,
    pub classes: Vec<String>,
    pub interfaces: Vec<String>,
    pub traits: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RepoConfig {
    pub name: String,
    pub path: String,
    #[serde(rename = "type")]
    pub repo_type: String,
    pub team: Option<String>,
    pub description: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LicenseDistribution {
    pub license: String,
    pub count: usize,
    pub packages: Vec<String>,
    #[serde(rename = "riskLevel")]
    pub risk_level: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VersionConflict {
    pub package: String,
    pub versions: Vec<RepoVersion>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RepoVersion {
    pub repo: String,
    pub version: String,
}
