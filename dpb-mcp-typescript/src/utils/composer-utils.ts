import * as fs from 'fs/promises';
import * as path from 'path';
import { ComposerJson, ComposerLock } from '../types/index.js';

export async function readComposerJson(repoPath: string): Promise<ComposerJson> {
  const composerPath = path.join(repoPath, 'composer.json');
  const content = await fs.readFile(composerPath, 'utf8');
  return JSON.parse(content);
}

export async function readComposerLock(repoPath: string): Promise<ComposerLock | null> {
  try {
    const lockPath = path.join(repoPath, 'composer.lock');
    const content = await fs.readFile(lockPath, 'utf8');
    return JSON.parse(content);
  } catch (error) {
    return null;
  }
}

export function normalizePath(basePath: string, relativePath: string): string {
  return path.resolve(basePath, relativePath);
}

export function expandPathArray(paths: string | string[]): string[] {
  return Array.isArray(paths) ? paths : [paths];
}
