package mcp

// ToolAnnotations represents hints about tool behavior for AI clients
type ToolAnnotations struct {
	Title          string   `json:"title,omitempty"`
	DestructiveHint bool    `json:"destructiveHint,omitempty"`
	IdempotentHint  bool    `json:"idempotentHint,omitempty"`
	ReadOnlyHint    bool    `json:"readOnlyHint,omitempty"`
	OpenWorldHint   bool    `json:"openWorldHint,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	CacheTTLSeconds int      `json:"cacheTtlSeconds,omitempty"`
}

// Standard annotation presets
var (
	// AnalysisAnnotation for read-only analysis tools
	AnalysisAnnotation = ToolAnnotations{
		ReadOnlyHint:    true,
		IdempotentHint:  true,
		DestructiveHint: false,
		OpenWorldHint:   false,
		CacheTTLSeconds: 300,
		Tags:            []string{"analysis", "read-only"},
	}

	// DocumentationAnnotation for doc generation tools
	DocumentationAnnotation = ToolAnnotations{
		ReadOnlyHint:    false,
		IdempotentHint:  true,
		DestructiveHint: false,
		OpenWorldHint:   false,
		Tags:            []string{"documentation", "generate"},
	}

	// SecurityAnnotation for security audit tools
	SecurityAnnotation = ToolAnnotations{
		ReadOnlyHint:    true,
		IdempotentHint:  true,
		DestructiveHint: false,
		OpenWorldHint:   true,
		CacheTTLSeconds: 60,
		Tags:            []string{"security", "audit"},
	}

	// VisualizationAnnotation for graph/diagram tools
	VisualizationAnnotation = ToolAnnotations{
		ReadOnlyHint:    true,
		IdempotentHint:  true,
		DestructiveHint: false,
		OpenWorldHint:   false,
		CacheTTLSeconds: 300,
		Tags:            []string{"visualization", "graph"},
	}

	// MultiRepoAnnotation for multi-repo analysis
	MultiRepoAnnotation = ToolAnnotations{
		ReadOnlyHint:    true,
		IdempotentHint:  true,
		DestructiveHint: false,
		OpenWorldHint:   false,
		CacheTTLSeconds: 120,
		Tags:            []string{"multi-repo", "analysis"},
	}
)

// GetToolAnnotation returns the appropriate annotation for a tool
func GetToolAnnotation(toolName string) ToolAnnotations {
	annotations := map[string]ToolAnnotations{
		"analyze_dependencies": {
			Title: "Analyze Dependencies",
			ReadOnlyHint:    AnalysisAnnotation.ReadOnlyHint,
			IdempotentHint:  AnalysisAnnotation.IdempotentHint,
			DestructiveHint: AnalysisAnnotation.DestructiveHint,
			OpenWorldHint:   AnalysisAnnotation.OpenWorldHint,
			CacheTTLSeconds: AnalysisAnnotation.CacheTTLSeconds,
			Tags:            AnalysisAnnotation.Tags,
		},
		"analyze_psr4": {
			Title: "Analyze PSR-4",
			ReadOnlyHint:    AnalysisAnnotation.ReadOnlyHint,
			IdempotentHint:  AnalysisAnnotation.IdempotentHint,
			DestructiveHint: AnalysisAnnotation.DestructiveHint,
			OpenWorldHint:   AnalysisAnnotation.OpenWorldHint,
			CacheTTLSeconds: AnalysisAnnotation.CacheTTLSeconds,
			Tags:            AnalysisAnnotation.Tags,
		},
		"detect_namespaces": {
			Title: "Detect Namespaces",
			ReadOnlyHint:    AnalysisAnnotation.ReadOnlyHint,
			IdempotentHint:  AnalysisAnnotation.IdempotentHint,
			DestructiveHint: AnalysisAnnotation.DestructiveHint,
			OpenWorldHint:   AnalysisAnnotation.OpenWorldHint,
			CacheTTLSeconds: AnalysisAnnotation.CacheTTLSeconds,
			Tags:            AnalysisAnnotation.Tags,
		},
		"analyze_namespace_usage": {
			Title: "Analyze Namespace Usage",
			ReadOnlyHint:    AnalysisAnnotation.ReadOnlyHint,
			IdempotentHint:  AnalysisAnnotation.IdempotentHint,
			DestructiveHint: AnalysisAnnotation.DestructiveHint,
			OpenWorldHint:   AnalysisAnnotation.OpenWorldHint,
			CacheTTLSeconds: AnalysisAnnotation.CacheTTLSeconds,
			Tags:            AnalysisAnnotation.Tags,
		},
		"audit_security": {
			Title: "Audit Security",
			ReadOnlyHint:    SecurityAnnotation.ReadOnlyHint,
			IdempotentHint:  SecurityAnnotation.IdempotentHint,
			DestructiveHint: SecurityAnnotation.DestructiveHint,
			OpenWorldHint:   SecurityAnnotation.OpenWorldHint,
			CacheTTLSeconds: SecurityAnnotation.CacheTTLSeconds,
			Tags:            SecurityAnnotation.Tags,
		},
		"analyze_licenses": {
			Title: "Analyze Licenses",
			ReadOnlyHint:    SecurityAnnotation.ReadOnlyHint,
			IdempotentHint:  SecurityAnnotation.IdempotentHint,
			DestructiveHint: SecurityAnnotation.DestructiveHint,
			OpenWorldHint:   SecurityAnnotation.OpenWorldHint,
			CacheTTLSeconds: SecurityAnnotation.CacheTTLSeconds,
			Tags:            SecurityAnnotation.Tags,
		},
		"generate_dependency_graph": {
			Title: "Generate Dependency Graph",
			ReadOnlyHint:    VisualizationAnnotation.ReadOnlyHint,
			IdempotentHint:  VisualizationAnnotation.IdempotentHint,
			DestructiveHint: VisualizationAnnotation.DestructiveHint,
			OpenWorldHint:   VisualizationAnnotation.OpenWorldHint,
			CacheTTLSeconds: VisualizationAnnotation.CacheTTLSeconds,
			Tags:            VisualizationAnnotation.Tags,
		},
		"find_circular_dependencies": {
			Title: "Find Circular Dependencies",
			ReadOnlyHint:    VisualizationAnnotation.ReadOnlyHint,
			IdempotentHint:  VisualizationAnnotation.IdempotentHint,
			DestructiveHint: VisualizationAnnotation.DestructiveHint,
			OpenWorldHint:   VisualizationAnnotation.OpenWorldHint,
			CacheTTLSeconds: VisualizationAnnotation.CacheTTLSeconds,
			Tags:            VisualizationAnnotation.Tags,
		},
		"analyze_multi_repo": {
			Title: "Analyze Multi-Repo",
			ReadOnlyHint:    MultiRepoAnnotation.ReadOnlyHint,
			IdempotentHint:  MultiRepoAnnotation.IdempotentHint,
			DestructiveHint: MultiRepoAnnotation.DestructiveHint,
			OpenWorldHint:   MultiRepoAnnotation.OpenWorldHint,
			CacheTTLSeconds: MultiRepoAnnotation.CacheTTLSeconds,
			Tags:            MultiRepoAnnotation.Tags,
		},
		"generate_comprehensive_docs": {
			Title: "Generate Documentation",
			ReadOnlyHint:    DocumentationAnnotation.ReadOnlyHint,
			IdempotentHint:  DocumentationAnnotation.IdempotentHint,
			DestructiveHint: DocumentationAnnotation.DestructiveHint,
			OpenWorldHint:   DocumentationAnnotation.OpenWorldHint,
			Tags:            DocumentationAnnotation.Tags,
		},
	}

	if ann, ok := annotations[toolName]; ok {
		return ann
	}
	return ToolAnnotations{}
}
