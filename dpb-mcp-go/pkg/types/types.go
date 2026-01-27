package types

// ComposerJSON represents the composer.json structure
type ComposerJSON struct {
	Name        string                       `json:"name,omitempty"`
	Description string                       `json:"description,omitempty"`
	Type        string                       `json:"type,omitempty"`
	License     interface{}                  `json:"license,omitempty"` // string or []string
	Require     map[string]string            `json:"require,omitempty"`
	RequireDev  map[string]string            `json:"require-dev,omitempty"`
	Autoload    *AutoloadConfig              `json:"autoload,omitempty"`
	AutoloadDev *AutoloadConfig              `json:"autoload-dev,omitempty"`
	Scripts     map[string]interface{}       `json:"scripts,omitempty"`
	Config      map[string]interface{}       `json:"config,omitempty"`
}

// AutoloadConfig represents autoload configuration
type AutoloadConfig struct {
	PSR4      map[string]interface{} `json:"psr-4,omitempty"` // string or []string
	PSR0      map[string]interface{} `json:"psr-0,omitempty"`
	Files     []string               `json:"files,omitempty"`
	Classmap  []string               `json:"classmap,omitempty"`
}

// ComposerLock represents the composer.lock structure
type ComposerLock struct {
	Packages        []PackageInfo `json:"packages"`
	PackagesDev     []PackageInfo `json:"packages-dev,omitempty"`
	ContentHash     string        `json:"content-hash,omitempty"`
	PluginAPIVersion string       `json:"plugin-api-version,omitempty"`
}

// PackageInfo represents a package in composer.lock
type PackageInfo struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	License     []string               `json:"license,omitempty"`
	Authors     []Author               `json:"authors,omitempty"`
	Require     map[string]string      `json:"require,omitempty"`
	RequireDev  map[string]string      `json:"require-dev,omitempty"`
	Autoload    *AutoloadConfig        `json:"autoload,omitempty"`
	Homepage    string                 `json:"homepage,omitempty"`
	Source      *SourceInfo            `json:"source,omitempty"`
	Dist        *DistInfo              `json:"dist,omitempty"`
	Time        string                 `json:"time,omitempty"`
}

// Author represents a package author
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

// SourceInfo represents source information
type SourceInfo struct {
	Type      string `json:"type"`
	URL       string `json:"url"`
	Reference string `json:"reference"`
}

// DistInfo represents distribution information
type DistInfo struct {
	Type      string `json:"type"`
	URL       string `json:"url"`
	Reference string `json:"reference"`
}

// PSR4Mapping represents a PSR-4 namespace mapping
type PSR4Mapping struct {
	Namespace string   `json:"namespace"`
	Paths     []string `json:"paths"`
	Type      string   `json:"type"` // "psr-4" or "psr-0"
	IsDev     bool     `json:"isDev"`
}

// PSR4Violation represents a PSR-4 compliance violation
type PSR4Violation struct {
	File              string  `json:"file"`
	ExpectedNamespace string  `json:"expectedNamespace"`
	ActualNamespace   *string `json:"actualNamespace,omitempty"`
	Issue             string  `json:"issue"`
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Type         string   `json:"type"` // "production" or "development"
	Dependencies []string `json:"dependencies"`
	UsedBy       []string `json:"usedBy"`
	License      string   `json:"license,omitempty"`
}

// SecurityVulnerability represents a security issue
type SecurityVulnerability struct {
	Package        string `json:"package"`
	Version        string `json:"version"`
	Severity       string `json:"severity"` // "low", "medium", "high", "critical"
	CVE            string `json:"cve,omitempty"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// NamespaceInfo represents namespace usage information
type NamespaceInfo struct {
	Namespace  string   `json:"namespace"`
	Files      []string `json:"files"`
	Classes    []string `json:"classes"`
	Interfaces []string `json:"interfaces"`
	Traits     []string `json:"traits"`
}

// RepoConfig represents a repository configuration
type RepoConfig struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "service", "library", "application"
	Team        string `json:"team,omitempty"`
	Description string `json:"description,omitempty"`
}

// MultiRepoAnalysis represents multi-repository analysis results
type MultiRepoAnalysis struct {
	Repositories       []RepoConfig               `json:"repositories"`
	SharedDependencies map[string][]string        `json:"sharedDependencies"`
	VersionConflicts   []VersionConflict          `json:"versionConflicts"`
	TotalPackages      int                        `json:"totalPackages"`
	CommonLicenses     map[string]int             `json:"commonLicenses"`
}

// VersionConflict represents a version conflict across repositories
type VersionConflict struct {
	Package  string          `json:"package"`
	Versions []RepoVersion   `json:"versions"`
}

// RepoVersion represents a version used by a repository
type RepoVersion struct {
	Repo    string `json:"repo"`
	Version string `json:"version"`
}

// LicenseDistribution represents license analysis results
type LicenseDistribution struct {
	License   string   `json:"license"`
	Count     int      `json:"count"`
	Packages  []string `json:"packages"`
	RiskLevel string   `json:"riskLevel"` // "safe", "caution", "review-required"
}
