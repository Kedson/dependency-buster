import { readComposerLock } from '../utils/composer-utils.js';

export async function generateDependencyGraph(
  repoPath: string,
  options: {
    maxDepth?: number;
    includeDevDeps?: boolean;
    focusPackage?: string;
  } = {}
): Promise<string> {
  const { maxDepth = 2, includeDevDeps = false, focusPackage } = options;
  const lock = await readComposerLock(repoPath);
  
  if (!lock) {
    return 'graph TD\n  NoLock[composer.lock not found]';
  }
  
  let mermaid = 'graph TD\n';
  const packages = [...lock.packages];
  
  if (includeDevDeps && lock['packages-dev']) {
    packages.push(...lock['packages-dev']);
  }
  
  let relevantPackages = packages;
  if (focusPackage) {
    const visited = new Set<string>();
    const toVisit = [focusPackage];
    let depth = 0;
    
    while (toVisit.length > 0 && depth < maxDepth) {
      const current = toVisit.shift()!;
      if (visited.has(current)) continue;
      visited.add(current);
      
      const pkg = packages.find(p => p.name === current);
      if (pkg && pkg.require) {
        for (const dep of Object.keys(pkg.require)) {
          if (!dep.startsWith('php') && !dep.startsWith('ext-')) {
            toVisit.push(dep);
          }
        }
      }
      depth++;
    }
    
    relevantPackages = packages.filter(p => visited.has(p.name));
  }
  
  if (focusPackage) {
    const sanitized = sanitizeForMermaid(focusPackage);
    mermaid += `  Root[${focusPackage}]\n`;
    
    const pkg = packages.find(p => p.name === focusPackage);
    if (pkg && pkg.require) {
      for (const [dep, version] of Object.entries(pkg.require)) {
        if (!dep.startsWith('php') && !dep.startsWith('ext-')) {
          const depSanitized = sanitizeForMermaid(dep);
          mermaid += `  Root --> ${depSanitized}["${dep}<br/>${version}"]\n`;
        }
      }
    }
  } else {
    mermaid += `  Root[Your Application]\n`;
    
    const topLevel = relevantPackages.slice(0, 15);
    
    for (const pkg of topLevel) {
      const sanitized = sanitizeForMermaid(pkg.name);
      mermaid += `  Root --> ${sanitized}["${pkg.name}<br/>${pkg.version}"]\n`;
      
      if (maxDepth > 1 && pkg.require) {
        const deps = Object.entries(pkg.require)
          .filter(([dep]) => !dep.startsWith('php') && !dep.startsWith('ext-'))
          .slice(0, 3);
        
        for (const [dep, version] of deps) {
          const depSanitized = sanitizeForMermaid(dep);
          mermaid += `  ${sanitized} --> ${depSanitized}["${dep}<br/>${version}"]\n`;
        }
      }
    }
  }
  
  return mermaid;
}

function sanitizeForMermaid(name: string): string {
  return name.replace(/[\/\-\.@]/g, '_');
}

export async function generateNamespaceGraph(
  repoPath: string,
  namespaces: Array<{ namespace: string; files: string[] }>
): Promise<string> {
  let mermaid = 'graph TD\n';
  mermaid += '  Root[Application Root]\n';
  
  const topLevel = new Map<string, string[]>();
  
  for (const ns of namespaces) {
    const parts = ns.namespace.split('\\');
    const top = parts[0];
    
    if (!topLevel.has(top)) {
      topLevel.set(top, []);
    }
    topLevel.get(top)!.push(ns.namespace);
  }
  
  for (const [top, children] of topLevel) {
    const topSanitized = sanitizeForMermaid(top);
    mermaid += `  Root --> ${topSanitized}["${top}<br/>${children.length} namespaces"]\n`;
    
    const samples = children.slice(0, 3);
    for (const child of samples) {
      if (child !== top) {
        const childSanitized = sanitizeForMermaid(child);
        mermaid += `  ${topSanitized} --> ${childSanitized}["${child}"]\n`;
      }
    }
  }
  
  return mermaid;
}
