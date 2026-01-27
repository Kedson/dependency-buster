package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kedson/dpb-mcp/pkg/composer"
	"github.com/kedson/dpb-mcp/pkg/types"
	"golang.org/x/sync/errgroup"
)

// AnalyzeMultipleRepositories analyzes dependencies across multiple repos
func AnalyzeMultipleRepositories(configPath string) (string, error) {
	// Read config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	var repos []types.RepoConfig
	if err := json.Unmarshal(data, &repos); err != nil {
		return "", err
	}

	// Load repository data concurrently
	type repoData struct {
		name         string
		composer     *types.ComposerJSON
		dependencies map[string]string
	}

	repoDataMap := make(map[string]*repoData)
	var mu sync.Mutex
	var g errgroup.Group

	for _, repo := range repos {
		repoCopy := repo
		g.Go(func() error {
			composerJSON, err := composer.ReadComposerJSON(repoCopy.Path)
			if err != nil {
				return nil // Skip repos with errors
			}

			deps := make(map[string]string)
			if composerJSON.Require != nil {
				for k, v := range composerJSON.Require {
					if k != "php" {
						deps[k] = v
					}
				}
			}
			if composerJSON.RequireDev != nil {
				for k, v := range composerJSON.RequireDev {
					deps[k] = v
				}
			}

			mu.Lock()
			repoDataMap[repoCopy.Name] = &repoData{
				name:         repoCopy.Name,
				composer:     composerJSON,
				dependencies: deps,
			}
			mu.Unlock()

			return nil
		})
	}

	g.Wait()

	// Find shared dependencies
	packageUsage := make(map[string][]string)
	for repoName, data := range repoDataMap {
		for pkg := range data.dependencies {
			packageUsage[pkg] = append(packageUsage[pkg], repoName)
		}
	}

	sharedDependencies := make(map[string][]string)
	for pkg, repos := range packageUsage {
		if len(repos) > 1 {
			sharedDependencies[pkg] = repos
		}
	}

	// Find version conflicts
	versionConflicts := make([]types.VersionConflict, 0)
	for pkg, usedByRepos := range packageUsage {
		versions := make(map[string][]string)
		for _, repoName := range usedByRepos {
			version := repoDataMap[repoName].dependencies[pkg]
			versions[version] = append(versions[version], repoName)
		}

		if len(versions) > 1 {
			conflict := types.VersionConflict{
				Package:  pkg,
				Versions: make([]types.RepoVersion, 0),
			}

			for version, repos := range versions {
				for _, repo := range repos {
					conflict.Versions = append(conflict.Versions, types.RepoVersion{
						Repo:    repo,
						Version: version,
					})
				}
			}

			versionConflicts = append(versionConflicts, conflict)
		}
	}

	// Aggregate licenses
	licenseCount := make(map[string]int)
	for _, data := range repoDataMap {
		licenses := composer.GetLicenses(data.composer)
		for _, license := range licenses {
			licenseCount[license]++
		}
	}

	// Count total unique packages
	allPackages := make(map[string]bool)
	for _, data := range repoDataMap {
		for pkg := range data.dependencies {
			allPackages[pkg] = true
		}
	}

	// Generate report
	report := generateConsolidatedReport(repos, sharedDependencies, versionConflicts, 
		len(allPackages), licenseCount)

	return report, nil
}

// generateConsolidatedReport generates a markdown report
func generateConsolidatedReport(repos []types.RepoConfig, sharedDeps map[string][]string,
	conflicts []types.VersionConflict, totalPkgs int, licenses map[string]int) string {

	var sb strings.Builder

	sb.WriteString("# Multi-Repository Dependency Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("## Repositories Analyzed\n\n")

	for _, repo := range repos {
		sb.WriteString(fmt.Sprintf("- **%s** (%s)", repo.Name, repo.Type))
		if repo.Team != "" {
			sb.WriteString(fmt.Sprintf(" - Team: %s", repo.Team))
		}
		if repo.Description != "" {
			sb.WriteString(fmt.Sprintf("\n  %s", repo.Description))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- Total unique packages: %d\n", totalPkgs))
	sb.WriteString(fmt.Sprintf("- Shared dependencies: %d\n", len(sharedDeps)))
	sb.WriteString(fmt.Sprintf("- Version conflicts: %d\n\n", len(conflicts)))

	if len(sharedDeps) > 0 {
		sb.WriteString("## Shared Dependencies\n\n")
		sb.WriteString("| Package | Used By |\n")
		sb.WriteString("|---------|----------|\n")
		for pkg, repos := range sharedDeps {
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", pkg, strings.Join(repos, ", ")))
		}
		sb.WriteString("\n")
	}

	if len(conflicts) > 0 {
		sb.WriteString("## ⚠️ Version Conflicts\n\n")
		for _, conflict := range conflicts {
			sb.WriteString(fmt.Sprintf("### %s\n\n", conflict.Package))
			for _, version := range conflict.Versions {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", version.Repo, version.Version))
			}
			sb.WriteString("\n")
		}
	}

	if len(licenses) > 0 {
		sb.WriteString("## License Distribution\n\n")
		sb.WriteString("| License | Count |\n")
		sb.WriteString("|---------|-------|\n")
		for license, count := range licenses {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", license, count))
		}
	}

	return sb.String()
}
