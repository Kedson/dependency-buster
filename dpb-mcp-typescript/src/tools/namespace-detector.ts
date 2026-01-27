import * as path from 'path';
import { findPhpFiles } from '../utils/file-scanner.js';
import { extractNamespaces } from '../utils/php-parser.js';

export async function detectNamespaces(repoPath: string): Promise<{
  namespaces: Array<{
    namespace: string;
    files: string[];
    classes: string[];
    interfaces: string[];
    traits: string[];
  }>;
  totalFiles: number;
  filesWithoutNamespace: string[];
}> {
  const phpFiles = await findPhpFiles(repoPath);
  const namespaceMap = new Map<
    string,
    {
      files: string[];
      classes: string[];
      interfaces: string[];
      traits: string[];
    }
  >();
  
  const filesWithoutNamespace: string[] = [];
  
  for (const file of phpFiles) {
    const info = await extractNamespaces(file);
    const relativePath = path.relative(repoPath, file);
    
    if (info.namespace) {
      if (!namespaceMap.has(info.namespace)) {
        namespaceMap.set(info.namespace, {
          files: [],
          classes: [],
          interfaces: [],
          traits: [],
        });
      }
      
      const entry = namespaceMap.get(info.namespace)!;
      entry.files.push(relativePath);
      entry.classes.push(...info.classes);
      entry.interfaces.push(...info.interfaces);
      entry.traits.push(...info.traits);
    } else {
      filesWithoutNamespace.push(relativePath);
    }
  }
  
  const namespaces = Array.from(namespaceMap.entries()).map(([namespace, data]) => ({
    namespace,
    ...data,
  }));
  
  return {
    namespaces,
    totalFiles: phpFiles.length,
    filesWithoutNamespace,
  };
}

export async function analyzeNamespaceUsage(
  repoPath: string,
  targetNamespace: string
): Promise<{
  definedIn: string[];
  importedBy: Array<{ file: string; imports: string[] }>;
  totalUsages: number;
}> {
  const phpFiles = await findPhpFiles(repoPath);
  const definedIn: string[] = [];
  const importedBy: Array<{ file: string; imports: string[] }> = [];
  
  for (const file of phpFiles) {
    const info = await extractNamespaces(file);
    const relativePath = path.relative(repoPath, file);
    
    if (info.namespace === targetNamespace) {
      definedIn.push(relativePath);
    }
    
    const relevantImports = info.uses.filter(use =>
      use.startsWith(targetNamespace)
    );
    
    if (relevantImports.length > 0) {
      importedBy.push({
        file: relativePath,
        imports: relevantImports,
      });
    }
  }
  
  const totalUsages = definedIn.length + importedBy.length;
  
  return {
    definedIn,
    importedBy,
    totalUsages,
  };
}
