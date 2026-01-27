import * as path from 'path';
import { PSR4Mapping } from '../types/index.js';
import { readComposerJson, expandPathArray, normalizePath } from '../utils/composer-utils.js';
import { findPhpFiles } from '../utils/file-scanner.js';
import { extractNamespaces } from '../utils/php-parser.js';

export async function analyzePSR4Autoloading(repoPath: string): Promise<{
  mappings: PSR4Mapping[];
  violations: Array<{
    file: string;
    expectedNamespace: string;
    actualNamespace?: string;
    issue: string;
  }>;
  stats: {
    totalMappings: number;
    totalFiles: number;
    validFiles: number;
    violationCount: number;
  };
}> {
  const composer = await readComposerJson(repoPath);
  const mappings: PSR4Mapping[] = [];
  
  if (composer.autoload?.['psr-4']) {
    for (const [namespace, paths] of Object.entries(composer.autoload['psr-4'])) {
      mappings.push({
        namespace,
        paths: expandPathArray(paths),
        type: 'psr-4',
        isDev: false,
      });
    }
  }
  
  if (composer['autoload-dev']?.['psr-4']) {
    for (const [namespace, paths] of Object.entries(composer['autoload-dev']['psr-4'])) {
      mappings.push({
        namespace,
        paths: expandPathArray(paths),
        type: 'psr-4',
        isDev: true,
      });
    }
  }
  
  const violations: Array<any> = [];
  let totalFiles = 0;
  let validFiles = 0;
  
  for (const mapping of mappings) {
    for (const relativePath of mapping.paths) {
      const absolutePath = normalizePath(repoPath, relativePath);
      const phpFiles = await findPhpFiles(absolutePath);
      
      for (const file of phpFiles) {
        totalFiles++;
        const fileInfo = await extractNamespaces(file);
        const relativeToPsrRoot = path.relative(absolutePath, file);
        const expectedNamespace = calculateExpectedNamespace(
          mapping.namespace,
          relativeToPsrRoot
        );
        
        if (fileInfo.namespace === expectedNamespace) {
          validFiles++;
        } else {
          violations.push({
            file: path.relative(repoPath, file),
            expectedNamespace,
            actualNamespace: fileInfo.namespace,
            issue: fileInfo.namespace
              ? 'Namespace mismatch'
              : 'Missing namespace declaration',
          });
        }
      }
    }
  }
  
  return {
    mappings,
    violations,
    stats: {
      totalMappings: mappings.length,
      totalFiles,
      validFiles,
      violationCount: violations.length,
    },
  };
}

function calculateExpectedNamespace(
  baseNamespace: string,
  filePath: string
): string {
  const withoutExtension = filePath.replace(/\.php$/, '');
  const parts = withoutExtension.split(path.sep);
  const dirParts = parts.slice(0, -1);
  
  let namespace = baseNamespace.replace(/\\+$/, '');
  
  if (dirParts.length > 0) {
    const subNamespace = dirParts.join('\\');
    namespace = `${namespace}\\${subNamespace}`;
  }
  
  return namespace;
}
