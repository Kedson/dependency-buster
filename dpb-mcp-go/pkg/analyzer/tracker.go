// Package analyzer - Dependency Tracker for Go implementation
// Tracks dependency changes with timestamps for reverting/replacing non-compliant dependencies

package analyzer

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const TrackerFile = ".dpb-dependency-tracker.json"

// DependencySnapshot represents a point-in-time view of all dependencies
type DependencySnapshot struct {
	Timestamp    string                   `json:"timestamp"`
	Checksum     string                   `json:"checksum"`
	Dependencies []TrackedDependency      `json:"dependencies"`
	Metadata     SnapshotMetadata         `json:"metadata"`
}

// TrackedDependency represents a single dependency with tracking info
type TrackedDependency struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	Type           string `json:"type"` // "production" or "development"
	AddedAt        string `json:"addedAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
	License        string `json:"license,omitempty"`
	SecurityStatus string `json:"securityStatus,omitempty"`
}

// SnapshotMetadata contains additional info about the snapshot
type SnapshotMetadata struct {
	RepoPath       string `json:"repoPath"`
	PackageManager string `json:"packageManager"`
	TotalCount     int    `json:"totalCount"`
}

// DependencyChange represents a change between snapshots
type DependencyChange struct {
	Type       string `json:"type"` // "added", "removed", "updated"
	Name       string `json:"name"`
	OldVersion string `json:"oldVersion,omitempty"`
	NewVersion string `json:"newVersion,omitempty"`
	Timestamp  string `json:"timestamp"`
	Reason     string `json:"reason,omitempty"`
}

// ComplianceIssue represents a dependency compliance problem
type ComplianceIssue struct {
	Dependency       string `json:"dependency"`
	Version          string `json:"version"`
	Issue            string `json:"issue"` // "license", "security", "outdated", "deprecated"
	Severity         string `json:"severity"` // "critical", "high", "medium", "low"
	Description      string `json:"description"`
	Recommendation   string `json:"recommendation"`
	AutoFixAvailable bool   `json:"autoFixAvailable"`
}

// DependencyHistory contains current snapshot and categorized dependencies
type DependencyHistory struct {
	CurrentSnapshot DependencySnapshot  `json:"currentSnapshot"`
	RecentlyAdded   []TrackedDependency `json:"recentlyAdded"`
	RecentlyUpdated []TrackedDependency `json:"recentlyUpdated"`
	Stale           []TrackedDependency `json:"stale"`
}

// CreateDependencySnapshot creates a new snapshot of all dependencies
func CreateDependencySnapshot(repoPath string) (*DependencySnapshot, error) {
	deps, err := AnalyzeDependenciesRaw(repoPath)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	
	// Load existing tracker to preserve timestamps
	existing, _ := LoadTracker(repoPath)
	existingDeps := make(map[string]TrackedDependency)
	if existing != nil {
		for _, d := range existing.Dependencies {
			existingDeps[d.Name] = d
		}
	}

	var tracked []TrackedDependency

	// Process production dependencies
	for _, pkg := range deps.Tree {
		if pkg.Type == "production" {
			existing, found := existingDeps[pkg.Name]
			td := TrackedDependency{
				Name:           pkg.Name,
				Version:        pkg.Version,
				Type:           "production",
				License:        pkg.License,
				SecurityStatus: "unknown",
			}
			if found {
				td.AddedAt = existing.AddedAt
				if existing.Version != pkg.Version {
					td.UpdatedAt = now
				} else {
					td.UpdatedAt = existing.UpdatedAt
				}
			} else {
				td.AddedAt = now
				td.UpdatedAt = now
			}
			tracked = append(tracked, td)
		}
	}

	// Process dev dependencies
	for _, pkg := range deps.Tree {
		if pkg.Type == "development" {
			existing, found := existingDeps[pkg.Name]
			td := TrackedDependency{
				Name:           pkg.Name,
				Version:        pkg.Version,
				Type:           "development",
				License:        pkg.License,
				SecurityStatus: "unknown",
			}
			if found {
				td.AddedAt = existing.AddedAt
				if existing.Version != pkg.Version {
					td.UpdatedAt = now
				} else {
					td.UpdatedAt = existing.UpdatedAt
				}
			} else {
				td.AddedAt = now
				td.UpdatedAt = now
			}
			tracked = append(tracked, td)
		}
	}

	// Calculate checksum
	var names []string
	for _, d := range tracked {
		names = append(names, fmt.Sprintf("%s@%s", d.Name, d.Version))
	}
	sort.Strings(names)
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v", names)))
	checksum := fmt.Sprintf("%x", hash)[:16]

	snapshot := &DependencySnapshot{
		Timestamp:    now,
		Checksum:     checksum,
		Dependencies: tracked,
		Metadata: SnapshotMetadata{
			RepoPath:       repoPath,
			PackageManager: "composer",
			TotalCount:     len(tracked),
		},
	}

	return snapshot, nil
}

// LoadTracker loads existing tracker data from file
func LoadTracker(repoPath string) (*DependencySnapshot, error) {
	trackerPath := filepath.Join(repoPath, TrackerFile)
	data, err := os.ReadFile(trackerPath)
	if err != nil {
		return nil, err
	}

	var snapshot DependencySnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// SaveSnapshot saves a snapshot to the tracker file
func SaveSnapshot(repoPath string, snapshot *DependencySnapshot) error {
	trackerPath := filepath.Join(repoPath, TrackerFile)
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(trackerPath, data, 0644)
}

// CompareSnapshots returns the differences between two snapshots
func CompareSnapshots(oldSnapshot, newSnapshot *DependencySnapshot) []DependencyChange {
	var changes []DependencyChange

	oldDeps := make(map[string]TrackedDependency)
	for _, d := range oldSnapshot.Dependencies {
		oldDeps[d.Name] = d
	}

	newDeps := make(map[string]TrackedDependency)
	for _, d := range newSnapshot.Dependencies {
		newDeps[d.Name] = d
	}

	// Find added and updated
	for name, newDep := range newDeps {
		if oldDep, found := oldDeps[name]; !found {
			changes = append(changes, DependencyChange{
				Type:       "added",
				Name:       name,
				NewVersion: newDep.Version,
				Timestamp:  newSnapshot.Timestamp,
			})
		} else if oldDep.Version != newDep.Version {
			changes = append(changes, DependencyChange{
				Type:       "updated",
				Name:       name,
				OldVersion: oldDep.Version,
				NewVersion: newDep.Version,
				Timestamp:  newSnapshot.Timestamp,
			})
		}
	}

	// Find removed
	for name, oldDep := range oldDeps {
		if _, found := newDeps[name]; !found {
			changes = append(changes, DependencyChange{
				Type:       "removed",
				Name:       name,
				OldVersion: oldDep.Version,
				Timestamp:  newSnapshot.Timestamp,
			})
		}
	}

	return changes
}

// GetDependencyHistory returns categorized dependency information
func GetDependencyHistory(repoPath string) (*DependencyHistory, error) {
	snapshot, err := CreateDependencySnapshot(repoPath)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	oneYearAgo := now.AddDate(-1, 0, 0)

	var recentlyAdded, recentlyUpdated, stale []TrackedDependency

	for _, dep := range snapshot.Dependencies {
		if dep.AddedAt != "" {
			addedTime, _ := time.Parse(time.RFC3339, dep.AddedAt)
			if addedTime.After(thirtyDaysAgo) {
				recentlyAdded = append(recentlyAdded, dep)
			}
		}

		if dep.UpdatedAt != "" && dep.UpdatedAt != dep.AddedAt {
			updatedTime, _ := time.Parse(time.RFC3339, dep.UpdatedAt)
			if updatedTime.After(thirtyDaysAgo) {
				recentlyUpdated = append(recentlyUpdated, dep)
			}
			if updatedTime.Before(oneYearAgo) {
				stale = append(stale, dep)
			}
		}
	}

	return &DependencyHistory{
		CurrentSnapshot: *snapshot,
		RecentlyAdded:   recentlyAdded,
		RecentlyUpdated: recentlyUpdated,
		Stale:           stale,
	}, nil
}

// CheckCompliance checks dependencies for compliance issues
func CheckCompliance(repoPath string) ([]ComplianceIssue, error) {
	snapshot, err := CreateDependencySnapshot(repoPath)
	if err != nil {
		return nil, err
	}

	var issues []ComplianceIssue

	restrictiveLicenses := []string{"GPL-3.0", "AGPL-3.0", "GPL-2.0", "SSPL"}

	for _, dep := range snapshot.Dependencies {
		// Check for restrictive licenses in production
		if dep.Type == "production" && dep.License != "" {
			for _, restricted := range restrictiveLicenses {
				if dep.License == restricted {
					issues = append(issues, ComplianceIssue{
						Dependency:       dep.Name,
						Version:          dep.Version,
						Issue:            "license",
						Severity:         "high",
						Description:      fmt.Sprintf("Uses restrictive license: %s", dep.License),
						Recommendation:   "Consider replacing with an MIT/Apache-2.0 licensed alternative",
						AutoFixAvailable: false,
					})
				}
			}
		}

		// Check for stale dependencies
		if dep.UpdatedAt != "" {
			updatedTime, _ := time.Parse(time.RFC3339, dep.UpdatedAt)
			twoYearsAgo := time.Now().AddDate(-2, 0, 0)
			if updatedTime.Before(twoYearsAgo) {
				issues = append(issues, ComplianceIssue{
					Dependency:       dep.Name,
					Version:          dep.Version,
					Issue:            "outdated",
					Severity:         "low",
					Description:      "Not updated in over 2 years",
					Recommendation:   "Check if a newer version is available",
					AutoFixAvailable: true,
				})
			}
		}
	}

	return issues, nil
}

// GenerateRevertCommand generates a command to revert a dependency change
func GenerateRevertCommand(change DependencyChange) string {
	switch change.Type {
	case "added":
		return fmt.Sprintf("composer remove %s", change.Name)
	case "removed":
		return fmt.Sprintf("composer require %s:%s", change.Name, change.OldVersion)
	case "updated":
		return fmt.Sprintf("composer require %s:%s", change.Name, change.OldVersion)
	default:
		return ""
	}
}
