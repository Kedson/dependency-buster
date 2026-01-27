/**
 * Agent Suggestion Hooks - Integration with IDE agents (Cursor, Cline, Claude Code)
 * Provides structured suggestions for non-compliant dependencies
 */

import { ComplianceIssue, suggestReplacements, checkCompliance, getDependencyHistory } from './dependency-tracker.js';

export interface AgentSuggestion {
  id: string;
  type: 'warning' | 'error' | 'info' | 'action';
  title: string;
  description: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  category: 'security' | 'license' | 'outdated' | 'deprecated' | 'performance';
  dependency?: string;
  version?: string;
  actions: AgentAction[];
  metadata: Record<string, unknown>;
}

export interface AgentAction {
  id: string;
  label: string;
  command: string;
  type: 'shell' | 'file-edit' | 'prompt' | 'link';
  autoApply?: boolean;
  confirmRequired?: boolean;
  description?: string;
}

export interface AgentContext {
  agentType: 'cursor' | 'cline' | 'claude-code' | 'vscode' | 'generic';
  capabilities: {
    canExecuteShell: boolean;
    canEditFiles: boolean;
    canShowPrompts: boolean;
    canOpenLinks: boolean;
  };
  workspacePath: string;
}

/**
 * Generate suggestions formatted for AI agents
 */
export async function generateAgentSuggestions(
  repoPath: string,
  context?: Partial<AgentContext>
): Promise<AgentSuggestion[]> {
  const suggestions: AgentSuggestion[] = [];
  
  // Get compliance issues
  const issues = await checkCompliance(repoPath);
  const replacements = await suggestReplacements(issues);
  
  // Get dependency history
  const history = await getDependencyHistory(repoPath);
  
  // Convert compliance issues to agent suggestions
  for (const { issue, suggestions: replacementOptions } of replacements) {
    const suggestionId = `dep-${issue.issue}-${issue.dependency.replace(/\//g, '-')}`;
    
    const actions: AgentAction[] = [];
    
    // Add replacement action if available
    if (replacementOptions.length > 0 && replacementOptions[0] !== 'No automatic suggestion available') {
      actions.push({
        id: `${suggestionId}-replace`,
        label: `Replace with ${replacementOptions[0]}`,
        command: `composer remove ${issue.dependency} && composer require ${replacementOptions[0]}`,
        type: 'shell',
        autoApply: false,
        confirmRequired: true,
        description: `Replace ${issue.dependency} with ${replacementOptions[0]}`,
      });
    }
    
    // Add update action for outdated packages
    if (issue.issue === 'outdated') {
      actions.push({
        id: `${suggestionId}-update`,
        label: 'Update to latest',
        command: `composer update ${issue.dependency}`,
        type: 'shell',
        autoApply: issue.severity === 'low',
        confirmRequired: issue.severity !== 'low',
      });
    }
    
    // Add documentation link
    actions.push({
      id: `${suggestionId}-docs`,
      label: 'View on Packagist',
      command: `https://packagist.org/packages/${issue.dependency}`,
      type: 'link',
    });
    
    suggestions.push({
      id: suggestionId,
      type: issue.severity === 'critical' || issue.severity === 'high' ? 'error' : 'warning',
      title: `${issue.issue.charAt(0).toUpperCase() + issue.issue.slice(1)} Issue: ${issue.dependency}`,
      description: issue.description,
      severity: issue.severity,
      category: issue.issue as AgentSuggestion['category'],
      dependency: issue.dependency,
      version: issue.version,
      actions,
      metadata: {
        recommendation: issue.recommendation,
        autoFixAvailable: issue.autoFixAvailable,
        replacementOptions,
      },
    });
  }
  
  // Add suggestions for stale dependencies
  for (const staleDep of history.stale.slice(0, 5)) {
    suggestions.push({
      id: `stale-${staleDep.name.replace(/\//g, '-')}`,
      type: 'info',
      title: `Stale Dependency: ${staleDep.name}`,
      description: `This dependency hasn't been updated since ${staleDep.updatedAt}`,
      severity: 'low',
      category: 'outdated',
      dependency: staleDep.name,
      version: staleDep.version,
      actions: [
        {
          id: `stale-${staleDep.name}-update`,
          label: 'Check for updates',
          command: `composer outdated ${staleDep.name}`,
          type: 'shell',
        },
      ],
      metadata: {
        lastUpdated: staleDep.updatedAt,
        addedAt: staleDep.addedAt,
      },
    });
  }
  
  // Add summary suggestion
  if (suggestions.length > 0) {
    const criticalCount = suggestions.filter(s => s.severity === 'critical').length;
    const highCount = suggestions.filter(s => s.severity === 'high').length;
    
    suggestions.unshift({
      id: 'summary',
      type: criticalCount > 0 ? 'error' : highCount > 0 ? 'warning' : 'info',
      title: 'Dependency Analysis Summary',
      description: `Found ${suggestions.length - 1} issues: ${criticalCount} critical, ${highCount} high severity`,
      severity: criticalCount > 0 ? 'critical' : highCount > 0 ? 'high' : 'medium',
      category: 'security',
      actions: [
        {
          id: 'summary-audit',
          label: 'Run full audit',
          command: 'composer audit',
          type: 'shell',
        },
        {
          id: 'summary-update-all',
          label: 'Update all dependencies',
          command: 'composer update',
          type: 'shell',
          confirmRequired: true,
        },
      ],
      metadata: {
        totalDependencies: history.currentSnapshot.metadata.totalCount,
        issueBreakdown: {
          critical: criticalCount,
          high: highCount,
          medium: suggestions.filter(s => s.severity === 'medium').length,
          low: suggestions.filter(s => s.severity === 'low').length,
        },
      },
    });
  }
  
  return suggestions;
}

