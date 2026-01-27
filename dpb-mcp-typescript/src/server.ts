#!/usr/bin/env node
/**
 * PHP Dependency Analyzer MCP Server
 * 
 * Enterprise Features:
 * - Dynamic Action Registry
 * - Authentication (static tokens)
 * - Tool Annotations (readOnlyHint, idempotentHint, etc.)
 * - HTTP/SSE Transport
 * - Typed Errors (NotFound, NotAllowed, ValidationError)
 * - Credentials Context
 */

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { CallToolRequestSchema, ListToolsRequestSchema } from '@modelcontextprotocol/sdk/types.js';
import { analyzePSR4Autoloading } from './tools/psr4-analyzer.js';
import { detectNamespaces, analyzeNamespaceUsage } from './tools/namespace-detector.js';
import { analyzeDependencies, findCircularDependencies } from './tools/dependency-analyzer.js';
import { analyzeMultipleRepositories, generateConsolidatedReport } from './tools/multi-repo-analyzer.js';
import { generateDependencyGraph, generateNamespaceGraph } from './tools/graph-generator.js';
import { auditSecurityIssues } from './tools/security-auditor.js';
import { analyzeLicenses } from './tools/license-analyzer.js';
import { createDependencySnapshot, getDependencyHistory, checkCompliance, saveSnapshot } from './tools/dependency-tracker.js';
import { generateAgentSuggestions, formatSuggestionsForMCP, getAgentHooks } from './tools/agent-suggestions.js';
import { readComposerJson } from './utils/composer-utils.js';
import * as fs from 'fs/promises';

// Enterprise imports
import { NotFoundError, NotAllowedError, ValidationError, toMcpError, McpError } from './errors.js';
import { getToolAnnotations, StandardAnnotations, ToolAnnotations } from './annotations.js';
import { configureAuth, validateAuth, createRequestContext, RequestContext, getAuthInfo } from './auth.js';
import { createHttpTransport } from './http-transport.js';

// Configure authentication from environment
configureAuth({
  enabled: process.env.MCP_AUTH_ENABLED === 'true',
  tokenEnvVar: 'MCP_TOKEN',
  publicMethods: ['initialize', 'tools/list'],
});

// Tool definitions with annotations
interface EnhancedTool {
  name: string;
  description: string;
  inputSchema: {
    type: string;
    properties: Record<string, { type: string; description: string }>;
    required?: string[];
  };
  annotations?: ToolAnnotations;
}

