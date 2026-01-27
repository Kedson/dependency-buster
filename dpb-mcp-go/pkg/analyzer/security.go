package analyzer

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/faithfm/php-dependency-mcp/pkg/composer"
	"github.com/faithfm/php-dependency-mcp/pkg/types"
)

// SecurityAuditResult represents security audit output
type SecurityAuditResult struct {
	Vulnerabilities []types.SecurityVulnerability `json:"vulnerabilities"`
	RiskLevel       string                        `json:"riskLevel"`
	Summary         SecuritySummary               `json:"summary"`
}

// SecuritySummary represents security statistics
type SecuritySummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

// AuditSecurity audits dependencies for security issues
func AuditSecurity(repoPath string) (string, error) {
	lock, err := composer.ReadComposerLock(repoPath)
	if err != nil {
		return "", err
	}

	vulnerabilities := make([]types.SecurityVulnerability, 0)

	allPackages := append([]types.PackageInfo{}, lock.Packages...)
	if lock.PackagesDev != nil {
		allPackages = append(allPackages, lock.PackagesDev...)
	}

	for _, pkg := range allPackages {
		// Check for dev versions
		if strings.Contains(pkg.Version, "dev") && !strings.Contains(pkg.Version, "dev-") {
			vulnerabilities = append(vulnerabilities, types.SecurityVulnerability{
				Package:        pkg.Name,
				Version:        pkg.Version,
				Severity:       "medium",
				Description:    "Using development version in production",
				Recommendation: "Pin to a stable release version",
			})
		}

		// Check for pre-1.0 versions
		if strings.HasPrefix(pkg.Version, "0.") {
			vulnerabilities = append(vulnerabilities, types.SecurityVulnerability{
				Package:        pkg.Name,
				Version:        pkg.Version,
				Severity:       "low",
				Description:    "Using pre-1.0 version (potentially unstable)",
				Recommendation: "Consider upgrading to a stable 1.x+ version if available",
			})
		}

		// Check for very old packages
		if pkg.Time != "" {
			pkgTime, err := time.Parse(time.RFC3339, pkg.Time)
			if err == nil {
				fiveYearsAgo := time.Now().AddDate(-5, 0, 0)
				if pkgTime.Before(fiveYearsAgo) {
					vulnerabilities = append(vulnerabilities, types.SecurityVulnerability{
						Package:        pkg.Name,
						Version:        pkg.Version,
						Severity:       "medium",
						Description:    "Package has not been updated in over 5 years",
						Recommendation: "Check for maintained alternatives or security advisories",
					})
				}
			}
		}
	}

	summary := SecuritySummary{}
	riskLevel := "low"

	for _, vuln := range vulnerabilities {
		switch vuln.Severity {
		case "critical":
			summary.Critical++
		case "high":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		}
	}

	if summary.Critical > 0 {
		riskLevel = "critical"
	} else if summary.High > 0 {
		riskLevel = "high"
	} else if summary.Medium > 0 {
		riskLevel = "medium"
	}

	result := SecurityAuditResult{
		Vulnerabilities: vulnerabilities,
		RiskLevel:       riskLevel,
		Summary:         summary,
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// LicenseAnalysisResult represents license analysis output
type LicenseAnalysisResult struct {
	Distribution        []types.LicenseDistribution `json:"distribution"`
	CompatibilityIssues []string                    `json:"compatibilityIssues"`
	Summary             LicenseSummary              `json:"summary"`
}

// LicenseSummary represents license statistics
type LicenseSummary struct {
	TotalPackages    int `json:"totalPackages"`
	UniqueLicenses   int `json:"uniqueLicenses"`
	UnknownLicenses  int `json:"unknownLicenses"`
}

// AnalyzeLicenses analyzes license distribution and compatibility
func AnalyzeLicenses(repoPath string) (string, error) {
	lock, err := composer.ReadComposerLock(repoPath)
	if err != nil {
		return "", err
	}

	licenseMap := make(map[string][]string)
	unknownCount := 0

	allPackages := append([]types.PackageInfo{}, lock.Packages...)
	if lock.PackagesDev != nil {
		allPackages = append(allPackages, lock.PackagesDev...)
	}

	for _, pkg := range allPackages {
		licenses := pkg.License
		if len(licenses) == 0 {
			licenses = []string{"Unknown"}
		}

		for _, license := range licenses {
			if license == "Unknown" {
				unknownCount++
			}
			licenseMap[license] = append(licenseMap[license], pkg.Name)
		}
	}

	distribution := make([]types.LicenseDistribution, 0, len(licenseMap))
	for license, packages := range licenseMap {
		distribution = append(distribution, types.LicenseDistribution{
			License:   license,
			Count:     len(packages),
			Packages:  packages,
			RiskLevel: assessLicenseRisk(license),
		})
	}

	// Check for compatibility issues
	compatibilityIssues := make([]string, 0)
	hasGPL := false
	hasProprietary := false

	for _, dist := range distribution {
		if strings.Contains(dist.License, "GPL") {
			hasGPL = true
		}
		if strings.Contains(dist.License, "Proprietary") {
			hasProprietary = true
		}
	}

	if hasGPL && hasProprietary {
		compatibilityIssues = append(compatibilityIssues,
			"Potential conflict: GPL and Proprietary licenses detected. Review compatibility.")
	}

	result := LicenseAnalysisResult{
		Distribution:        distribution,
		CompatibilityIssues: compatibilityIssues,
		Summary: LicenseSummary{
			TotalPackages:   len(allPackages),
			UniqueLicenses:  len(licenseMap),
			UnknownLicenses: unknownCount,
		},
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// assessLicenseRisk determines risk level for a license
func assessLicenseRisk(license string) string {
	safeLicenses := []string{"MIT", "Apache-2.0", "BSD-3-Clause", "BSD-2-Clause", "ISC"}
	for _, safe := range safeLicenses {
		if license == safe {
			return "safe"
		}
	}

	cautionLicenses := []string{"LGPL", "MPL", "EPL"}
	for _, caution := range cautionLicenses {
		if strings.Contains(license, caution) {
			return "caution"
		}
	}

	if strings.Contains(license, "GPL") || license == "Unknown" || strings.Contains(license, "Proprietary") {
		return "review-required"
	}

	return "caution"
}
