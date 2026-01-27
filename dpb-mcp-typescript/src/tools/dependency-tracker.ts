/**
 * Dependency Tracker - Timestamps and versioning for dependency changes
 * Enables reverting or replacing non-compliant dependencies
 */

import * as fs from 'fs';
import * as path from 'path';
import { readComposerJson, readComposerLock } from '../utils/composer-utils.js';

export interface DependencySnapshot {
  timestamp: string;
  checksum: string;
  dependencies: {
    name: string;
    version: string;
    type: 'production' | 'development';
    addedAt?: string;
    updatedAt?: string;
    license?: string;
    securityStatus?: 'safe' | 'vulnerable' | 'unknown';
  }[];
  metadata: {
    repoPath: string;
    packageManager: string;
    totalCount: number;
  };
}

export interface DependencyChange {
  type: 'added' | 'removed' | 'updated';
  name: string;
  oldVersion?: string;
  newVersion?: string;
  timestamp: string;
  reason?: string;
}

export interface ComplianceIssue {
  dependency: string;
  version: string;
  issue: 'license' | 'security' | 'outdated' | 'deprecated';
  severity: 'critical' | 'high' | 'medium' | 'low';
  description: string;
  recommendation: string;
  autoFixAvailable: boolean;
}

const TRACKER_FILE = '.dpb-dependency-tracker.json';

/**
 * Create a snapshot of current dependencies with timestamps
 */
export async function createDependencySnapshot(repoPath: string): Promise<DependencySnapshot> {
  const composer = await readComposerJson(repoPath);
  const lock = await readComposerLock(repoPath);
  
  const now = new Date().toISOString();
  const dependencies: DependencySnapshot['dependencies'] = [];
  
  // Load existing tracker to preserve addedAt timestamps
  const existingTracker = await loadTracker(repoPath);
  const existingDeps = new Map(
    existingTracker?.dependencies.map(d => [d.name, d]) || []
  );
  
  // Process production dependencies
  if (lock?.packages) {
    for (const pkg of lock.packages) {
      const existing = existingDeps.get(pkg.name);
      dependencies.push({
        name: pkg.name,
        version: pkg.version,
        type: 'production',
        addedAt: existing?.addedAt || now,
        updatedAt: existing?.version !== pkg.version ? now : existing?.updatedAt,
        license: Array.isArray(pkg.license) ? pkg.license[0] : pkg.license,
        securityStatus: 'unknown',
      });
    }
  }
  
  // Process dev dependencies
  if (lock?.['packages-dev']) {
    for (const pkg of lock['packages-dev']) {
      const existing = existingDeps.get(pkg.name);
      dependencies.push({
        name: pkg.name,
        version: pkg.version,
        type: 'development',
        addedAt: existing?.addedAt || now,
        updatedAt: existing?.version !== pkg.version ? now : existing?.updatedAt,
        license: Array.isArray(pkg.license) ? pkg.license[0] : pkg.license,
        securityStatus: 'unknown',
      });
    }
  }
  
  // Calculate checksum from sorted dependency list
  const checksum = Buffer.from(
    dependencies.map(d => `${d.name}@${d.version}`).sort().join('|')
  ).toString('base64').slice(0, 16);
  
  const snapshot: DependencySnapshot = {
    timestamp: now,
    checksum,
    dependencies,
    metadata: {
      repoPath,
      packageManager: 'composer',
      totalCount: dependencies.length,
    },
  };
  
  return snapshot;
}

/**
 * Load existing tracker data
 */
async function loadTracker(repoPath: string): Promise<DependencySnapshot | null> {
  const trackerPath = path.join(repoPath, TRACKER_FILE);
  try {
    const content = await fs.promises.readFile(trackerPath, 'utf-8');
    return JSON.parse(content);
  } catch {
    return null;
  }
}

/**
 * Save tracker data
 */
export async function saveSnapshot(repoPath: string, snapshot: DependencySnapshot): Promise<void> {
  const trackerPath = path.join(repoPath, TRACKER_FILE);
  await fs.promises.writeFile(trackerPath, JSON.stringify(snapshot, null, 2));
}

/**
 * Compare two snapshots and return changes
 */
export function compareSnapshots(
  oldSnapshot: DependencySnapshot,
  newSnapshot: DependencySnapshot
): DependencyChange[] {
  const changes: DependencyChange[] = [];
  const oldDeps = new Map(oldSnapshot.dependencies.map(d => [d.name, d]));
  const newDeps = new Map(newSnapshot.dependencies.map(d => [d.name, d]));
  
  // Find added and updated
  for (const [name, newDep] of newDeps) {
    const oldDep = oldDeps.get(name);
    if (!oldDep) {
      changes.push({
        type: 'added',
        name,
        newVersion: newDep.version,
        timestamp: newSnapshot.timestamp,
      });
    } else if (oldDep.version !== newDep.version) {
      changes.push({
        type: 'updated',
        name,
        oldVersion: oldDep.version,
        newVersion: newDep.version,
        timestamp: newSnapshot.timestamp,
      });
    }
  }
  
  // Find removed
  for (const [name, oldDep] of oldDeps) {
    if (!newDeps.has(name)) {
      changes.push({
        type: 'removed',
        name,
        oldVersion: oldDep.version,
        timestamp: newSnapshot.timestamp,
      });
    }
  }
  
  return changes;
}

