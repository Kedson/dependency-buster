package analyzer

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/faithfm/php-dependency-mcp/pkg/composer"
	"github.com/faithfm/php-dependency-mcp/pkg/types"
)

// DependencyAnalysisResult represents dependency analysis output
type DependencyAnalysisResult struct {
	Production map[string]string        `json:"production"`
	Development map[string]string       `json:"development"`
	Tree       []types.DependencyNode   `json:"tree"`
	Stats      DependencyStats          `json:"stats"`
}

// DependencyStats represents dependency statistics
type DependencyStats struct {
	TotalProduction  int `json:"totalProduction"`
	TotalDevelopment int `json:"totalDevelopment"`
	Outdated         int `json:"outdated"`
	UpToDate         int `json:"upToDate"`
}

// AnalyzeDependenciesRaw performs comprehensive dependency analysis and returns struct
func AnalyzeDependenciesRaw(repoPath string) (*DependencyAnalysisResult, error) {
	composerJSON, err := composer.ReadComposerJSON(repoPath)
	if err != nil {
		return nil, err
	}

	lock, err := composer.ReadComposerLock(repoPath)
	if err != nil {
		lock = nil // It's okay if lock doesn't exist
	}

	production := make(map[string]string)
	if composerJSON.Require != nil {
		production = composer.FilterPHPDependencies(composerJSON.Require)
	}

	development := make(map[string]string)
	if composerJSON.RequireDev != nil {
		development = composerJSON.RequireDev
	}

	tree := make([]types.DependencyNode, 0)

	if lock != nil {
		// Build dependency tree with concurrency
		tree = buildDependencyTree(lock)
	}

	result := &DependencyAnalysisResult{
		Production:  production,
		Development: development,
		Tree:        tree,
		Stats: DependencyStats{
			TotalProduction:  len(production),
			TotalDevelopment: len(development),
			Outdated:         0,
			UpToDate:         0,
		},
	}

	return result, nil
}

// AnalyzeDependencies performs comprehensive dependency analysis (returns JSON string)
func AnalyzeDependencies(repoPath string) (string, error) {
	result, err := AnalyzeDependenciesRaw(repoPath)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// buildDependencyTree builds the dependency tree using goroutines
func buildDependencyTree(lock *types.ComposerLock) []types.DependencyNode {
	allPackages := append([]types.PackageInfo{}, lock.Packages...)
	if lock.PackagesDev != nil {
		allPackages = append(allPackages, lock.PackagesDev...)
	}

	// Concurrent tree building
	tree := make([]types.DependencyNode, len(allPackages))
	var wg sync.WaitGroup
	
	for i, pkg := range allPackages {
		wg.Add(1)
		go func(index int, p types.PackageInfo) {
			defer wg.Done()
			
			pkgType := "production"
			if index >= len(lock.Packages) {
				pkgType = "development"
			}

			deps := make([]string, 0)
			for dep := range p.Require {
				if !strings.HasPrefix(dep, "php") && !strings.HasPrefix(dep, "ext-") {
					deps = append(deps, dep)
				}
			}

			license := ""
			if len(p.License) > 0 {
				license = p.License[0]
			}

			tree[index] = types.DependencyNode{
				Name:         p.Name,
				Version:      p.Version,
				Type:         pkgType,
				Dependencies: deps,
				UsedBy:       make([]string, 0),
				License:      license,
			}
		}(i, pkg)
	}

	wg.Wait()

	// Calculate reverse dependencies
	for i := range tree {
		for j := range tree {
			if i != j {
				for _, dep := range tree[j].Dependencies {
					if tree[i].Name == dep {
						tree[i].UsedBy = append(tree[i].UsedBy, tree[j].Name)
					}
				}
			}
		}
	}

	return tree
}

// FindCircularDependencies detects circular dependency chains
func FindCircularDependencies(repoPath string) (string, error) {
	lock, err := composer.ReadComposerLock(repoPath)
	if err != nil {
		return "", err
	}

	tree := buildDependencyTree(lock)
	
	cycles := make([][]string, 0)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(pkgName string, path []string)
	dfs = func(pkgName string, path []string) {
		visited[pkgName] = true
		recStack[pkgName] = true
		path = append(path, pkgName)

		// Find the node
		var node *types.DependencyNode
		for i := range tree {
			if tree[i].Name == pkgName {
				node = &tree[i]
				break
			}
		}

		if node != nil {
			for _, dep := range node.Dependencies {
				if !visited[dep] {
					dfs(dep, path)
				} else if recStack[dep] {
					// Found a cycle
					cycleStart := -1
					for i, p := range path {
						if p == dep {
							cycleStart = i
							break
						}
					}
					if cycleStart >= 0 {
						cycle := append(path[cycleStart:], dep)
						cycles = append(cycles, cycle)
					}
				}
			}
		}

		recStack[pkgName] = false
	}

	for _, node := range tree {
		if !visited[node.Name] {
			dfs(node.Name, []string{})
		}
	}

	result := map[string]interface{}{
		"cycles": cycles,
		"count":  len(cycles),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
