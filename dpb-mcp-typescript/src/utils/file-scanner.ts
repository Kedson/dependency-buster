import * as fs from 'fs/promises';
import * as path from 'path';
import { glob } from 'glob';

export async function findPhpFiles(repoPath: string): Promise<string[]> {
  const pattern = path.join(repoPath, '**/*.php');
  const files = await glob(pattern, {
    ignore: ['**/vendor/**', '**/node_modules/**', '**/.git/**'],
  });
  return files;
}

export async function fileExists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

export async function isDirectory(dirPath: string): Promise<boolean> {
  try {
    const stats = await fs.stat(dirPath);
    return stats.isDirectory();
  } catch {
    return false;
  }
}