/**
 * Get dependency history with all timestamps
 */
export async function getDependencyHistory(repoPath: string): Promise<{
  currentSnapshot: DependencySnapshot;
  recentlyAdded: DependencySnapshot['dependencies'];
  recentlyUpdated: DependencySnapshot['dependencies'];
  stale: DependencySnapshot['dependencies'];
}> {
  const snapshot = await createDependencySnapshot(repoPath);
  const now = new Date();
  const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
  const oneYearAgo = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000);
  
  const recentlyAdded = snapshot.dependencies.filter(d => 
    d.addedAt && new Date(d.addedAt) > thirtyDaysAgo
  );
  
  const recentlyUpdated = snapshot.dependencies.filter(d =>
    d.updatedAt && new Date(d.updatedAt) > thirtyDaysAgo && d.addedAt !== d.updatedAt
  );
  
  const stale = snapshot.dependencies.filter(d =>
    d.updatedAt && new Date(d.updatedAt) < oneYearAgo
  );
  
  return {
    currentSnapshot: snapshot,
    recentlyAdded,
    recentlyUpdated,
    stale,
  };
}

/**
 * Identify compliance issues with dependencies
 */
export async function checkCompliance(repoPath: string): Promise<ComplianceIssue[]> {
  const snapshot = await createDependencySnapshot(repoPath);
  const issues: ComplianceIssue[] = [];
  
  // License compliance checks
  const restrictiveLicenses = ['GPL-3.0', 'AGPL-3.0', 'GPL-2.0', 'SSPL'];
  const unknownLicenses = ['proprietary', 'unknown', undefined];
  
  for (const dep of snapshot.dependencies) {
    // Check for restrictive licenses in production
    if (dep.type === 'production' && dep.license) {
      if (restrictiveLicenses.some(l => dep.license?.toUpperCase().includes(l.toUpperCase()))) {
        issues.push({
          dependency: dep.name,
          version: dep.version,
          issue: 'license',
          severity: 'high',
          description: `Uses restrictive license: ${dep.license}`,
          recommendation: `Consider replacing with an MIT/Apache-2.0 licensed alternative`,
          autoFixAvailable: false,
        });
      }
      
      if (unknownLicenses.includes(dep.license?.toLowerCase())) {
        issues.push({
          dependency: dep.name,
          version: dep.version,
          issue: 'license',
          severity: 'medium',
          description: `License is unknown or proprietary`,
          recommendation: `Review the package license before production use`,
          autoFixAvailable: false,
        });
      }
    }
    
    // Check for stale dependencies (more than 2 years old)
    if (dep.updatedAt) {
      const updateDate = new Date(dep.updatedAt);
      const twoYearsAgo = new Date(Date.now() - 2 * 365 * 24 * 60 * 60 * 1000);
      if (updateDate < twoYearsAgo) {
        issues.push({
          dependency: dep.name,
          version: dep.version,
          issue: 'outdated',
          severity: 'low',
          description: `Not updated in over 2 years`,
          recommendation: `Check if a newer version is available or if package is abandoned`,
          autoFixAvailable: true,
        });
      }
    }
  }
  
  return issues;
}

/**
 * Generate revert command for a dependency change
 */
export function generateRevertCommand(change: DependencyChange): string {
  switch (change.type) {
    case 'added':
      return `composer remove ${change.name}`;
    case 'removed':
      return `composer require ${change.name}:${change.oldVersion}`;
    case 'updated':
      return `composer require ${change.name}:${change.oldVersion}`;
    default:
      return '';
  }
}

/**
 * Generate replacement suggestions for non-compliant dependencies
 */
export async function suggestReplacements(issues: ComplianceIssue[]): Promise<{
  issue: ComplianceIssue;
  suggestions: string[];
}[]> {
  const suggestions: { issue: ComplianceIssue; suggestions: string[] }[] = [];
  
  // Common replacement mappings
  const replacements: Record<string, string[]> = {
    // GPL alternatives
    'phpmailer/phpmailer': ['symfony/mailer', 'swiftmailer/swiftmailer'],
    'intervention/image': ['spatie/image'],
    // Deprecated packages
    'fzaninotto/faker': ['fakerphp/faker'],
    'phpunit/php-token-stream': ['(removed in PHPUnit 10)'],
  };
  
  for (const issue of issues) {
    const depName = issue.dependency.toLowerCase();
    const replacement = Object.entries(replacements).find(([k]) => 
      depName.includes(k.split('/')[1])
    );
    
    suggestions.push({
      issue,
      suggestions: replacement ? replacement[1] : ['No automatic suggestion available'],
    });
  }
  
  return suggestions;
}
