package analyzer

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/faithfm/php-dependency-mcp/pkg/composer"
	"github.com/faithfm/php-dependency-mcp/pkg/types"
	"golang.org/x/sync/errgroup"
)

// PSR4AnalysisResult represents PSR-4 analysis output
type PSR4AnalysisResult struct {
	Mappings   []types.PSR4Mapping   `json:"mappings"`
	Violations []types.PSR4Violation `json:"violations"`
	Stats      PSR4Stats             `json:"stats"`
}

// PSR4Stats represents PSR-4 statistics
type PSR4Stats struct {
	TotalMappings   int `json:"totalMappings"`
	TotalFiles      int `json:"totalFiles"`
	ValidFiles      int `json:"validFiles"`
	ViolationCount  int `json:"violationCount"`
}

var namespaceRegex = regexp.MustCompile(`namespace\s+([\w\\]+)\s*;`)

// AnalyzePSR4Autoloading analyzes PSR-4 compliance
func AnalyzePSR4Autoloading(repoPath string) (string, error) {
	composerJSON, err := composer.ReadComposerJSON(repoPath)
	if err != nil {
		return "", err
	}

	mappings := composer.GetPSR4Mappings(composerJSON)

	violations := make([]types.PSR4Violation, 0)
	var violationsMu sync.Mutex
	
	totalFiles := 0
	validFiles := 0
	var statsMu sync.Mutex

	// Process each mapping concurrently
	var g errgroup.Group

	for _, mapping := range mappings {
		for _, relativePath := range mapping.Paths {
			absPath := filepath.Join(repoPath, relativePath)
			mappingCopy := mapping

			g.Go(func() error {
				phpFiles, err := findPHPFiles(absPath)
				if err != nil {
					return nil // Skip if directory doesn't exist
				}

				// Process files concurrently
				var fileWg sync.WaitGroup
				for _, file := range phpFiles {
					fileWg.Add(1)
					go func(f string) {
						defer fileWg.Done()

						statsMu.Lock()
						totalFiles++
						statsMu.Unlock()

						namespace, err := extractNamespace(f)
						if err != nil {
							return
						}

						relToRoot, _ := filepath.Rel(absPath, f)
						expectedNS := composer.CalculateExpectedNamespace(mappingCopy.Namespace, relToRoot)

						if namespace == expectedNS {
							statsMu.Lock()
							validFiles++
							statsMu.Unlock()
						} else {
							issue := "Namespace mismatch"
							if namespace == "" {
								issue = "Missing namespace declaration"
							}

							violationsMu.Lock()
							violations = append(violations, types.PSR4Violation{
								File:              filepath.Join(relativePath, relToRoot),
								ExpectedNamespace: expectedNS,
								ActualNamespace:   &namespace,
								Issue:             issue,
							})
							violationsMu.Unlock()
						}
					}(file)
				}
				fileWg.Wait()
				return nil
			})
		}
	}

	g.Wait()

	result := PSR4AnalysisResult{
		Mappings:   mappings,
		Violations: violations,
		Stats: PSR4Stats{
			TotalMappings:  len(mappings),
			TotalFiles:     totalFiles,
			ValidFiles:     validFiles,
			ViolationCount: len(violations),
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// findPHPFiles finds all PHP files in a directory
func findPHPFiles(dir string) ([]string, error) {
	files := make([]string, 0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
		}

		if !info.IsDir() && strings.HasSuffix(path, ".php") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// extractNamespace extracts namespace from a PHP file
func extractNamespace(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := namespaceRegex.FindStringSubmatch(line); matches != nil {
			return matches[1], nil
		}
	}

	return "", nil
}
