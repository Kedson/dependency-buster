export interface ComposerJson {
  name?: string;
  description?: string;
  type?: string;
  license?: string | string[];
  require?: Record<string, string>;
  'require-dev'?: Record<string, string>;
  autoload?: {
    'psr-4'?: Record<string, string | string[]>;
    'psr-0'?: Record<string, string | string[]>;
    files?: string[];
    classmap?: string[];
  };
  'autoload-dev'?: {
    'psr-4'?: Record<string, string | string[]>;
  };
  scripts?: Record<string, string | string[]>;
  config?: Record<string, any>;
}

export interface ComposerLock {
  packages: PackageInfo[];
  'packages-dev'?: PackageInfo[];
  'content-hash'?: string;
  'plugin-api-version'?: string;
}

export interface PackageInfo {
  name: string;
  version: string;
  description?: string;
  type?: string;
  license?: string[];
  authors?: Array<{ name: string; email?: string }>;
  require?: Record<string, string>;
  'require-dev'?: Record<string, string>;
  autoload?: ComposerJson['autoload'];
  homepage?: string;
  source?: {
    type: string;
    url: string;
    reference: string;
  };
  dist?: {
    type: string;
    url: string;
    reference: string;
  };
  time?: string;
}

export interface PSR4Mapping {
  namespace: string;
  paths: string[];
  type: 'psr-4' | 'psr-0';
  isDev: boolean;
}

export interface NamespaceUsage {
  namespace: string;
  file: string;
  line: number;
  type: 'class' | 'interface' | 'trait' | 'function' | 'use';
}

export interface DependencyNode {
  name: string;
  version: string;
  type: 'production' | 'development';
  dependencies: string[];
  usedBy: string[];
  license?: string;
}

export interface SecurityVulnerability {
  package: string;
  version: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  cve?: string;
  description: string;
  recommendation: string;
}

export interface RepoConfig {
  name: string;
  path: string;
  type: 'service' | 'library' | 'application';
  team?: string;
  description?: string;
}

export interface MultiRepoAnalysis {
  repositories: RepoConfig[];
  sharedDependencies: Record<string, string[]>;
  versionConflicts: Array<{
    package: string;
    versions: Array<{ repo: string; version: string }>;
  }>;
  totalPackages: number;
  commonLicenses: Record<string, number>;
}
