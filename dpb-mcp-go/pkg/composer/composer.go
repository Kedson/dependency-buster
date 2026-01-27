package composer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kedson/dpb-mcp/pkg/types"
)

// ReadComposerJSON reads and parses composer.json
func ReadComposerJSON(repoPath string) (*types.ComposerJSON, error) {
	composerPath := filepath.Join(repoPath, "composer.json")
	data, err := os.ReadFile(composerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read composer.json: %w", err)
	}

	var composer types.ComposerJSON
	if err := json.Unmarshal(data, &composer); err != nil {
		return nil, fmt.Errorf("failed to parse composer.json: %w", err)
	}

	return &composer, nil
}

// ReadComposerLock reads and parses composer.lock
func ReadComposerLock(repoPath string) (*types.ComposerLock, error) {
	lockPath := filepath.Join(repoPath, "composer.lock")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err // Not an error if lock doesn't exist
	}

	var lock types.ComposerLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse composer.lock: %w", err)
	}

	return &lock, nil
}

// GetPSR4Mappings extracts PSR-4 mappings from composer.json
func GetPSR4Mappings(composer *types.ComposerJSON) []types.PSR4Mapping {
	mappings := make([]types.PSR4Mapping, 0)

	// Production autoload
	if composer.Autoload != nil && composer.Autoload.PSR4 != nil {
		for namespace, paths := range composer.Autoload.PSR4 {
			mappings = append(mappings, types.PSR4Mapping{
				Namespace: namespace,
				Paths:     normalizePaths(paths),
				Type:      "psr-4",
				IsDev:     false,
			})
		}
	}

	// Dev autoload
	if composer.AutoloadDev != nil && composer.AutoloadDev.PSR4 != nil {
		for namespace, paths := range composer.AutoloadDev.PSR4 {
			mappings = append(mappings, types.PSR4Mapping{
				Namespace: namespace,
				Paths:     normalizePaths(paths),
				Type:      "psr-4",
				IsDev:     true,
			})
		}
	}

	return mappings
}

// normalizePaths converts path(s) to string slice
func normalizePaths(paths interface{}) []string {
	switch v := paths.(type) {
	case string:
		return []string{v}
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, p := range v {
			if str, ok := p.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return []string{}
	}
}

// GetLicenses extracts licenses from composer.json
func GetLicenses(composer *types.ComposerJSON) []string {
	if composer.License == nil {
		return []string{}
	}

	switch v := composer.License.(type) {
	case string:
		return []string{v}
	case []interface{}:
		licenses := make([]string, 0, len(v))
		for _, l := range v {
			if str, ok := l.(string); ok {
				licenses = append(licenses, str)
			}
		}
		return licenses
	default:
		return []string{}
	}
}

// CalculateExpectedNamespace calculates the expected namespace for a file
func CalculateExpectedNamespace(baseNamespace, relativeFilePath string) string {
	// Remove .php extension
	withoutExt := strings.TrimSuffix(relativeFilePath, ".php")
	
	// Split by directory separator
	parts := strings.Split(filepath.ToSlash(withoutExt), "/")
	
	// Remove the filename (last part) to get directory structure
	dirParts := parts[:len(parts)-1]
	
	// Build expected namespace
	namespace := strings.TrimSuffix(baseNamespace, "\\")
	
	if len(dirParts) > 0 {
		subNamespace := strings.Join(dirParts, "\\")
		namespace = namespace + "\\" + subNamespace
	}
	
	return namespace
}

// FilterPHPDependencies removes PHP and extension dependencies
func FilterPHPDependencies(deps map[string]string) map[string]string {
	filtered := make(map[string]string)
	for name, version := range deps {
		if !strings.HasPrefix(name, "php") && !strings.HasPrefix(name, "ext-") {
			filtered[name] = version
		}
	}
	return filtered
}
