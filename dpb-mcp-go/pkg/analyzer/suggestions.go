// Package analyzer - Agent Suggestions for Go implementation
// Provides structured suggestions for AI agents (Cursor, Cline, Claude Code)

package analyzer

import (
	"fmt"
	"strings"
)

// AgentSuggestion represents a structured suggestion for AI agents
type AgentSuggestion struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "warning", "error", "info", "action"
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"` // "critical", "high", "medium", "low"
	Category    string                 `json:"category"` // "security", "license", "outdated", "deprecated"
	Dependency  string                 `json:"dependency,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Actions     []AgentAction          `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AgentAction represents an actionable command
type AgentAction struct {
	ID              string `json:"id"`
	Label           string `json:"label"`
	Command         string `json:"command"`
	Type            string `json:"type"` // "shell", "file-edit", "prompt", "link"
	AutoApply       bool   `json:"autoApply,omitempty"`
	ConfirmRequired bool   `json:"confirmRequired,omitempty"`
	Description     string `json:"description,omitempty"`
}

// SuggestionSummary provides an overview of all suggestions
type SuggestionSummary struct {
	Total      int            `json:"total"`
	BySeverity map[string]int `json:"bySeverity"`
	ByCategory map[string]int `json:"byCategory"`
}

// AgentSuggestionsResponse is the full response for MCP
type AgentSuggestionsResponse struct {
	Summary        SuggestionSummary   `json:"summary"`
	Suggestions    []AgentSuggestion   `json:"suggestions"`
	TerminalOutput string              `json:"terminalOutput"`
}

// GenerateAgentSuggestions creates structured suggestions from compliance issues
func GenerateAgentSuggestions(repoPath string) (*AgentSuggestionsResponse, error) {
	issues, err := CheckCompliance(repoPath)
	if err != nil {
		return nil, err
	}

	history, err := GetDependencyHistory(repoPath)
	if err != nil {
		return nil, err
	}

	var suggestions []AgentSuggestion

	// Convert compliance issues to suggestions
	for _, issue := range issues {
		suggestionID := fmt.Sprintf("dep-%s-%s", issue.Issue, strings.ReplaceAll(issue.Dependency, "/", "-"))

		var actions []AgentAction

		// Add update action for outdated packages
		if issue.Issue == "outdated" {
			actions = append(actions, AgentAction{
				ID:              suggestionID + "-update",
				Label:           "Update to latest",
				Command:         fmt.Sprintf("composer update %s", issue.Dependency),
				Type:            "shell",
				AutoApply:       issue.Severity == "low",
				ConfirmRequired: issue.Severity != "low",
			})
		}

		// Add documentation link
		actions = append(actions, AgentAction{
			ID:      suggestionID + "-docs",
			Label:   "View on Packagist",
			Command: fmt.Sprintf("https://packagist.org/packages/%s", issue.Dependency),
			Type:    "link",
		})

		suggestionType := "warning"
		if issue.Severity == "critical" || issue.Severity == "high" {
			suggestionType = "error"
		}

		suggestions = append(suggestions, AgentSuggestion{
			ID:          suggestionID,
			Type:        suggestionType,
			Title:       fmt.Sprintf("%s Issue: %s", strings.Title(issue.Issue), issue.Dependency),
			Description: issue.Description,
			Severity:    issue.Severity,
			Category:    issue.Issue,
			Dependency:  issue.Dependency,
			Version:     issue.Version,
			Actions:     actions,
			Metadata: map[string]interface{}{
				"recommendation":   issue.Recommendation,
				"autoFixAvailable": issue.AutoFixAvailable,
			},
		})
	}

	// Add suggestions for stale dependencies
	for i, staleDep := range history.Stale {
		if i >= 5 {
			break // Limit to 5 stale suggestions
		}
		suggestions = append(suggestions, AgentSuggestion{
			ID:          fmt.Sprintf("stale-%s", strings.ReplaceAll(staleDep.Name, "/", "-")),
			Type:        "info",
			Title:       fmt.Sprintf("Stale Dependency: %s", staleDep.Name),
			Description: fmt.Sprintf("This dependency hasn't been updated since %s", staleDep.UpdatedAt),
			Severity:    "low",
			Category:    "outdated",
			Dependency:  staleDep.Name,
			Version:     staleDep.Version,
			Actions: []AgentAction{
				{
					ID:      fmt.Sprintf("stale-%s-update", staleDep.Name),
					Label:   "Check for updates",
					Command: fmt.Sprintf("composer outdated %s", staleDep.Name),
					Type:    "shell",
				},
			},
			Metadata: map[string]interface{}{
				"lastUpdated": staleDep.UpdatedAt,
				"addedAt":     staleDep.AddedAt,
			},
		})
	}

	// Add summary suggestion if there are issues
	if len(suggestions) > 0 {
		criticalCount := 0
		highCount := 0
		for _, s := range suggestions {
			if s.Severity == "critical" {
				criticalCount++
			} else if s.Severity == "high" {
				highCount++
			}
		}

		summaryType := "info"
		summarySeverity := "medium"
		if criticalCount > 0 {
			summaryType = "error"
			summarySeverity = "critical"
		} else if highCount > 0 {
			summaryType = "warning"
			summarySeverity = "high"
		}

		summary := AgentSuggestion{
			ID:          "summary",
			Type:        summaryType,
			Title:       "Dependency Analysis Summary",
			Description: fmt.Sprintf("Found %d issues: %d critical, %d high severity", len(suggestions), criticalCount, highCount),
			Severity:    summarySeverity,
			Category:    "security",
			Actions: []AgentAction{
				{
					ID:      "summary-audit",
					Label:   "Run full audit",
					Command: "composer audit",
					Type:    "shell",
				},
				{
					ID:              "summary-update-all",
					Label:           "Update all dependencies",
					Command:         "composer update",
					Type:            "shell",
					ConfirmRequired: true,
				},
			},
			Metadata: map[string]interface{}{
				"totalDependencies": history.CurrentSnapshot.Metadata.TotalCount,
			},
		}
		suggestions = append([]AgentSuggestion{summary}, suggestions...)
	}

	// Calculate summary
	bySeverity := make(map[string]int)
	byCategory := make(map[string]int)
	for _, s := range suggestions {
		bySeverity[s.Severity]++
		byCategory[s.Category]++
	}

	return &AgentSuggestionsResponse{
		Summary: SuggestionSummary{
			Total:      len(suggestions),
			BySeverity: bySeverity,
			ByCategory: byCategory,
		},
		Suggestions:    suggestions,
		TerminalOutput: FormatSuggestionsForTerminal(suggestions),
	}, nil
}