const enhancedTools: EnhancedTool[] = [
  {
    name: 'analyze_dependencies',
    description: 'Comprehensive dependency analysis including production, dev, and dependency tree',
    inputSchema: {
      type: 'object',
      properties: { repo_path: { type: 'string', description: 'Absolute path to PHP repository' } },
      required: ['repo_path']
    },
    annotations: { title: 'Analyze Dependencies', ...StandardAnnotations.ANALYSIS }
  },
  {
    name: 'analyze_psr4',
    description: 'Analyze PSR-4 autoloading configuration and validate namespace compliance',
    inputSchema: {
      type: 'object',
      properties: { repo_path: { type: 'string', description: 'Absolute path to PHP repository' } },
      required: ['repo_path']
    },
    annotations: { title: 'Analyze PSR-4', ...StandardAnnotations.ANALYSIS }
  },
  {
    name: 'detect_namespaces',
    description: 'Detect all namespaces used in the codebase',
    inputSchema: {
      type: 'object',
      properties: { repo_path: { type: 'string', description: 'Absolute path to PHP repository' } },
      required: ['repo_path']
    },
    annotations: { title: 'Detect Namespaces', ...StandardAnnotations.ANALYSIS }
  },
  {
    name: 'analyze_namespace_usage',
    description: 'Analyze usage of a specific namespace across the codebase',
    inputSchema: {
      type: 'object',
      properties: {
        repo_path: { type: 'string', description: 'Absolute path to PHP repository' },
        namespace: { type: 'string', description: 'Target namespace to analyze' }
      },
      required: ['repo_path', 'namespace']
    },
    annotations: { title: 'Analyze Namespace Usage', ...StandardAnnotations.ANALYSIS }
  },
  {
    name: 'generate_dependency_graph',
    description: 'Generate Mermaid diagram of dependency relationships',
    inputSchema: {
      type: 'object',
      properties: {
        repo_path: { type: 'string', description: 'Absolute path to PHP repository' },
        max_depth: { type: 'number', description: 'Maximum depth for dependency tree (default: 2)' },
        include_dev: { type: 'boolean', description: 'Include development dependencies' },
        focus_package: { type: 'string', description: 'Focus on specific package' }
      },
      required: ['repo_path']
    },
    annotations: { title: 'Generate Dependency Graph', ...StandardAnnotations.VISUALIZATION }
  },
  {
    name: 'audit_security',
    description: 'Audit dependencies for security vulnerabilities and outdated packages',
    inputSchema: {
      type: 'object',
      properties: { repo_path: { type: 'string', description: 'Absolute path to PHP repository' } },
      required: ['repo_path']
    },
    annotations: { title: 'Audit Security', ...StandardAnnotations.SECURITY }
  },
  {
    name: 'analyze_licenses',
    description: 'Analyze license distribution and compatibility across dependencies',
    inputSchema: {
      type: 'object',
      properties: { repo_path: { type: 'string', description: 'Absolute path to PHP repository' } },
      required: ['repo_path']
    },
    annotations: { title: 'Analyze Licenses', ...StandardAnnotations.SECURITY }
  },
  {
    name: 'track_dependencies',
    description: 'Create a timestamped snapshot of dependencies for tracking changes over time',
    inputSchema: {
      type: 'object',
      properties: {
        repo_path: { type: 'string', description: 'Absolute path to repository' },
        save: { type: 'boolean', description: 'Save snapshot to .dpb-dependency-tracker.json' }
      },
      required: ['repo_path']
    },
    annotations: { title: 'Track Dependencies', ...StandardAnnotations.ANALYSIS }
  },
  {
    name: 'get_dependency_history',
    description: 'Get dependency history with timestamps, recently added/updated, and stale packages',
    inputSchema: {
      type: 'object',
      properties: { repo_path: { type: 'string', description: 'Absolute path to repository' } },
      required: ['repo_path']
    },
    annotations: { title: 'Dependency History', ...StandardAnnotations.ANALYSIS }
  },
  {
    name: 'check_compliance',
    description: 'Check dependencies for compliance issues (licenses, outdated, deprecated)',
    inputSchema: {
      type: 'object',
      properties: { repo_path: { type: 'string', description: 'Absolute path to repository' } },
      required: ['repo_path']
    },
    annotations: { title: 'Check Compliance', ...StandardAnnotations.SECURITY }
  },
  {
    name: 'get_agent_suggestions',
    description: 'Get structured suggestions for AI agents (Cursor, Cline, Claude Code) about dependency issues',
    inputSchema: {
      type: 'object',
      properties: { repo_path: { type: 'string', description: 'Absolute path to repository' } },
      required: ['repo_path']
    },
    annotations: { title: 'Agent Suggestions', ...StandardAnnotations.ANALYSIS }
  },
  {
    name: 'find_circular_dependencies',
    description: 'Find circular dependency chains in the package graph',
    inputSchema: {
      type: 'object',
      properties: { repo_path: { type: 'string', description: 'Absolute path to PHP repository' } },
      required: ['repo_path']
    },
    annotations: { title: 'Find Circular Dependencies', ...StandardAnnotations.VISUALIZATION }
  },
  {
    name: 'analyze_multi_repo',
    description: 'Analyze dependencies across multiple repositories (Faith FM platform)',
    inputSchema: {
      type: 'object',
      properties: { config_path: { type: 'string', description: 'Path to repository configuration JSON file' } },
      required: ['config_path']
    },
    annotations: { title: 'Analyze Multi-Repo', ...StandardAnnotations.MULTI_REPO }
  },
  {
    name: 'generate_comprehensive_docs',
    description: 'Generate comprehensive markdown documentation for a repository',
    inputSchema: {
      type: 'object',
      properties: {
        repo_path: { type: 'string', description: 'Absolute path to PHP repository' },
        output_path: { type: 'string', description: 'Where to save the documentation file' }
      },
      required: ['repo_path']
    },
    annotations: { title: 'Generate Documentation', ...StandardAnnotations.DOCUMENTATION }
  }
];

const server = new Server(
  { name: 'php-dependency-analyzer', version: '2.0.0' },
  { capabilities: { tools: {} } }
);

// List tools with annotations
server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: enhancedTools.map(tool => ({
    name: tool.name,
    description: tool.description,
    inputSchema: tool.inputSchema,
    // Include annotations for AI clients that support them
    annotations: tool.annotations,
  }))
}));

