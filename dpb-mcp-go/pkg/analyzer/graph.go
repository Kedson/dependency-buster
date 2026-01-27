package analyzer

import (
	"fmt"
	"strings"

	"github.com/kedson/dpb-mcp/pkg/composer"
	"github.com/kedson/dpb-mcp/pkg/types"
)

// GenerateDependencyGraph generates a Mermaid diagram of dependencies
func GenerateDependencyGraph(repoPath string, maxDepth int, includeDevDeps bool, focusPackage string) (string, error) {
	lock, err := composer.ReadComposerLock(repoPath)
	if err != nil {
		return "graph TD\n  NoLock[composer.lock not found]", nil
	}

	if maxDepth == 0 {
		maxDepth = 2
	}

	mermaid := "graph TD\n"
	
	packages := append([]types.PackageInfo{}, lock.Packages...)
	if includeDevDeps && lock.PackagesDev != nil {
		packages = append(packages, lock.PackagesDev...)
	}

	if focusPackage != "" {
		mermaid += generateFocusedGraph(packages, focusPackage, maxDepth)
	} else {
		mermaid += generateFullGraph(packages, maxDepth)
	}

	return mermaid, nil
}

// generateFullGraph generates a full dependency graph
func generateFullGraph(packages []types.PackageInfo, maxDepth int) string {
	var sb strings.Builder
	sb.WriteString("  Root[Your Application]\n")

	// Limit to first 15 packages to avoid overwhelming graph
	limit := 15
	if len(packages) < limit {
		limit = len(packages)
	}

	for i := 0; i < limit; i++ {
		pkg := packages[i]
		sanitized := sanitizeForMermaid(pkg.Name)
		sb.WriteString(fmt.Sprintf("  Root --> %s[\"%s<br/>%s\"]\n", sanitized, pkg.Name, pkg.Version))

		if maxDepth > 1 && pkg.Require != nil {
			depCount := 0
			for dep, version := range pkg.Require {
				if !strings.HasPrefix(dep, "php") && !strings.HasPrefix(dep, "ext-") {
					if depCount < 3 {
						depSanitized := sanitizeForMermaid(dep)
						sb.WriteString(fmt.Sprintf("  %s --> %s[\"%s<br/>%s\"]\n", 
							sanitized, depSanitized, dep, version))
						depCount++
					}
				}
			}
		}
	}

	return sb.String()
}

// generateFocusedGraph generates a graph focused on a specific package
func generateFocusedGraph(packages []types.PackageInfo, focusPackage string, _ int) string {
	var sb strings.Builder
	focusSanitized := sanitizeForMermaid(focusPackage)
	sb.WriteString(fmt.Sprintf("  %s[%s]\n", focusSanitized, focusPackage))

	// Find the focus package
	var focusPkg *types.PackageInfo
	for i := range packages {
		if packages[i].Name == focusPackage {
			focusPkg = &packages[i]
			break
		}
	}

	if focusPkg != nil && focusPkg.Require != nil {
		for dep, version := range focusPkg.Require {
			if !strings.HasPrefix(dep, "php") && !strings.HasPrefix(dep, "ext-") {
				depSanitized := sanitizeForMermaid(dep)
				sb.WriteString(fmt.Sprintf("  %s --> %s[\"%s<br/>%s\"]\n", 
					focusSanitized, depSanitized, dep, version))
			}
		}
	}

	return sb.String()
}

// sanitizeForMermaid sanitizes package names for Mermaid
func sanitizeForMermaid(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"-", "_",
		".", "_",
		"@", "_",
	)
	return replacer.Replace(name)
}
