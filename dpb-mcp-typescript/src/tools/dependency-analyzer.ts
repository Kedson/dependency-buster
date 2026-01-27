import { readComposerJson, readComposerLock } from '../utils/composer-utils.js';
import { DependencyNode } from '../types/index.js';

export async function analyzeDependencies(repoPath: string): Promise<{
  production: Record<string, string>;
  development: Record<string, string>;
  tree: DependencyNode[];
  stats: {
    totalProduction: number;
    totalDevelopment: number;
    outdated: number;
    upToDate: number;
  };
}> {
  const composer = await readComposerJson(repoPath);
  const lock = await readComposerLock(repoPath);
  
  const production = { ...composer.require };
  delete production.php;
  
  const development = { ...composer['require-dev'] };
  
  const tree: DependencyNode[] = [];
  
  if (lock) {
    for (const pkg of lock.packages) {
      const node: DependencyNode = {
        name: pkg.name,
        version: pkg.version,
        type: 'production',
        dependencies: Object.keys(pkg.require || {}).filter(dep => !dep.startsWith('php')),
        usedBy: [],
        license: pkg.license?.[0],
      };
      tree.push(node);
    }
    
    if (lock['packages-dev']) {
      for (const pkg of lock['packages-dev']) {
        const node: DependencyNode = {
          name: pkg.name,
          version: pkg.version,
          type: 'development',
          dependencies: Object.keys(pkg.require || {}).filter(dep => !dep.startsWith('php')),
          usedBy: [],
          license: pkg.license?.[0],
        };
        tree.push(node);
      }
    }
    
    for (const node of tree) {
      for (const dep of node.dependencies) {
        const depNode = tree.find(n => n.name === dep);
        if (depNode) {
          depNode.usedBy.push(node.name);
        }
      }
    }
  }
  
  return {
    production,
    development,
    tree,
    stats: {
      totalProduction: Object.keys(production).length,
      totalDevelopment: Object.keys(development).length,
      outdated: 0,
      upToDate: 0,
    },
  };
}

export async function findCircularDependencies(
  repoPath: string
): Promise<string[][]> {
  const { tree } = await analyzeDependencies(repoPath);
  const cycles: string[][] = [];
  const visited = new Set<string>();
  const recursionStack = new Set<string>();
  
  function dfs(packageName: string, path: string[]): void {
    visited.add(packageName);
    recursionStack.add(packageName);
    path.push(packageName);
    
    const node = tree.find(n => n.name === packageName);
    if (node) {
      for (const dep of node.dependencies) {
        if (!visited.has(dep)) {
          dfs(dep, [...path]);
        } else if (recursionStack.has(dep)) {
          const cycleStart = path.indexOf(dep);
          const cycle = path.slice(cycleStart);
          cycle.push(dep);
          cycles.push(cycle);
        }
      }
    }
    
    recursionStack.delete(packageName);
  }
  
  for (const node of tree) {
    if (!visited.has(node.name)) {
      dfs(node.name, []);
    }
  }
  
  return cycles;
}
