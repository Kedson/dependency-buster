import { RepoConfig, MultiRepoAnalysis, ComposerJson } from '../types/index.js';
import { readComposerJson } from '../utils/composer-utils.js';

export async function analyzeMultipleRepositories(
  repos: RepoConfig[]
): Promise<MultiRepoAnalysis> {
  const repoData = new Map<string, {
    composer: ComposerJson;
    dependencies: Record<string, string>;
  }>();
  
  for (const repo of repos) {
    try {
      const composer = await readComposerJson(repo.path);
      const dependencies = { ...composer.require, ...composer['require-dev'] };
      delete dependencies.php;
      
      repoData.set(repo.name, { composer, dependencies });
    } catch (error) {
      console.error(`Failed to load ${repo.name}:`, error);
    }
  }
  
  const packageUsage = new Map<string, string[]>();
  
  for (const [repoName, data] of repoData) {
    for (const pkg of Object.keys(data.dependencies)) {
      if (!packageUsage.has(pkg)) {
        packageUsage.set(pkg, []);
      }
      packageUsage.get(pkg)!.push(repoName);
    }
  }
  
  const sharedDependencies: Record<string, string[]> = {};
  for (const [pkg, repos] of packageUsage) {
    if (repos.length > 1) {
      sharedDependencies[pkg] = repos;
    }
  }
  
  const versionConflicts: MultiRepoAnalysis['versionConflicts'] = [];
  
  for (const [pkg, usedByRepos] of packageUsage) {
    const versions = new Map<string, string[]>();
    
    for (const repoName of usedByRepos) {
      const version = repoData.get(repoName)!.dependencies[pkg];
      if (!versions.has(version)) {
        versions.set(version, []);
      }
      versions.get(version)!.push(repoName);
    }
    
    if (versions.size > 1) {
      const versionArray = Array.from(versions.entries())
        .map(([version, repos]) => repos.map(repo => ({ repo, version })))
        .flat();
      
      versionConflicts.push({
        package: pkg,
        versions: versionArray,
      });
    }
  }
  
  const licenseCount: Record<string, number> = {};
  for (const data of repoData.values()) {
    const license = data.composer.license;
    if (license) {
      const licenses = Array.isArray(license) ? license : [license];
      for (const lic of licenses) {
        licenseCount[lic] = (licenseCount[lic] || 0) + 1;
      }
    }
  }
  
  const allPackages = new Set<string>();
  for (const data of repoData.values()) {
    for (const pkg of Object.keys(data.dependencies)) {
      allPackages.add(pkg);
    }
  }
  
  return {
    repositories: repos,
    sharedDependencies,
    versionConflicts,
    totalPackages: allPackages.size,
    commonLicenses: licenseCount,
  };
}

export async function generateConsolidatedReport(
  repos: RepoConfig[]
): Promise<string> {
  const analysis = await analyzeMultipleRepositories(repos);
  
  let report = `# Multi-Repository Dependency Analysis\n\n`;
  report += `**Generated:** ${new Date().toISOString()}\n\n`;
  report += `## Repositories Analyzed\n\n`;
  
  for (const repo of analysis.repositories) {
    report += `- **${repo.name}** (${repo.type})`;
    if (repo.team) report += ` - Team: ${repo.team}`;
    if (repo.description) report += `\n  ${repo.description}`;
    report += `\n`;
  }
  
  report += `\n## Summary\n\n`;
  report += `- Total unique packages: ${analysis.totalPackages}\n`;
  report += `- Shared dependencies: ${Object.keys(analysis.sharedDependencies).length}\n`;
  report += `- Version conflicts: ${analysis.versionConflicts.length}\n\n`;
  
  if (Object.keys(analysis.sharedDependencies).length > 0) {
    report += `## Shared Dependencies\n\n`;
    report += `| Package | Used By |\n`;
    report += `|---------|----------|\n`;
    
    for (const [pkg, repos] of Object.entries(analysis.sharedDependencies)) {
      report += `| ${pkg} | ${repos.join(', ')} |\n`;
    }
    report += `\n`;
  }
  
  if (analysis.versionConflicts.length > 0) {
    report += `## ⚠️ Version Conflicts\n\n`;
    
    for (const conflict of analysis.versionConflicts) {
      report += `### ${conflict.package}\n\n`;
      for (const version of conflict.versions) {
        report += `- **${version.repo}**: ${version.version}\n`;
      }
      report += `\n`;
    }
  }
  
  if (Object.keys(analysis.commonLicenses).length > 0) {
    report += `## License Distribution\n\n`;
    report += `| License | Count |\n`;
    report += `|---------|-------|\n`;
    
    const sorted = Object.entries(analysis.commonLicenses)
      .sort((a, b) => b[1] - a[1]);
    
    for (const [license, count] of sorted) {
      report += `| ${license} | ${count} |\n`;
    }
  }
  
  return report;
}
