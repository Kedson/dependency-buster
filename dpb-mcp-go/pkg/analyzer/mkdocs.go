package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kedson/dpb-mcp/pkg/composer"
	"github.com/kedson/dpb-mcp/pkg/types"
)

// MkDocsOptions contains options for MkDocs generation
type MkDocsOptions struct {
	RepoPath        string
	OutputDir       string
	IncludeChangelog bool
	Format          string // "mkdocs", "html", "markdown"
	SiteName        string
	SiteDescription string
}

// GenerateMkDocsDocs generates MkDocs-compatible documentation structure
func GenerateMkDocsDocs(options MkDocsOptions) (string, error) {
	if options.OutputDir == "" {
		options.OutputDir = filepath.Join(options.RepoPath, "docs")
	}
	if options.Format == "" {
		options.Format = "mkdocs"
	}
	if !options.IncludeChangelog {
		options.IncludeChangelog = true
	}

	// Ensure output directory exists
	if err := os.MkdirAll(options.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Gather all analysis data
	composerJSON, err := composer.ReadComposerJSON(options.RepoPath)
	if err != nil {
		return "", err
	}

	deps, _ := AnalyzeDependencies(options.RepoPath)
	psr4, _ := AnalyzePSR4Autoloading(options.RepoPath)
	namespaces, _ := DetectNamespaces(options.RepoPath)
	security, _ := AuditSecurity(options.RepoPath)
	licenses, _ := AnalyzeLicenses(options.RepoPath)
	depGraph, _ := GenerateDependencyGraph(options.RepoPath, 2, false, "")

	// Get project info
	projectName := options.SiteName
	if projectName == "" {
		projectName = composerJSON.Name
		if projectName == "" {
			projectName = filepath.Base(options.RepoPath)
		}
	}

	projectDesc := options.SiteDescription
	if projectDesc == "" {
		projectDesc = composerJSON.Description
		if projectDesc == "" {
			projectDesc = "Dependency Analysis Documentation"
		}
	}

	// Generate changelog if requested
	var changelogContent string
	if options.IncludeChangelog {
		changelogContent, _ = generateChangelog(options.RepoPath)
	}

	// Generate individual markdown files
	indexContent := generateIndex(projectName, projectDesc, composerJSON, deps, options.IncludeChangelog)
	dependenciesContent := generateDependenciesDoc(deps, depGraph)
	securityContent := generateSecurityDoc(security)
	licensesContent := generateLicensesDoc(licenses)
	architectureContent := generateArchitectureDoc(psr4, namespaces)

	// Write markdown files
	if err := os.WriteFile(filepath.Join(options.OutputDir, "index.md"), []byte(indexContent), 0644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(options.OutputDir, "dependencies.md"), []byte(dependenciesContent), 0644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(options.OutputDir, "security.md"), []byte(securityContent), 0644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(options.OutputDir, "licenses.md"), []byte(licensesContent), 0644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(options.OutputDir, "architecture.md"), []byte(architectureContent), 0644); err != nil {
		return "", err
	}

	if changelogContent != "" {
		if err := os.WriteFile(filepath.Join(options.OutputDir, "changelog.md"), []byte(changelogContent), 0644); err != nil {
			return "", err
		}
	}

	// Generate mkdocs.yml if format is mkdocs
	if options.Format == "mkdocs" {
		mkdocsConfig := generateMkDocsConfig(projectName, projectDesc, options.IncludeChangelog)
		if err := os.WriteFile(filepath.Join(options.OutputDir, "mkdocs.yml"), []byte(mkdocsConfig), 0644); err != nil {
			return "", err
		}
	}

	// Generate HTML if format is html
	if options.Format == "html" {
		htmlContent := generateHTMLSite(projectName, projectDesc, indexContent, dependenciesContent, securityContent, licensesContent, architectureContent, changelogContent)
		if err := os.WriteFile(filepath.Join(options.OutputDir, "index.html"), []byte(htmlContent), 0644); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("Documentation generated successfully in %s", options.OutputDir), nil
}

func generateIndex(projectName, description string, composerJSON *types.ComposerJSON, depsJSON string, includeChangelog bool) string {
	now := time.Now().Format(time.RFC3339)
	
	// Parse dependency JSON to get counts
	var deps DependencyAnalysisResult
	prodCount := 0
	devCount := 0
	if err := json.Unmarshal([]byte(depsJSON), &deps); err == nil {
		prodCount = deps.Stats.TotalProduction
		devCount = deps.Stats.TotalDevelopment
	} else {
		// Fallback: count from composer.json
		if composerJSON.Require != nil {
			prodCount = len(composer.FilterPHPDependencies(composerJSON.Require))
		}
		if composerJSON.RequireDev != nil {
			devCount = len(composerJSON.RequireDev)
		}
	}
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", projectName))
	sb.WriteString(fmt.Sprintf("%s\n\n", description))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", now))
	sb.WriteString("## Quick Overview\n\n")
	
	projectType := "library"
	if composerJSON.Type != "" {
		projectType = composerJSON.Type
	}
	sb.WriteString(fmt.Sprintf("- **Project Type:** %s\n", projectType))
	
	licenseStr := "Not specified"
	if licenses := composer.GetLicenses(composerJSON); len(licenses) > 0 {
		licenseStr = strings.Join(licenses, ", ")
	}
	sb.WriteString(fmt.Sprintf("- **License:** %s\n", licenseStr))
	sb.WriteString(fmt.Sprintf("- **Production Dependencies:** %d\n", prodCount))
	sb.WriteString(fmt.Sprintf("- **Development Dependencies:** %d\n\n", devCount))
	
	sb.WriteString("## Documentation Sections\n\n")
	sb.WriteString("- [Dependencies](./dependencies.md) - Complete dependency analysis and tree\n")
	sb.WriteString("- [Security](./security.md) - Security audit and vulnerability report\n")
	sb.WriteString("- [Licenses](./licenses.md) - License compliance and distribution\n")
	sb.WriteString("- [Architecture](./architecture.md) - Namespace structure and PSR-4 compliance\n")
	if includeChangelog {
		sb.WriteString("- [Changelog](./changelog.md) - Dependency change history\n")
	}
	sb.WriteString("\n## Getting Started\n\n")
	sb.WriteString("This documentation was automatically generated by dependency-buster MCP.\n\n")
	sb.WriteString("To view with MkDocs:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("cd docs\n")
	sb.WriteString("mkdocs serve\n")
	sb.WriteString("```\n")
	
	return sb.String()
}

func generateDependenciesDoc(depsJSON string, graph string) string {
	var sb strings.Builder
	sb.WriteString("# Dependencies\n\n")
	
	// Parse dependency JSON
	var deps DependencyAnalysisResult
	if err := json.Unmarshal([]byte(depsJSON), &deps); err == nil {
		sb.WriteString("## Summary\n\n")
		sb.WriteString(fmt.Sprintf("- **Production:** %d packages\n", deps.Stats.TotalProduction))
		sb.WriteString(fmt.Sprintf("- **Development:** %d packages\n", deps.Stats.TotalDevelopment))
		sb.WriteString(fmt.Sprintf("- **Total:** %d packages\n\n", deps.Stats.TotalProduction+deps.Stats.TotalDevelopment))
		
		// Show sample production dependencies
		if len(deps.Production) > 0 {
			sb.WriteString("## Production Dependencies\n\n")
			sb.WriteString("| Package | Version |\n")
			sb.WriteString("|---------|----------|\n")
			count := 0
			for name, version := range deps.Production {
				if count >= 50 {
					break
				}
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", name, version))
				count++
			}
			if len(deps.Production) > 50 {
				sb.WriteString(fmt.Sprintf("\n*... and %d more*\n\n", len(deps.Production)-50))
			} else {
				sb.WriteString("\n")
			}
		}
		
		// Show sample development dependencies
		if len(deps.Development) > 0 {
			sb.WriteString("## Development Dependencies\n\n")
			sb.WriteString("| Package | Version |\n")
			sb.WriteString("|---------|----------|\n")
			count := 0
			for name, version := range deps.Development {
				if count >= 50 {
					break
				}
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", name, version))
				count++
			}
			if len(deps.Development) > 50 {
				sb.WriteString(fmt.Sprintf("\n*... and %d more*\n\n", len(deps.Development)-50))
			} else {
				sb.WriteString("\n")
			}
		}
	} else {
		sb.WriteString("## Summary\n\n")
		sb.WriteString("See full dependency analysis below.\n\n")
	}
	
	sb.WriteString("## Dependency Graph\n\n")
	sb.WriteString("```mermaid\n")
	sb.WriteString(graph)
	sb.WriteString("\n```\n\n")
	sb.WriteString("*For detailed dependency information, use the `analyze_dependencies` tool.*\n")
	return sb.String()
}

func generateSecurityDoc(securityJSON string) string {
	var sb strings.Builder
	sb.WriteString("# Security Audit\n\n")
	
	// Parse security JSON
	var security SecurityAuditResult
	if err := json.Unmarshal([]byte(securityJSON), &security); err == nil {
		sb.WriteString(fmt.Sprintf("## Risk Level: %s\n\n", strings.ToUpper(security.RiskLevel)))
		sb.WriteString("## Summary\n\n")
		sb.WriteString(fmt.Sprintf("- **Critical:** %d\n", security.Summary.Critical))
		sb.WriteString(fmt.Sprintf("- **High:** %d\n", security.Summary.High))
		sb.WriteString(fmt.Sprintf("- **Medium:** %d\n", security.Summary.Medium))
		sb.WriteString(fmt.Sprintf("- **Low:** %d\n", security.Summary.Low))
		sb.WriteString(fmt.Sprintf("- **Total Issues:** %d\n\n", len(security.Vulnerabilities)))
		
		if len(security.Vulnerabilities) > 0 {
			sb.WriteString("## Vulnerabilities\n\n")
			sb.WriteString("| Package | Version | Severity | Description |\n")
			sb.WriteString("|---------|---------|----------|-------------|\n")
			for i, vuln := range security.Vulnerabilities {
				if i >= 100 {
					break
				}
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s |\n", 
					vuln.Package, vuln.Version, vuln.Severity, vuln.Description))
			}
			if len(security.Vulnerabilities) > 100 {
				sb.WriteString(fmt.Sprintf("\n*... and %d more vulnerabilities*\n", len(security.Vulnerabilities)-100))
			}
		} else {
			sb.WriteString("## Status\n\n✅ No known vulnerabilities found.\n")
		}
	} else {
		sb.WriteString("*For detailed security information, use the `audit_security` tool.*\n")
	}
	return sb.String()
}

func generateLicensesDoc(licensesJSON string) string {
	var sb strings.Builder
	sb.WriteString("# License Compliance\n\n")
	
	// Parse licenses JSON
	var licenses LicenseAnalysisResult
	if err := json.Unmarshal([]byte(licensesJSON), &licenses); err == nil {
		sb.WriteString("## Summary\n\n")
		sb.WriteString(fmt.Sprintf("- **Total Packages:** %d\n", licenses.Summary.TotalPackages))
		sb.WriteString(fmt.Sprintf("- **Unique Licenses:** %d\n", licenses.Summary.UniqueLicenses))
		sb.WriteString(fmt.Sprintf("- **Unknown Licenses:** %d\n\n", licenses.Summary.UnknownLicenses))
		
		if len(licenses.Distribution) > 0 {
			sb.WriteString("## License Distribution\n\n")
			sb.WriteString("| License | Count | Percentage |\n")
			sb.WriteString("|---------|-------|------------|\n")
			total := licenses.Summary.TotalPackages
			for _, dist := range licenses.Distribution {
				percentage := 0.0
				if total > 0 {
					percentage = float64(dist.Count) / float64(total) * 100
				}
				sb.WriteString(fmt.Sprintf("| %s | %d | %.1f%% |\n", dist.License, dist.Count, percentage))
			}
			sb.WriteString("\n")
		}
		
		if len(licenses.CompatibilityIssues) > 0 {
			sb.WriteString("## Compatibility Issues\n\n")
			for _, issue := range licenses.CompatibilityIssues {
				sb.WriteString(fmt.Sprintf("- %s\n", issue))
			}
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("*For detailed license information, use the `analyze_licenses` tool.*\n")
	}
	return sb.String()
}

func generateArchitectureDoc(psr4JSON string, namespacesJSON string) string {
	var sb strings.Builder
	sb.WriteString("# Architecture\n\n")
	
	// Parse PSR-4 JSON
	var psr4 PSR4AnalysisResult
	if err := json.Unmarshal([]byte(psr4JSON), &psr4); err == nil {
		sb.WriteString("## PSR-4 Autoloading\n\n")
		sb.WriteString("### Summary\n\n")
		sb.WriteString(fmt.Sprintf("- **Total Mappings:** %d\n", psr4.Stats.TotalMappings))
		sb.WriteString(fmt.Sprintf("- **Files Analyzed:** %d\n", psr4.Stats.TotalFiles))
		sb.WriteString(fmt.Sprintf("- **PSR-4 Compliant:** %d\n", psr4.Stats.ValidFiles))
		sb.WriteString(fmt.Sprintf("- **Violations:** %d\n\n", psr4.Stats.ViolationCount))
		
		if len(psr4.Mappings) > 0 {
			sb.WriteString("### Mappings\n\n")
			sb.WriteString("| Namespace Prefix | Directory |\n")
			sb.WriteString("|------------------|-----------|\n")
			for i, mapping := range psr4.Mappings {
				if i >= 20 {
					break
				}
				dirs := strings.Join(mapping.Paths, ", ")
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", mapping.Namespace, dirs))
			}
			if len(psr4.Mappings) > 20 {
				sb.WriteString(fmt.Sprintf("\n*... and %d more mappings*\n", len(psr4.Mappings)-20))
			}
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString("## PSR-4 Autoloading\n\n")
		sb.WriteString("*For detailed PSR-4 information, use the `analyze_psr4` tool.*\n\n")
	}
	
	// Parse namespaces JSON
	var namespaces NamespaceDetectionResult
	if err := json.Unmarshal([]byte(namespacesJSON), &namespaces); err == nil {
		sb.WriteString("## Namespaces\n\n")
		if len(namespaces.Namespaces) > 0 {
			sb.WriteString(fmt.Sprintf("Found **%d** namespaces:\n\n", len(namespaces.Namespaces)))
			for i, ns := range namespaces.Namespaces {
				if i >= 30 {
					break
				}
				sb.WriteString(fmt.Sprintf("- `%s` (%d files)\n", ns.Namespace, len(ns.Files)))
			}
			if len(namespaces.Namespaces) > 30 {
				sb.WriteString(fmt.Sprintf("\n*... and %d more namespaces*\n", len(namespaces.Namespaces)-30))
			}
		} else {
			sb.WriteString("*No namespaces detected.*\n")
		}
	} else {
		sb.WriteString("## Namespaces\n\n")
		sb.WriteString("*For detailed namespace information, use the `detect_namespaces` tool.*\n")
	}
	return sb.String()
}

func generateChangelog(repoPath string) (string, error) {
	currentSnapshot, err := CreateDependencySnapshot(repoPath)
	if err != nil {
		return "", err
	}

	oldSnapshot, err := LoadTracker(repoPath)
	if err != nil {
		// No previous snapshot - create initial changelog
		now := time.Now().Format("2006-01-02")
		return fmt.Sprintf(`# Dependency Changelog

## %s

Initial snapshot created.

**Total Dependencies:** %d
`, now, currentSnapshot.Metadata.TotalCount), nil
	}

	changes := CompareSnapshots(oldSnapshot, currentSnapshot)
	if len(changes) == 0 {
		now := time.Now().Format("2006-01-02")
		return fmt.Sprintf(`# Dependency Changelog

## %s

No changes detected since last snapshot.

**Total Dependencies:** %d
`, now, currentSnapshot.Metadata.TotalCount), nil
	}

	var sb strings.Builder
	now := time.Now().Format("2006-01-02")
	sb.WriteString(fmt.Sprintf("# Dependency Changelog\n\n## %s\n\n", now))
	sb.WriteString("### Summary\n\n")

	added := 0
	updated := 0
	removed := 0
	for _, change := range changes {
		switch change.Type {
		case "added":
			added++
		case "updated":
			updated++
		case "removed":
			removed++
		}
	}

	sb.WriteString(fmt.Sprintf("- **Added:** %d\n", added))
	sb.WriteString(fmt.Sprintf("- **Updated:** %d\n", updated))
	sb.WriteString(fmt.Sprintf("- **Removed:** %d\n\n", removed))

	if added > 0 {
		sb.WriteString("### Added\n\n")
		for _, change := range changes {
			if change.Type == "added" {
				sb.WriteString(fmt.Sprintf("- `%s` `%s`\n", change.Name, change.NewVersion))
			}
		}
		sb.WriteString("\n")
	}

	if updated > 0 {
		sb.WriteString("### Updated\n\n")
		for _, change := range changes {
			if change.Type == "updated" {
				sb.WriteString(fmt.Sprintf("- `%s`: `%s` → `%s`\n", change.Name, change.OldVersion, change.NewVersion))
			}
		}
		sb.WriteString("\n")
	}

	if removed > 0 {
		sb.WriteString("### Removed\n\n")
		for _, change := range changes {
			if change.Type == "removed" {
				sb.WriteString(fmt.Sprintf("- `%s` `%s`\n", change.Name, change.OldVersion))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func generateMkDocsConfig(siteName, siteDescription string, includeChangelog bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("site_name: %s\n", siteName))
	sb.WriteString(fmt.Sprintf("site_description: %s\n", siteDescription))
	sb.WriteString("site_url: https://example.com\n\n")
	sb.WriteString("theme:\n")
	sb.WriteString("  name: material\n")
	sb.WriteString("  palette:\n")
	sb.WriteString("    primary: blue\n")
	sb.WriteString("    accent: blue\n\n")
	sb.WriteString("markdown_extensions:\n")
	sb.WriteString("  - pymdownx.highlight:\n")
	sb.WriteString("      anchor_linenums: true\n")
	sb.WriteString("  - pymdownx.inlinehilite\n")
	sb.WriteString("  - pymdownx.snippets\n")
	sb.WriteString("  - pymdownx.superfences:\n")
	sb.WriteString("      custom_fences:\n")
	sb.WriteString("        - name: mermaid\n")
	sb.WriteString("          class: mermaid\n")
	sb.WriteString("          format: !!python/name:pymdownx.superfences.fence_code_format\n\n")
	sb.WriteString("nav:\n")
	sb.WriteString("  - Home: index.md\n")
	sb.WriteString("  - Dependencies: dependencies.md\n")
	sb.WriteString("  - Security: security.md\n")
	sb.WriteString("  - Licenses: licenses.md\n")
	sb.WriteString("  - Architecture: architecture.md\n")
	if includeChangelog {
		sb.WriteString("  - Changelog: changelog.md\n")
	}

	return sb.String()
}

func generateHTMLSite(siteName, siteDescription, index, dependencies, security, licenses, architecture, changelog string) string {
	// Escape markdown content for JavaScript template literals (backticks)
	escapeJS := func(s string) string {
		// Escape backslashes first (must be first!)
		s = strings.ReplaceAll(s, "\\", "\\\\")
		// Escape backticks for template literals
		s = strings.ReplaceAll(s, "`", "\\`")
		// Escape dollar signs (for template literal expressions)
		s = strings.ReplaceAll(s, "${", "\\${")
		// Escape newlines
		s = strings.ReplaceAll(s, "\n", "\\n")
		// Escape carriage returns
		s = strings.ReplaceAll(s, "\r", "\\r")
		return s
	}
	
	indexEscaped := escapeJS(index)
	depsEscaped := escapeJS(dependencies)
	secEscaped := escapeJS(security)
	licEscaped := escapeJS(licenses)
	archEscaped := escapeJS(architecture)
	
	changelogNav := ""
	changelogSection := ""
	changelogScript := ""
	if changelog != "" {
		changelogNav = "\n    <a href=\"#changelog\">Changelog</a>"
		changelogSection = `
  <div id="changelog" class="section">
    <h2>Changelog</h2>
    <div id="changelog-content"></div>
  </div>`
		changelogEscaped := escapeJS(changelog)
		backtick := "`"
		changelogScript = fmt.Sprintf(`
    const changelogMD = `+backtick+`%s`+backtick+`;
    if (document.getElementById('changelog-content')) {
      document.getElementById('changelog-content').innerHTML = markdownToHTML(changelogMD);
    }`, changelogEscaped)
	}
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; line-height: 1.6; }
    nav { background: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
    nav a { margin-right: 20px; text-decoration: none; color: #0066cc; font-weight: 500; }
    nav a:hover { text-decoration: underline; }
    h1 { color: #333; border-bottom: 2px solid #0066cc; padding-bottom: 10px; }
    h2 { color: #555; margin-top: 30px; border-bottom: 1px solid #ddd; padding-bottom: 5px; }
    h3 { color: #666; margin-top: 20px; }
    code { background: #f5f5f5; padding: 2px 6px; border-radius: 3px; font-family: 'Courier New', monospace; }
    pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; border-left: 3px solid #0066cc; }
    table { border-collapse: collapse; width: 100%%; margin: 20px 0; }
    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
    th { background: #f5f5f5; font-weight: 600; }
    tr:nth-child(even) { background: #fafafa; }
    a { color: #0066cc; }
    .section { margin-bottom: 40px; }
    .meta { color: #666; font-size: 0.9em; margin-bottom: 20px; }
  </style>
  <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
</head>
<body>
  <nav>
    <a href="#index">Home</a>
    <a href="#dependencies">Dependencies</a>
    <a href="#security">Security</a>
    <a href="#licenses">Licenses</a>
    <a href="#architecture">Architecture</a>%s
  </nav>
  
  <div id="index" class="section">
    <h1>%s</h1>
    <p class="meta">%s</p>
    <div id="index-content"></div>
  </div>
  
  <div id="dependencies" class="section">
    <h2>Dependencies</h2>
    <div id="dependencies-content"></div>
  </div>
  
  <div id="security" class="section">
    <h2>Security</h2>
    <div id="security-content"></div>
  </div>
  
  <div id="licenses" class="section">
    <h2>Licenses</h2>
    <div id="licenses-content"></div>
  </div>
  
  <div id="architecture" class="section">
    <h2>Architecture</h2>
    <div id="architecture-content"></div>
  </div>%s
  
  <script>
    function markdownToHTML(md) {
      if (typeof marked !== 'undefined') {
        return marked.parse(md);
      }
      return md
        .replace(/^# (.*$)/gim, '<h1>$1</h1>')
        .replace(/^## (.*$)/gim, '<h2>$1</h2>')
        .replace(/^### (.*$)/gim, '<h3>$1</h3>')
        .replace(/\\*\\*(.*?)\\*\\*/gim, '<strong>$1</strong>')
        .replace(/\\*(.*?)\\*/gim, '<em>$1</em>')
        .replace(/\x60([^\x60]+)\x60/gim, '<code>$1</code>')
        .replace(/\\n/gim, '<br>');
    }
    
    const indexMD = ` + "`" + `%s` + "`" + `;
    const depsMD = ` + "`" + `%s` + "`" + `;
    const secMD = ` + "`" + `%s` + "`" + `;
    const licMD = ` + "`" + `%s` + "`" + `;
    const archMD = ` + "`" + `%s` + "`" + `;%s
    
    // Wait for DOM to be ready
    function renderContent() {
      document.getElementById('index-content').innerHTML = markdownToHTML(indexMD);
      document.getElementById('dependencies-content').innerHTML = markdownToHTML(depsMD);
      document.getElementById('security-content').innerHTML = markdownToHTML(secMD);
      document.getElementById('licenses-content').innerHTML = markdownToHTML(licMD);
      document.getElementById('architecture-content').innerHTML = markdownToHTML(archMD);%s
    }
    
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', renderContent);
    } else {
      renderContent();
    }
  </script>
</body>
</html>`, siteName, changelogNav, siteName, siteDescription, changelogSection, indexEscaped, depsEscaped, secEscaped, licEscaped, archEscaped, changelogScript, changelogScript)
	
	return html
}
