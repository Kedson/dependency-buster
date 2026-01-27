/**
 * Tool Annotations for MCP
 * Based on MCP specification and Backstage patterns
 */

export interface ToolAnnotations {
  /**
   * Human-readable title for the tool
   */
  title?: string;
  
  /**
   * Hint that this tool may have destructive side effects
   * (e.g., deletes files, modifies data)
   */
  destructiveHint?: boolean;
  
  /**
   * Hint that calling this tool multiple times with the same
   * arguments produces the same result
   */
  idempotentHint?: boolean;
  
  /**
   * Hint that this tool only reads data and has no side effects
   */
  readOnlyHint?: boolean;
  
  /**
   * Hint that this tool interacts with external systems
   * beyond the local environment
   */
  openWorldHint?: boolean;
  
  /**
   * Optional tags for categorization
   */
  tags?: string[];
  
  /**
   * Optional cache TTL in seconds (for read-only operations)
   */
  cacheTtlSeconds?: number;
}

/**
 * Enhanced tool definition with annotations
 */
export interface EnhancedTool {
  name: string;
  description: string;
  inputSchema: {
    type: string;
    properties: Record<string, {
      type: string;
      description: string;
    }>;
    required?: string[];
  };
  annotations?: ToolAnnotations;
}

/**
 * Standard annotations for PHP MCP tools
 */
export const StandardAnnotations = {
  // Analysis tools (read-only, cacheable)
  ANALYSIS: {
    readOnlyHint: true,
    idempotentHint: true,
    destructiveHint: false,
    openWorldHint: false,
    cacheTtlSeconds: 300, // 5 minutes
    tags: ['analysis', 'read-only'],
  } as ToolAnnotations,
  
  // Documentation generation (may write files)
  DOCUMENTATION: {
    readOnlyHint: false,
    idempotentHint: true,
    destructiveHint: false,
    openWorldHint: false,
    tags: ['documentation', 'generate'],
  } as ToolAnnotations,
  
  // Security audit (read-only, may access external services)
  SECURITY: {
    readOnlyHint: true,
    idempotentHint: true,
    destructiveHint: false,
    openWorldHint: true, // May query vulnerability databases
    cacheTtlSeconds: 60, // 1 minute (security data changes)
    tags: ['security', 'audit'],
  } as ToolAnnotations,
  
  // Graph generation (read-only)
  VISUALIZATION: {
    readOnlyHint: true,
    idempotentHint: true,
    destructiveHint: false,
    openWorldHint: false,
    cacheTtlSeconds: 300,
    tags: ['visualization', 'graph'],
  } as ToolAnnotations,
  
  // Multi-repo analysis (reads multiple repos)
  MULTI_REPO: {
    readOnlyHint: true,
    idempotentHint: true,
    destructiveHint: false,
    openWorldHint: false,
    cacheTtlSeconds: 120,
    tags: ['multi-repo', 'analysis'],
  } as ToolAnnotations,
};

/**
 * Get annotations for a tool by name
 */
export function getToolAnnotations(toolName: string): ToolAnnotations {
  const annotationMap: Record<string, ToolAnnotations> = {
    // Core analysis tools
    'analyze_dependencies': {
      title: 'Analyze Dependencies',
      ...StandardAnnotations.ANALYSIS,
    },
    'analyze_psr4': {
      title: 'Analyze PSR-4 Autoloading',
      ...StandardAnnotations.ANALYSIS,
    },
    'detect_namespaces': {
      title: 'Detect Namespaces',
      ...StandardAnnotations.ANALYSIS,
    },
    'analyze_namespace_usage': {
      title: 'Analyze Namespace Usage',
      ...StandardAnnotations.ANALYSIS,
    },
    
    // Security tools
    'audit_security': {
      title: 'Audit Security',
      ...StandardAnnotations.SECURITY,
    },
    'analyze_licenses': {
      title: 'Analyze Licenses',
      ...StandardAnnotations.SECURITY,
    },
    
    // Visualization tools
    'generate_dependency_graph': {
      title: 'Generate Dependency Graph',
      ...StandardAnnotations.VISUALIZATION,
    },
    'find_circular_dependencies': {
      title: 'Find Circular Dependencies',
      ...StandardAnnotations.VISUALIZATION,
    },
    
    // Multi-repo tools
    'analyze_multi_repo': {
      title: 'Analyze Multiple Repositories',
      ...StandardAnnotations.MULTI_REPO,
    },
    
    // Documentation tools
    'generate_comprehensive_docs': {
      title: 'Generate Documentation',
      ...StandardAnnotations.DOCUMENTATION,
    },
  };
  
  return annotationMap[toolName] || {};
}