/**
 * Format suggestions for terminal/CLI output
 */
export function formatSuggestionsForTerminal(suggestions: AgentSuggestion[]): string {
  const lines: string[] = [];
  
  lines.push('');
  lines.push('╔══════════════════════════════════════════════════════════════════╗');
  lines.push('║  dependency-buster // Agent Suggestions                          ║');
  lines.push('╚══════════════════════════════════════════════════════════════════╝');
  lines.push('');
  
  const severityIcons: Record<string, string> = {
    critical: '🔴',
    high: '🟠',
    medium: '🟡',
    low: '🟢',
  };
  
  const typeIcons: Record<string, string> = {
    error: '✗',
    warning: '⚠',
    info: 'ℹ',
    action: '→',
  };
  
  for (const suggestion of suggestions) {
    const icon = severityIcons[suggestion.severity] || '○';
    const typeIcon = typeIcons[suggestion.type] || '•';
    
    lines.push(`${icon} ${typeIcon} ${suggestion.title}`);
    lines.push(`   ${suggestion.description}`);
    
    if (suggestion.dependency) {
      lines.push(`   Package: ${suggestion.dependency}@${suggestion.version || 'unknown'}`);
    }
    
    if (suggestion.actions.length > 0) {
      lines.push('   Actions:');
      for (const action of suggestion.actions) {
        if (action.type === 'shell') {
          lines.push(`     $ ${action.command}`);
        } else if (action.type === 'link') {
          lines.push(`     🔗 ${action.command}`);
        }
      }
    }
    
    lines.push('');
  }
  
  return lines.join('\n');
}

/**
 * Format suggestions for JSON output (for MCP response)
 */
export function formatSuggestionsForMCP(suggestions: AgentSuggestion[]): {
  summary: {
    total: number;
    bySeverity: Record<string, number>;
    byCategory: Record<string, number>;
  };
  suggestions: AgentSuggestion[];
  terminalOutput: string;
} {
  const bySeverity: Record<string, number> = {};
  const byCategory: Record<string, number> = {};
  
  for (const s of suggestions) {
    bySeverity[s.severity] = (bySeverity[s.severity] || 0) + 1;
    byCategory[s.category] = (byCategory[s.category] || 0) + 1;
  }
  
  return {
    summary: {
      total: suggestions.length,
      bySeverity,
      byCategory,
    },
    suggestions,
    terminalOutput: formatSuggestionsForTerminal(suggestions),
  };
}

/**
 * Hook for Cursor/Cline/Claude Code agents
 * Returns structured data that agents can use for inline suggestions
 */
export async function getAgentHooks(repoPath: string): Promise<{
  inlineSuggestions: {
    file: string;
    line?: number;
    message: string;
    severity: string;
    quickFix?: string;
  }[];
  diagnostics: {
    code: string;
    message: string;
    severity: 'error' | 'warning' | 'info';
    source: 'dependency-buster';
  }[];
  codeActions: {
    title: string;
    kind: string;
    command: string;
    arguments: string[];
  }[];
}> {
  const suggestions = await generateAgentSuggestions(repoPath);
  
  // Convert to inline suggestions (for composer.json)
  const inlineSuggestions = suggestions
    .filter(s => s.dependency)
    .map(s => ({
      file: 'composer.json',
      message: s.description,
      severity: s.severity,
      quickFix: s.actions.find(a => a.type === 'shell')?.command,
    }));
  
  // Convert to LSP-style diagnostics
  const diagnostics = suggestions.map(s => ({
    code: s.id,
    message: `${s.title}: ${s.description}`,
    severity: s.type === 'error' ? 'error' as const : s.type === 'warning' ? 'warning' as const : 'info' as const,
    source: 'dependency-buster' as const,
  }));
  
  // Convert to code actions
  const codeActions = suggestions
    .flatMap(s => s.actions)
    .filter(a => a.type === 'shell')
    .map(a => ({
      title: a.label,
      kind: 'quickfix',
      command: 'workbench.action.terminal.sendSequence',
      arguments: [a.command],
    }));
  
  return {
    inlineSuggestions,
    diagnostics,
    codeActions,
  };
}
