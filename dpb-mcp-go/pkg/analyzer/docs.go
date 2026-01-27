package analyzer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kedson/dpb-mcp/pkg/composer"
)

// GenerateComprehensiveDocs generates comprehensive markdown documentation
func GenerateComprehensiveDocs(repoPath, outputPath string) (string, error) {
	// Gather all data
	composerJSON, err := composer.ReadComposerJSON(repoPath)
	if err != nil {
		return "", err
	}

	depsJSON, _ := AnalyzeDependencies(repoPath)
	psr4JSON, _ := AnalyzePSR4Autoloading(repoPath)
	nsJSON, _ := DetectNamespaces(repoPath)
	secJSON, _ := AuditSecurity(repoPath)
	licJSON, _ := AnalyzeLicenses(repoPath)

	// Parse results (simplified - in production would properly unmarshal)
	var sb strings.Builder

	sb.WriteString("# PHP Dependency Documentation\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format(time.RFC3339)))

	sb.WriteString("## Project Information\n\n")
	sb.WriteString(fmt.Sprintf("- **Name:** %s\n", getStringOrDefault(composerJSON.Name, "Unknown")))
	sb.WriteString(fmt.Sprintf("- **Description:** %s\n", getStringOrDefault(composerJSON.Description, "N/A")))
	sb.WriteString(fmt.Sprintf("- **Type:** %s\n", getStringOrDefault(composerJSON.Type, "library")))
	
	licenses := composer.GetLicenses(composerJSON)
	licenseStr := "Not specified"
	if len(licenses) > 0 {
		licenseStr = strings.Join(licenses, ", ")
	}
	sb.WriteString(fmt.Sprintf("- **License:** %s\n\n", licenseStr))

	sb.WriteString("## Dependency Summary\n\n")
	prodCount := 0
	devCount := 0
	if composerJSON.Require != nil {
		prodCount = len(composer.FilterPHPDependencies(composerJSON.Require))
	}
	if composerJSON.RequireDev != nil {
		devCount = len(composerJSON.RequireDev)
	}
	sb.WriteString(fmt.Sprintf("- **Production Dependencies:** %d\n", prodCount))
	sb.WriteString(fmt.Sprintf("- **Development Dependencies:** %d\n\n", devCount))

	sb.WriteString("## Analysis Results\n\n")
	sb.WriteString("### PSR-4 Autoloading\n\n")
	sb.WriteString(psr4JSON)
	sb.WriteString("\n\n")

	sb.WriteString("### Namespace Detection\n\n")
	sb.WriteString(nsJSON)
	sb.WriteString("\n\n")

	sb.WriteString("### Security Audit\n\n")
	sb.WriteString(secJSON)
	sb.WriteString("\n\n")

	sb.WriteString("### License Analysis\n\n")
	sb.WriteString(licJSON)
	sb.WriteString("\n\n")

	sb.WriteString("### Dependency Tree\n\n")
	sb.WriteString(depsJSON)
	sb.WriteString("\n")

	documentation := sb.String()

	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(documentation), 0644); err != nil {
			return "", err
		}
		return fmt.Sprintf("Documentation saved to: %s", outputPath), nil
	}

	return documentation, nil
}

// getStringOrDefault returns string or default value
func getStringOrDefault(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}