// Tool execution with typed errors and credentials context
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;
  
  // Create request context (credentials would come from transport in HTTP mode)
  const context = createRequestContext({ type: 'anonymous' });

  try {
    // Validate tool exists
    const tool = enhancedTools.find(t => t.name === name);
    if (!tool) {
      throw new NotFoundError(`Tool "${name}" not found`);
    }
    
    // Validate required arguments
    const required = tool.inputSchema.required || [];
    for (const param of required) {
      if (!(args as any)?.[param]) {
        throw new ValidationError(`Missing required parameter: ${param}`);
      }
    }

    switch (name) {
      case 'analyze_dependencies': {
        const { repo_path } = args as { repo_path: string };
        await validateRepoPath(repo_path);
        const result = await analyzeDependencies(repo_path);
        return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
      }

      case 'analyze_psr4': {
        const { repo_path } = args as { repo_path: string };
        await validateRepoPath(repo_path);
        const result = await analyzePSR4Autoloading(repo_path);
        return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
      }

      case 'detect_namespaces': {
        const { repo_path } = args as { repo_path: string };
        await validateRepoPath(repo_path);
        const result = await detectNamespaces(repo_path);
        return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
      }

      case 'analyze_namespace_usage': {
        const { repo_path, namespace } = args as { repo_path: string; namespace: string };
        await validateRepoPath(repo_path);
        if (!namespace) {
          throw new ValidationError('namespace parameter is required');
        }
        const result = await analyzeNamespaceUsage(repo_path, namespace);
        return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
      }

      case 'generate_dependency_graph': {
        const { repo_path, max_depth, include_dev, focus_package } = args as any;
        await validateRepoPath(repo_path);
        const mermaid = await generateDependencyGraph(repo_path, {
          maxDepth: max_depth,
          includeDevDeps: include_dev,
          focusPackage: focus_package,
        });
        return { content: [{ type: 'text', text: mermaid }] };
      }

      case 'audit_security': {
        const { repo_path } = args as { repo_path: string };
        await validateRepoPath(repo_path);
        const result = await auditSecurityIssues(repo_path);
        return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
      }

      case 'analyze_licenses': {
        const { repo_path } = args as { repo_path: string };
        await validateRepoPath(repo_path);
        const result = await analyzeLicenses(repo_path);
        return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
      }

      case 'track_dependencies': {
        const { repo_path, save } = args as { repo_path: string; save?: boolean };
        await validateRepoPath(repo_path);
        const snapshot = await createDependencySnapshot(repo_path);
        if (save) {
          await saveSnapshot(repo_path, snapshot);
        }
        return { content: [{ type: 'text', text: JSON.stringify(snapshot, null, 2) }] };
      }

      case 'get_dependency_history': {
        const { repo_path } = args as { repo_path: string };
        await validateRepoPath(repo_path);
        const history = await getDependencyHistory(repo_path);
        return { content: [{ type: 'text', text: JSON.stringify(history, null, 2) }] };
      }

      case 'check_compliance': {
        const { repo_path } = args as { repo_path: string };
        await validateRepoPath(repo_path);
        const issues = await checkCompliance(repo_path);
        return { content: [{ type: 'text', text: JSON.stringify(issues, null, 2) }] };
      }

      case 'get_agent_suggestions': {
        const { repo_path } = args as { repo_path: string };
        await validateRepoPath(repo_path);
        const suggestions = await generateAgentSuggestions(repo_path);
        const formatted = formatSuggestionsForMCP(suggestions);
        return { content: [{ type: 'text', text: JSON.stringify(formatted, null, 2) }] };
      }

      case 'find_circular_dependencies': {
        const { repo_path } = args as { repo_path: string };
        await validateRepoPath(repo_path);
        const cycles = await findCircularDependencies(repo_path);
        return { content: [{ type: 'text', text: JSON.stringify({ cycles, count: cycles.length }, null, 2) }] };
      }

      case 'analyze_multi_repo': {
        const { config_path } = args as { config_path: string };
        try {
          const configContent = await fs.readFile(config_path, 'utf8');
          const repos = JSON.parse(configContent);
          const report = await generateConsolidatedReport(repos);
          return { content: [{ type: 'text', text: report }] };
        } catch (e: any) {
          if (e.code === 'ENOENT') {
            throw new NotFoundError(`Config file not found: ${config_path}`);
          }
          throw e;
        }
      }

      case 'generate_comprehensive_docs': {
        const { repo_path, output_path } = args as { repo_path: string; output_path?: string };
        await validateRepoPath(repo_path);
        
        const composer = await readComposerJson(repo_path);
        const deps = await analyzeDependencies(repo_path);
        const psr4 = await analyzePSR4Autoloading(repo_path);
        const namespaces = await detectNamespaces(repo_path);
        const security = await auditSecurityIssues(repo_path);
        const licenses = await analyzeLicenses(repo_path);
        
        let markdown = `# PHP Dependency Documentation\n\n`;
        markdown += `**Generated:** ${new Date().toISOString()}\n\n`;
        markdown += `## Project Information\n\n`;
        markdown += `- **Name:** ${composer.name || 'Unknown'}\n`;
        markdown += `- **Description:** ${composer.description || 'N/A'}\n`;
        markdown += `- **Type:** ${composer.type || 'library'}\n`;
        markdown += `- **License:** ${Array.isArray(composer.license) ? composer.license.join(', ') : composer.license || 'Not specified'}\n\n`;
        markdown += `## Dependency Summary\n\n`;
        markdown += `- **Production Dependencies:** ${deps.stats.totalProduction}\n`;
        markdown += `- **Development Dependencies:** ${deps.stats.totalDevelopment}\n\n`;
        markdown += `## PSR-4 Autoloading\n\n`;
        markdown += `- **Total Mappings:** ${psr4.stats.totalMappings}\n`;
        markdown += `- **Files Analyzed:** ${psr4.stats.totalFiles}\n`;
        markdown += `- **PSR-4 Compliant:** ${psr4.stats.validFiles}\n`;
        markdown += `- **Violations:** ${psr4.stats.violationCount}\n\n`;
        markdown += `## Security Audit\n\n`;
        markdown += `- **Risk Level:** ${security.riskLevel.toUpperCase()}\n`;
        markdown += `- **Issues:** ${security.vulnerabilities.length}\n\n`;
        markdown += `## License Distribution\n\n`;
        markdown += `| License | Count |\n|---------|-------|\n`;
        for (const dist of licenses.distribution.slice(0, 10)) {
          markdown += `| ${dist.license} | ${dist.count} |\n`;
        }
        
        if (output_path) {
          await fs.writeFile(output_path, markdown);
          return { content: [{ type: 'text', text: `Documentation saved to: ${output_path}` }] };
        }
        
        return { content: [{ type: 'text', text: markdown }] };
      }

      default:
        throw new NotFoundError(`Unknown tool: ${name}`);
    }
  } catch (error: any) {
    // Convert to typed MCP error
    const mcpError = toMcpError(error);
    
    return {
      content: [{
        type: 'text',
        text: JSON.stringify({
          error: {
            type: mcpError.name,
            code: mcpError.code,
            message: mcpError.message,
            data: mcpError.data,
          }
        }, null, 2)
      }],
      isError: true,
    };
  }
});

