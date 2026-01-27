use anyhow::Result;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::path::Path;

use crate::composer::read_composer_lock;
use crate::types::{LicenseDistribution, SecurityVulnerability};

#[derive(Debug, Serialize, Deserialize)]
pub struct SecurityAuditResult {
    pub vulnerabilities: Vec<SecurityVulnerability>,
    #[serde(rename = "riskLevel")]
    pub risk_level: String,
    pub summary: SecuritySummary,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SecuritySummary {
    pub critical: usize,
    pub high: usize,
    pub medium: usize,
    pub low: usize,
}

pub fn audit_security<P: AsRef<Path>>(repo_path: P) -> Result<String> {
    let lock = read_composer_lock(&repo_path)?;

    let mut vulnerabilities = Vec::new();

    let mut all_packages = lock.packages.clone();
    if let Some(dev_packages) = &lock.packages_dev {
        all_packages.extend(dev_packages.clone());
    }

    for pkg in &all_packages {
        // Check for dev versions
        if pkg.version.contains("dev") && !pkg.version.starts_with("dev-") {
            vulnerabilities.push(SecurityVulnerability {
                package: pkg.name.clone(),
                version: pkg.version.clone(),
                severity: "medium".to_string(),
                cve: None,
                description: "Using development version in production".to_string(),
                recommendation: "Pin to a stable release version".to_string(),
            });
        }

        // Check for pre-1.0 versions
        if pkg.version.starts_with("0.") {
            vulnerabilities.push(SecurityVulnerability {
                package: pkg.name.clone(),
                version: pkg.version.clone(),
                severity: "low".to_string(),
                cve: None,
                description: "Using pre-1.0 version (potentially unstable)".to_string(),
                recommendation: "Consider upgrading to a stable 1.x+ version if available"
                    .to_string(),
            });
        }

        // Check for very old packages (5+ years)
        if let Some(time_str) = &pkg.time {
            if let Ok(pkg_time) = time_str.parse::<DateTime<Utc>>() {
                let five_years_ago = Utc::now() - chrono::Duration::days(5 * 365);
                if pkg_time < five_years_ago {
                    vulnerabilities.push(SecurityVulnerability {
                        package: pkg.name.clone(),
                        version: pkg.version.clone(),
                        severity: "medium".to_string(),
                        cve: None,
                        description: "Package has not been updated in over 5 years".to_string(),
                        recommendation: "Check for maintained alternatives or security advisories"
                            .to_string(),
                    });
                }
            }
        }
    }

    let mut summary = SecuritySummary {
        critical: 0,
        high: 0,
        medium: 0,
        low: 0,
    };

    let mut risk_level = "low";

    for vuln in &vulnerabilities {
        match vuln.severity.as_str() {
            "critical" => {
                summary.critical += 1;
                risk_level = "critical";
            }
            "high" => {
                summary.high += 1;
                if risk_level != "critical" {
                    risk_level = "high";
                }
            }
            "medium" => {
                summary.medium += 1;
                if risk_level != "critical" && risk_level != "high" {
                    risk_level = "medium";
                }
            }
            "low" => summary.low += 1,
            _ => {}
        }
    }

    let result = SecurityAuditResult {
        vulnerabilities,
        risk_level: risk_level.to_string(),
        summary,
    };

    Ok(serde_json::to_string_pretty(&result)?)
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LicenseAnalysisResult {
    pub distribution: Vec<LicenseDistribution>,
    #[serde(rename = "compatibilityIssues")]
    pub compatibility_issues: Vec<String>,
    pub summary: LicenseSummary,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LicenseSummary {
    #[serde(rename = "totalPackages")]
    pub total_packages: usize,
    #[serde(rename = "uniqueLicenses")]
    pub unique_licenses: usize,
    #[serde(rename = "unknownLicenses")]
    pub unknown_licenses: usize,
}

pub fn analyze_licenses<P: AsRef<Path>>(repo_path: P) -> Result<String> {
    let lock = read_composer_lock(&repo_path)?;

    let mut license_map: HashMap<String, Vec<String>> = HashMap::new();
    let mut unknown_count = 0;

    let mut all_packages = lock.packages.clone();
    if let Some(dev_packages) = &lock.packages_dev {
        all_packages.extend(dev_packages.clone());
    }

    for pkg in &all_packages {
        let licenses = pkg.license.clone().unwrap_or_else(|| vec!["Unknown".to_string()]);

        for license in licenses {
            if license == "Unknown" {
                unknown_count += 1;
            }
            license_map
                .entry(license)
                .or_insert_with(Vec::new)
                .push(pkg.name.clone());
        }
    }

    let unique_license_count = license_map.len();
    
    let distribution: Vec<LicenseDistribution> = license_map
        .into_iter()
        .map(|(license, packages)| LicenseDistribution {
            risk_level: assess_license_risk(&license),
            count: packages.len(),
            license,
            packages,
        })
        .collect();

    // Check for compatibility issues
    let mut compatibility_issues = Vec::new();
    let has_gpl = distribution.iter().any(|d| d.license.contains("GPL"));
    let has_proprietary = distribution.iter().any(|d| d.license.contains("Proprietary"));

    if has_gpl && has_proprietary {
        compatibility_issues.push(
            "Potential conflict: GPL and Proprietary licenses detected. Review compatibility."
                .to_string(),
        );
    }

    let result = LicenseAnalysisResult {
        distribution,
        compatibility_issues,
        summary: LicenseSummary {
            total_packages: all_packages.len(),
            unique_licenses: unique_license_count,
            unknown_licenses: unknown_count,
        },
    };

    Ok(serde_json::to_string_pretty(&result)?)
}

fn assess_license_risk(license: &str) -> String {
    let safe_licenses = ["MIT", "Apache-2.0", "BSD-3-Clause", "BSD-2-Clause", "ISC"];

    if safe_licenses.contains(&license) {
        return "safe".to_string();
    }

    let caution_licenses = ["LGPL", "MPL", "EPL"];
    for caution in &caution_licenses {
        if license.contains(caution) {
            return "caution".to_string();
        }
    }

    if license.contains("GPL") || license == "Unknown" || license.contains("Proprietary") {
        return "review-required".to_string();
    }

    "caution".to_string()
}
