//! Agent Suggestion Hooks - Integration with IDE agents (Cursor, Cline, Claude Code)
//! Provides structured suggestions for non-compliant dependencies

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

use super::tracker::{check_compliance, get_dependency_history, ComplianceIssue, TrackedDependency};

/// A structured suggestion for AI agents
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentSuggestion {
    pub id: String,
    #[serde(rename = "type")]
    pub suggestion_type: String, // "warning", "error", "info", "action"
    pub title: String,
    pub description: String,
    pub severity: String, // "critical", "high", "medium", "low"
    pub category: String, // "security", "license", "outdated", "deprecated"
    #[serde(skip_serializing_if = "Option::is_none")]
    pub dependency: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub version: Option<String>,
    pub actions: Vec<AgentAction>,
    pub metadata: HashMap<String, serde_json::Value>,
}

/// An actionable command for agents
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentAction {
    pub id: String,
    pub label: String,
    pub command: String,
    #[serde(rename = "type")]
    pub action_type: String, // "shell", "file-edit", "prompt", "link"
    #[serde(skip_serializing_if = "Option::is_none")]
    pub auto_apply: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub confirm_required: Option<bool>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
}

/// Summary of all suggestions
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SuggestionSummary {
    pub total: usize,
    pub by_severity: HashMap<String, usize>,
    pub by_category: HashMap<String, usize>,
}

/// Full response for MCP
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentSuggestionsResponse {
    pub summary: SuggestionSummary,
    pub suggestions: Vec<AgentSuggestion>,
    pub terminal_output: String,
}

/// Generate structured suggestions for AI agents
pub fn generate_agent_suggestions(repo_path: &str) -> Result<AgentSuggestionsResponse> {
    let issues = check_compliance(repo_path)?;
    let history = get_dependency_history(repo_path)?;
    
    let mut suggestions: Vec<AgentSuggestion> = Vec::new();
    
    // Convert compliance issues to suggestions
    for issue in &issues {
        let suggestion_id = format!("dep-{}-{}", issue.issue, issue.dependency.replace("/", "-"));
        
        let mut actions = Vec::new();
        
        // Add update action for outdated packages
        if issue.issue == "outdated" {
            actions.push(AgentAction {
                id: format!("{}-update", suggestion_id),
                label: "Update to latest".to_string(),
                command: format!("composer update {}", issue.dependency),
                action_type: "shell".to_string(),
                auto_apply: Some(issue.severity == "low"),
                confirm_required: Some(issue.severity != "low"),
                description: None,
            });
        }
        
        // Add documentation link
        actions.push(AgentAction {
            id: format!("{}-docs", suggestion_id),
            label: "View on Packagist".to_string(),
            command: format!("https://packagist.org/packages/{}", issue.dependency),
            action_type: "link".to_string(),
            auto_apply: None,
            confirm_required: None,
            description: None,
        });
        
        let suggestion_type = if issue.severity == "critical" || issue.severity == "high" {
            "error"
        } else {
            "warning"
        };
        
        let mut metadata = HashMap::new();
        metadata.insert("recommendation".to_string(), serde_json::json!(issue.recommendation));
        metadata.insert("autoFixAvailable".to_string(), serde_json::json!(issue.auto_fix_available));
        
        suggestions.push(AgentSuggestion {
            id: suggestion_id,
            suggestion_type: suggestion_type.to_string(),
            title: format!("{} Issue: {}", capitalize(&issue.issue), issue.dependency),
            description: issue.description.clone(),
            severity: issue.severity.clone(),
            category: issue.issue.clone(),
            dependency: Some(issue.dependency.clone()),
            version: Some(issue.version.clone()),
            actions,
            metadata,
        });
    }
    
    // Add suggestions for stale dependencies (limit to 5)
    for stale_dep in history.stale.iter().take(5) {
        let mut metadata = HashMap::new();
        if let Some(ref updated_at) = stale_dep.updated_at {
            metadata.insert("lastUpdated".to_string(), serde_json::json!(updated_at));
        }
        if let Some(ref added_at) = stale_dep.added_at {
            metadata.insert("addedAt".to_string(), serde_json::json!(added_at));
        }
        
        suggestions.push(AgentSuggestion {
            id: format!("stale-{}", stale_dep.name.replace("/", "-")),
            suggestion_type: "info".to_string(),
            title: format!("Stale Dependency: {}", stale_dep.name),
            description: format!("This dependency hasn't been updated since {}", 
                stale_dep.updated_at.as_deref().unwrap_or("unknown")),
            severity: "low".to_string(),
            category: "outdated".to_string(),
            dependency: Some(stale_dep.name.clone()),
            version: Some(stale_dep.version.clone()),
            actions: vec![AgentAction {
                id: format!("stale-{}-update", stale_dep.name),
                label: "Check for updates".to_string(),
                command: format!("composer outdated {}", stale_dep.name),
                action_type: "shell".to_string(),
                auto_apply: None,
                confirm_required: None,
                description: None,
            }],
            metadata,
        });
    }
    
    // Add summary suggestion if there are issues
    if !suggestions.is_empty() {
        let critical_count = suggestions.iter().filter(|s| s.severity == "critical").count();
        let high_count = suggestions.iter().filter(|s| s.severity == "high").count();
        
        let (summary_type, summary_severity) = if critical_count > 0 {
            ("error", "critical")
        } else if high_count > 0 {
            ("warning", "high")
        } else {
            ("info", "medium")
        };
        
        let mut metadata = HashMap::new();
        metadata.insert("totalDependencies".to_string(), 
            serde_json::json!(history.current_snapshot.metadata.total_count));
        
        let summary = AgentSuggestion {
            id: "summary".to_string(),
            suggestion_type: summary_type.to_string(),
            title: "Dependency Analysis Summary".to_string(),
            description: format!("Found {} issues: {} critical, {} high severity", 
                suggestions.len(), critical_count, high_count),
            severity: summary_severity.to_string(),
            category: "security".to_string(),
            dependency: None,
            version: None,
            actions: vec![
                AgentAction {
                    id: "summary-audit".to_string(),
                    label: "Run full audit".to_string(),
                    command: "composer audit".to_string(),
                    action_type: "shell".to_string(),
                    auto_apply: None,
                    confirm_required: None,
                    description: None,
                },
                AgentAction {
                    id: "summary-update-all".to_string(),
                    label: "Update all dependencies".to_string(),
                    command: "composer update".to_string(),
                    action_type: "shell".to_string(),
                    auto_apply: None,
                    confirm_required: Some(true),
                    description: None,
                },
            ],
            metadata,
        };
        suggestions.insert(0, summary);
    }
    
    // Calculate summary
    let mut by_severity: HashMap<String, usize> = HashMap::new();
    let mut by_category: HashMap<String, usize> = HashMap::new();
    
    for s in &suggestions {
        *by_severity.entry(s.severity.clone()).or_insert(0) += 1;
        *by_category.entry(s.category.clone()).or_insert(0) += 1;
    }
    
    Ok(AgentSuggestionsResponse {
        summary: SuggestionSummary {
            total: suggestions.len(),
            by_severity,
            by_category,
        },
        suggestions: suggestions.clone(),
        terminal_output: format_suggestions_for_terminal(&suggestions),
    })
}

