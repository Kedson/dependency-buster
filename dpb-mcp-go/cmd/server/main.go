package main

import (
	"log"

	"github.com/kedson/dpb-mcp/pkg/analyzer"
	"github.com/kedson/dpb-mcp/pkg/mcp"
)

func main() {
	server := mcp.NewServer("php-dependency-analyzer", "1.0.0")

	// Register all tools
	registerTools(server)

	// Start the server
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func registerTools(server *mcp.Server) {
	// Tool 1: Analyze Dependencies
	server.RegisterTool(mcp.Tool{
		Name:        "analyze_dependencies",
		Description: "Comprehensive dependency analysis including production, dev, and dependency tree",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to PHP repository",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		return analyzer.AnalyzeDependencies(repoPath)
	})

	// Tool 2: Analyze PSR-4
	server.RegisterTool(mcp.Tool{
		Name:        "analyze_psr4",
		Description: "Analyze PSR-4 autoloading configuration and validate namespace compliance",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to PHP repository",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		return analyzer.AnalyzePSR4Autoloading(repoPath)
	})

	// Tool 3: Detect Namespaces
	server.RegisterTool(mcp.Tool{
		Name:        "detect_namespaces",
		Description: "Detect all namespaces used in the codebase",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to PHP repository",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		return analyzer.DetectNamespaces(repoPath)
	})

	// Tool 4: Analyze Namespace Usage
	server.RegisterTool(mcp.Tool{
		Name:        "analyze_namespace_usage",
		Description: "Analyze usage of a specific namespace across the codebase",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to PHP repository",
				},
				"namespace": {
					Type:        "string",
					Description: "Target namespace to analyze",
				},
			},
			Required: []string{"repo_path", "namespace"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		namespace := args["namespace"].(string)
		return analyzer.AnalyzeNamespaceUsage(repoPath, namespace)
	})

	// Tool 5: Generate Dependency Graph
	server.RegisterTool(mcp.Tool{
		Name:        "generate_dependency_graph",
		Description: "Generate Mermaid diagram of dependency relationships",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to PHP repository",
				},
				"max_depth": {
					Type:        "number",
					Description: "Maximum depth for dependency tree (default: 2)",
				},
				"include_dev": {
					Type:        "boolean",
					Description: "Include development dependencies",
				},
				"focus_package": {
					Type:        "string",
					Description: "Focus on specific package and its dependencies",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		maxDepth := 2
		if md, ok := args["max_depth"].(float64); ok {
			maxDepth = int(md)
		}
		includeDev := false
		if id, ok := args["include_dev"].(bool); ok {
			includeDev = id
		}
		focusPackage := ""
		if fp, ok := args["focus_package"].(string); ok {
			focusPackage = fp
		}
		return analyzer.GenerateDependencyGraph(repoPath, maxDepth, includeDev, focusPackage)
	})

	// Tool 6: Audit Security
	server.RegisterTool(mcp.Tool{
		Name:        "audit_security",
		Description: "Audit dependencies for security vulnerabilities and outdated packages",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to PHP repository",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		return analyzer.AuditSecurity(repoPath)
	})

	// Tool 7: Analyze Licenses
	server.RegisterTool(mcp.Tool{
		Name:        "analyze_licenses",
		Description: "Analyze license distribution and compatibility across dependencies",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to PHP repository",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		return analyzer.AnalyzeLicenses(repoPath)
	})

	// Tool 8: Find Circular Dependencies
	server.RegisterTool(mcp.Tool{
		Name:        "find_circular_dependencies",
		Description: "Find circular dependency chains in the package graph",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to PHP repository",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		return analyzer.FindCircularDependencies(repoPath)
	})

	// Tool 9: Analyze Multi Repo
	server.RegisterTool(mcp.Tool{
		Name:        "analyze_multi_repo",
		Description: "Analyze dependencies across multiple repositories (Dependency Buster platform)",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"config_path": {
					Type:        "string",
					Description: "Path to repository configuration JSON file",
				},
			},
			Required: []string{"config_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		configPath := args["config_path"].(string)
		return analyzer.AnalyzeMultipleRepositories(configPath)
	})

	// Tool 10: Generate Comprehensive Docs
	server.RegisterTool(mcp.Tool{
		Name:        "generate_comprehensive_docs",
		Description: "Generate comprehensive markdown documentation for a repository",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to PHP repository",
				},
				"output_path": {
					Type:        "string",
					Description: "Where to save the documentation file",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		outputPath := ""
		if op, ok := args["output_path"].(string); ok {
			outputPath = op
		}
		return analyzer.GenerateComprehensiveDocs(repoPath, outputPath)
	})

	// Tool 15: Generate MkDocs Docs
	server.RegisterTool(mcp.Tool{
		Name:        "generate_mkdocs_docs",
		Description: "Generate MkDocs-compatible documentation site with multi-file structure, navigation, and changelog",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to repository",
				},
				"output_dir": {
					Type:        "string",
					Description: "Output directory for docs (default: docs/)",
				},
				"include_changelog": {
					Type:        "boolean",
					Description: "Include dependency changelog (default: true)",
				},
				"format": {
					Type:        "string",
					Description: "Output format: mkdocs, html, or markdown (default: mkdocs)",
				},
				"site_name": {
					Type:        "string",
					Description: "Site name for mkdocs.yml (optional)",
				},
				"site_description": {
					Type:        "string",
					Description: "Site description for mkdocs.yml (optional)",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		outputDir := ""
		if od, ok := args["output_dir"].(string); ok {
			outputDir = od
		}
		includeChangelog := true
		if ic, ok := args["include_changelog"].(bool); ok {
			includeChangelog = ic
		}
		format := "mkdocs"
		if f, ok := args["format"].(string); ok && f != "" {
			format = f
		}
		siteName := ""
		if sn, ok := args["site_name"].(string); ok {
			siteName = sn
		}
		siteDescription := ""
		if sd, ok := args["site_description"].(string); ok {
			siteDescription = sd
		}
		return analyzer.GenerateMkDocsDocs(analyzer.MkDocsOptions{
			RepoPath:        repoPath,
			OutputDir:       outputDir,
			IncludeChangelog: includeChangelog,
			Format:          format,
			SiteName:        siteName,
			SiteDescription: siteDescription,
		})
	})

	// Tool 11: Track Dependencies
	server.RegisterTool(mcp.Tool{
		Name:        "track_dependencies",
		Description: "Create a timestamped snapshot of dependencies for tracking changes over time",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to repository",
				},
				"save": {
					Type:        "boolean",
					Description: "Save snapshot to disk for future comparison (default: true)",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		save := true
		if s, ok := args["save"].(bool); ok {
			save = s
		}
		snapshot, err := analyzer.CreateDependencySnapshot(repoPath)
		if err != nil {
			return nil, err
		}
		if save {
			if err := analyzer.SaveSnapshot(repoPath, snapshot); err != nil {
				return nil, err
			}
		}
		return snapshot, nil
	})

	// Tool 12: Get Dependency History
	server.RegisterTool(mcp.Tool{
		Name:        "get_dependency_history",
		Description: "Get dependency history with timestamps, recently added/updated, and stale packages",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to repository",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		return analyzer.GetDependencyHistory(repoPath)
	})

	// Tool 13: Check Compliance
	server.RegisterTool(mcp.Tool{
		Name:        "check_compliance",
		Description: "Check dependencies for compliance issues (licenses, outdated, deprecated)",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to repository",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		return analyzer.CheckCompliance(repoPath)
	})

	// Tool 14: Get Agent Suggestions
	server.RegisterTool(mcp.Tool{
		Name:        "get_agent_suggestions",
		Description: "Get structured suggestions for AI agents (Cursor, Cline, Claude Code) about dependency issues",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"repo_path": {
					Type:        "string",
					Description: "Absolute path to repository",
				},
			},
			Required: []string{"repo_path"},
		},
	}, func(args map[string]interface{}) (interface{}, error) {
		repoPath := args["repo_path"].(string)
		return analyzer.GenerateAgentSuggestions(repoPath)
	})
}