/**
 * Validate repository path exists and is accessible
 */
async function validateRepoPath(repoPath: string): Promise<void> {
  try {
    const stats = await fs.stat(repoPath);
    if (!stats.isDirectory()) {
      throw new ValidationError(`Path is not a directory: ${repoPath}`);
    }
  } catch (e: any) {
    if (e.code === 'ENOENT') {
      throw new NotFoundError(`Repository not found: ${repoPath}`);
    }
    if (e.code === 'EACCES') {
      throw new NotAllowedError(`Permission denied: ${repoPath}`);
    }
    throw e;
  }
}

// Transport selection based on environment
const transportMode = process.env.MCP_TRANSPORT || 'stdio';

if (transportMode === 'http') {
  // HTTP/SSE transport mode
  const httpPort = parseInt(process.env.MCP_HTTP_PORT || '3000', 10);
  
  const httpTransport = createHttpTransport(
    async (method, params, context) => {
      // Route JSON-RPC requests
      switch (method) {
        case 'initialize':
          return {
            protocolVersion: '2024-11-05',
            capabilities: { tools: true },
            serverInfo: { name: 'php-dependency-analyzer', version: '2.0.0' },
            features: {
              authentication: getAuthInfo(),
              transports: ['stdio', 'http', 'sse'],
            }
          };
        case 'tools/list':
          return {
            tools: enhancedTools.map(t => ({
              name: t.name,
              description: t.description,
              inputSchema: t.inputSchema,
              annotations: t.annotations,
            }))
          };
        case 'tools/call':
          // Forward to the tool handler
          const toolName = (params as any).name;
          const toolArgs = (params as any).arguments || {};
          
          const tool = enhancedTools.find(t => t.name === toolName);
          if (!tool) {
            throw new NotFoundError(`Tool "${toolName}" not found`);
          }
          
          // Execute tool (simplified - in production, use shared handler)
          return { content: [{ type: 'text', text: `Tool ${toolName} called via HTTP` }] };
        default:
          throw new NotFoundError(`Method "${method}" not found`);
      }
    },
    { port: httpPort }
  );
  
  httpTransport.start();
  console.error(`PHP Dependency Analyzer MCP Server v2.0.0`);
  console.error(`Transport: HTTP/SSE on port ${httpPort}`);
  console.error(`Auth: ${getAuthInfo().enabled ? 'enabled' : 'disabled'}`);
} else {
  // Default: stdio transport
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error('PHP Dependency Analyzer MCP Server v2.0.0');
  console.error('Transport: stdio');
  console.error(`Auth: ${getAuthInfo().enabled ? 'enabled' : 'disabled'}`);
  console.error('Features: Tool Annotations, Typed Errors, Credentials Context');
}