// FormatSuggestionsForTerminal formats suggestions as ASCII terminal output (Claude Code CLI style)
func FormatSuggestionsForTerminal(suggestions []AgentSuggestion) string {
	var sb strings.Builder
	
	// ANSI colors
	red := "\x1b[31m"
	yellow := "\x1b[33m"
	cyan := "\x1b[36m"
	gray := "\x1b[90m"
	dim := "\x1b[2m"
	reset := "\x1b[0m"

	// Count by severity
	counts := map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0}
	for _, s := range suggestions {
		if s.ID != "summary" {
			counts[s.Severity]++
		}
	}

	sb.WriteString("\n")
	sb.WriteString("╭─────────────────────────────────────────────────────────────────╮\n")
	sb.WriteString("│  dependency-buster                                              │\n")
	sb.WriteString("╰─────────────────────────────────────────────────────────────────╯\n")
	sb.WriteString("\n")

	// Summary line
	total := counts["critical"] + counts["high"] + counts["medium"] + counts["low"]
	if total == 0 {
		sb.WriteString("  ✓ No issues found\n\n")
		return sb.String()
	}

	var parts []string
	if counts["critical"] > 0 {
		parts = append(parts, fmt.Sprintf("%d critical", counts["critical"]))
	}
	if counts["high"] > 0 {
		parts = append(parts, fmt.Sprintf("%d high", counts["high"]))
	}
	if counts["medium"] > 0 {
		parts = append(parts, fmt.Sprintf("%d medium", counts["medium"]))
	}
	if counts["low"] > 0 {
		parts = append(parts, fmt.Sprintf("%d low", counts["low"]))
	}

	plural := ""
	if total != 1 {
		plural = "s"
	}
	sb.WriteString(fmt.Sprintf("  Found %d issue%s: %s\n\n", total, plural, strings.Join(parts, ", ")))

	// Group by category
	byCategory := make(map[string][]AgentSuggestion)
	for _, s := range suggestions {
		if s.ID == "summary" {
			continue
		}
		byCategory[s.Category] = append(byCategory[s.Category], s)
	}

	for category, items := range byCategory {
		categoryTitle := strings.Title(category)
		sb.WriteString(fmt.Sprintf("  ▸ %s\n\n", categoryTitle))

		for _, item := range items {
			// Severity indicator
			var indicator, color string
			switch item.Severity {
			case "critical":
				indicator = "●"
				color = red
			case "high":
				indicator = "●"
				color = yellow
			case "medium":
				indicator = "○"
				color = cyan
			default:
				indicator = "·"
				color = gray
			}

			// Package name and version
			if item.Dependency != "" {
				version := item.Version
				if version == "" {
					version = "?"
				}
				sb.WriteString(fmt.Sprintf("    %s%s%s %s%s@%s%s\n", color, indicator, reset, item.Dependency, dim, version, reset))
			} else {
				sb.WriteString(fmt.Sprintf("    %s%s%s %s\n", color, indicator, reset, item.Title))
			}

			// Description (dimmed)
			sb.WriteString(fmt.Sprintf("      %s%s%s\n", dim, item.Description, reset))

			// Quick fix if available
			for _, action := range item.Actions {
				if action.Type == "shell" {
					sb.WriteString(fmt.Sprintf("      %sfix:%s %s\n", dim, reset, action.Command))
					break
				}
			}

			sb.WriteString("\n")
		}
	}

	// Footer with quick commands
	sb.WriteString("  ─────────────────────────────────────────────────────────────\n\n")
	sb.WriteString(fmt.Sprintf("  %sQuick commands:%s\n", dim, reset))
	sb.WriteString("    composer audit          Run security audit\n")
	sb.WriteString("    composer update         Update all dependencies\n\n")

	return sb.String()
}
