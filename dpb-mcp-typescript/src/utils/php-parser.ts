import * as fs from 'fs/promises';
import * as path from 'path';
import { NamespaceUsage } from '../types/index.js';

export async function extractNamespaces(
  filePath: string
): Promise<{
  namespace?: string;
  uses: string[];
  classes: string[];
  interfaces: string[];
  traits: string[];
}> {
  const content = await fs.readFile(filePath, 'utf8');
  
  // Extract namespace
  const namespaceMatch = content.match(/namespace\s+([\w\\]+)\s*;/);
  const namespace = namespaceMatch ? namespaceMatch[1] : undefined;
  
  // Extract use statements
  const useMatches = content.matchAll(/use\s+([\w\\]+)(?:\s+as\s+\w+)?;/g);
  const uses = Array.from(useMatches).map(match => match[1]);
  
  // Extract class definitions
  const classMatches = content.matchAll(/(?:abstract\s+)?class\s+(\w+)/g);
  const classes = Array.from(classMatches).map(match => match[1]);
  
  // Extract interface definitions
  const interfaceMatches = content.matchAll(/interface\s+(\w+)/g);
  const interfaces = Array.from(interfaceMatches).map(match => match[1]);
  
  // Extract trait definitions
  const traitMatches = content.matchAll(/trait\s+(\w+)/g);
  const traits = Array.from(traitMatches).map(match => match[1]);
  
  return { namespace, uses, classes, interfaces, traits };
}

export async function findNamespaceUsages(
  repoPath: string,
  targetNamespace: string
): Promise<NamespaceUsage[]> {
  const usages: NamespaceUsage[] = [];
  
  async function scanDirectory(dir: string): Promise<void> {
    const entries = await fs.readdir(dir, { withFileTypes: true });
    
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name);
      
      if (entry.isDirectory()) {
        if (!entry.name.startsWith('.') && entry.name !== 'vendor') {
          await scanDirectory(fullPath);
        }
      } else if (entry.name.endsWith('.php')) {
        const content = await fs.readFile(fullPath, 'utf8');
        const lines = content.split('\n');
        
        lines.forEach((line, index) => {
          if (line.includes(targetNamespace)) {
            usages.push({
              namespace: targetNamespace,
              file: path.relative(repoPath, fullPath),
              line: index + 1,
              type: line.includes('use ') ? 'use' : 'class',
            });
          }
        });
      }
    }
  }
  
  await scanDirectory(repoPath);
  return usages;
}

export function extractFullyQualifiedClassName(
  namespace: string | undefined,
  className: string
): string {
  if (!namespace) return className;
  return `${namespace}\\${className}`;
}