/// Format suggestions as ASCII terminal output
pub fn format_suggestions_for_terminal(suggestions: &[AgentSuggestion]) -> String {
    let mut output = String::new();
    
    output.push_str("\n");
    output.push_str("╔══════════════════════════════════════════════════════════════════╗\n");
    output.push_str("║  dependency-buster // Agent Suggestions                          ║\n");
    output.push_str("╚══════════════════════════════════════════════════════════════════╝\n");
    output.push_str("\n");
    
    for suggestion in suggestions {
        let icon = match suggestion.severity.as_str() {
            "critical" => "🔴",
            "high" => "🟠",
            "medium" => "🟡",
            "low" => "🟢",
            _ => "○",
        };
        
        let type_icon = match suggestion.suggestion_type.as_str() {
            "error" => "✗",
            "warning" => "⚠",
            "info" => "ℹ",
            "action" => "→",
            _ => "•",
        };
        
        output.push_str(&format!("{} {} {}\n", icon, type_icon, suggestion.title));
        output.push_str(&format!("   {}\n", suggestion.description));
        
        if let Some(ref dep) = suggestion.dependency {
            output.push_str(&format!("   Package: {}@{}\n", 
                dep, suggestion.version.as_deref().unwrap_or("unknown")));
        }
        
        if !suggestion.actions.is_empty() {
            output.push_str("   Actions:\n");
            for action in &suggestion.actions {
                match action.action_type.as_str() {
                    "shell" => output.push_str(&format!("     $ {}\n", action.command)),
                    "link" => output.push_str(&format!("     🔗 {}\n", action.command)),
                    _ => {}
                }
            }
        }
        
        output.push_str("\n");
    }
    
    output
}

fn capitalize(s: &str) -> String {
    let mut chars = s.chars();
    match chars.next() {
        None => String::new(),
        Some(first) => first.to_uppercase().collect::<String>() + chars.as_str(),
    }
}
