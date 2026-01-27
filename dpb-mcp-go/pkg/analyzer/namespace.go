package analyzer

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/kedson/dpb-mcp/pkg/types"
)

var (
	classRegex     = regexp.MustCompile(`(?:abstract\s+)?class\s+(\w+)`)
	interfaceRegex = regexp.MustCompile(`interface\s+(\w+)`)
	traitRegex     = regexp.MustCompile(`trait\s+(\w+)`)
	useRegex       = regexp.MustCompile(`use\s+([\w\\]+)(?:\s+as\s+\w+)?;`)
)

// NamespaceDetectionResult represents namespace detection output
type NamespaceDetectionResult struct {
	Namespaces            []types.NamespaceInfo `json:"namespaces"`
	TotalFiles            int                   `json:"totalFiles"`
	FilesWithoutNamespace []string              `json:"filesWithoutNamespace"`
}

// DetectNamespaces detects all namespaces in a repository
func DetectNamespaces(repoPath string) (string, error) {
	phpFiles, err := findPHPFiles(repoPath)
	if err != nil {
		return "", err
	}

	namespaceMap := make(map[string]*types.NamespaceInfo)
	filesWithout := make([]string, 0)
	
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Process files concurrently
	for _, file := range phpFiles {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()

			info, err := analyzeFile(f)
			if err != nil {
				return
			}

			relativePath, _ := filepath.Rel(repoPath, f)

			mu.Lock()
			defer mu.Unlock()

			if info.Namespace != "" {
				if _, exists := namespaceMap[info.Namespace]; !exists {
					namespaceMap[info.Namespace] = &types.NamespaceInfo{
						Namespace:  info.Namespace,
						Files:      make([]string, 0),
						Classes:    make([]string, 0),
						Interfaces: make([]string, 0),
						Traits:     make([]string, 0),
					}
				}

				ns := namespaceMap[info.Namespace]
				ns.Files = append(ns.Files, relativePath)
				ns.Classes = append(ns.Classes, info.Classes...)
				ns.Interfaces = append(ns.Interfaces, info.Interfaces...)
				ns.Traits = append(ns.Traits, info.Traits...)
			} else {
				filesWithout = append(filesWithout, relativePath)
			}
		}(file)
	}

	wg.Wait()

	namespaces := make([]types.NamespaceInfo, 0, len(namespaceMap))
	for _, info := range namespaceMap {
		namespaces = append(namespaces, *info)
	}

	result := NamespaceDetectionResult{
		Namespaces:            namespaces,
		TotalFiles:            len(phpFiles),
		FilesWithoutNamespace: filesWithout,
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// fileInfo represents PHP file analysis
type fileInfo struct {
	Namespace  string
	Classes    []string
	Interfaces []string
	Traits     []string
	Uses       []string
}

// analyzeFile extracts namespace and definitions from a PHP file
func analyzeFile(filePath string) (*fileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info := &fileInfo{
		Classes:    make([]string, 0),
		Interfaces: make([]string, 0),
		Traits:     make([]string, 0),
		Uses:       make([]string, 0),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Extract namespace
		if matches := namespaceRegex.FindStringSubmatch(line); matches != nil {
			info.Namespace = matches[1]
		}

		// Extract classes
		if matches := classRegex.FindStringSubmatch(line); matches != nil {
			info.Classes = append(info.Classes, matches[1])
		}

		// Extract interfaces
		if matches := interfaceRegex.FindStringSubmatch(line); matches != nil {
			info.Interfaces = append(info.Interfaces, matches[1])
		}

		// Extract traits
		if matches := traitRegex.FindStringSubmatch(line); matches != nil {
			info.Traits = append(info.Traits, matches[1])
		}

		// Extract use statements
		if matches := useRegex.FindStringSubmatch(line); matches != nil {
			info.Uses = append(info.Uses, matches[1])
		}
	}

	return info, nil
}

// AnalyzeNamespaceUsage analyzes usage of a specific namespace
func AnalyzeNamespaceUsage(repoPath, targetNamespace string) (string, error) {
	phpFiles, err := findPHPFiles(repoPath)
	if err != nil {
		return "", err
	}

	definedIn := make([]string, 0)
	importedBy := make([]struct {
		File    string   `json:"file"`
		Imports []string `json:"imports"`
	}, 0)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, file := range phpFiles {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()

			info, err := analyzeFile(f)
			if err != nil {
				return
			}

			relativePath, _ := filepath.Rel(repoPath, f)

			mu.Lock()
			defer mu.Unlock()

			if info.Namespace == targetNamespace {
				definedIn = append(definedIn, relativePath)
			}

			relevantImports := make([]string, 0)
			for _, use := range info.Uses {
				if strings.HasPrefix(use, targetNamespace) {
					relevantImports = append(relevantImports, use)
				}
			}

			if len(relevantImports) > 0 {
				importedBy = append(importedBy, struct {
					File    string   `json:"file"`
					Imports []string `json:"imports"`
				}{
					File:    relativePath,
					Imports: relevantImports,
				})
			}
		}(file)
	}

	wg.Wait()

	result := map[string]interface{}{
		"definedIn":   definedIn,
		"importedBy":  importedBy,
		"totalUsages": len(definedIn) + len(importedBy),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
