import { readComposerLock } from '../utils/composer-utils.js';
import { SecurityVulnerability } from '../types/index.js';

export async function auditSecurityIssues(
  repoPath: string
): Promise<{
  vulnerabilities: SecurityVulnerability[];
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  summary: {
    critical: number;
    high: number;
    medium: number;
    low: number;
  };
}> {
  const lock = await readComposerLock(repoPath);
  const vulnerabilities: SecurityVulnerability[] = [];
  
  if (!lock) {
    return {
      vulnerabilities: [],
      riskLevel: 'low',
      summary: { critical: 0, high: 0, medium: 0, low: 0 },
    };
  }
  
  const allPackages = [...lock.packages];
  if (lock['packages-dev']) {
    allPackages.push(...lock['packages-dev']);
  }
  
  for (const pkg of allPackages) {
    if (pkg.version.includes('dev') && !pkg.version.includes('dev-')) {
      vulnerabilities.push({
        package: pkg.name,
        version: pkg.version,
        severity: 'medium',
        description: 'Using development version in production',
        recommendation: 'Pin to a stable release version',
      });
    }
    
    if (pkg.version.startsWith('0.')) {
      vulnerabilities.push({
        package: pkg.name,
        version: pkg.version,
        severity: 'low',
        description: 'Using pre-1.0 version (potentially unstable)',
        recommendation: 'Consider upgrading to a stable 1.x+ version if available',
      });
    }
    
    if (!pkg.time || new Date(pkg.time) < new Date('2020-01-01')) {
      vulnerabilities.push({
        package: pkg.name,
        version: pkg.version,
        severity: 'medium',
        description: 'Package has not been updated in over 5 years',
        recommendation: 'Check for maintained alternatives or security advisories',
      });
    }
  }
  
  const summary = {
    critical: vulnerabilities.filter(v => v.severity === 'critical').length,
    high: vulnerabilities.filter(v => v.severity === 'high').length,
    medium: vulnerabilities.filter(v => v.severity === 'medium').length,
    low: vulnerabilities.filter(v => v.severity === 'low').length,
  };
  
  let riskLevel: 'low' | 'medium' | 'high' | 'critical' = 'low';
  if (summary.critical > 0) riskLevel = 'critical';
  else if (summary.high > 0) riskLevel = 'high';
  else if (summary.medium > 0) riskLevel = 'medium';
  
  return {
    vulnerabilities,
    riskLevel,
    summary,
  };
}
