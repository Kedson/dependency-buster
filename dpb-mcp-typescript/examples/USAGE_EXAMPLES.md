# Usage Examples

## Basic Dependency Analysis

Ask Claude/Cursor:
```
Analyze the dependencies in /path/to/my/php/project
```

Expected output:
- Production dependencies list
- Development dependencies list
- Dependency tree with relationships
- Statistics

## PSR-4 Validation

Ask:
```
Check if /path/to/project follows PSR-4 standards
```

Expected output:
- PSR-4 mappings
- List of violations with expected vs actual namespaces
- Compliance percentage

## Security Audit

Ask:
```
Run a security audit on this PHP project at /path/to/project
```

Expected output:
- List of vulnerabilities
- Risk level (low/medium/high/critical)
- Recommendations

## Generate Dependency Graph

Ask:
```
Generate a dependency graph for /path/to/project, focusing on Symfony packages, max depth 2
```

Expected output:
- Mermaid diagram syntax
- Can be rendered in documentation

## Namespace Analysis

Ask:
```
What namespaces are used in /path/to/project and where is App\Services\Authentication used?
```

Expected output:
- All namespaces with their files
- Specific namespace usage patterns

## License Analysis

Ask:
```
Analyze the licenses used in /path/to/project and check for compatibility issues
```

Expected output:
- License distribution
- Risk levels
- Compatibility warnings

## Multi-Repository Analysis

Ask:
```
Analyze all Faith FM repositories using config/faith-fm-repos.json:
- What dependencies are shared across services?
- Are there version conflicts?
- Generate a consolidated report
```

Expected output:
- Shared dependencies matrix
- Version conflict list
- Consolidated markdown report

## Comprehensive Documentation

Ask:
```
Generate comprehensive dependency documentation for /path/to/project and save it to DEPENDENCIES.md
```

Expected output:
- Full markdown documentation
- Saved to specified path

## Circular Dependency Detection

Ask:
```
Check /path/to/project for circular dependencies
```

Expected output:
- List of circular dependency chains
- Count of cycles

## Advanced: Combining Multiple Tools

Ask:
```
For the project at /path/to/my/project:
1. Analyze all dependencies
2. Check PSR-4 compliance
3. Run security audit
4. Generate dependency graph
5. Analyze licenses
6. Create comprehensive documentation in DOCS.md

Summarize any critical issues found.
```

This will use multiple tools and provide a comprehensive analysis.

## Faith FM Platform Example

```
I need to prepare for the Faith FM platform rebuild. Using config/faith-fm-repos.json:

1. Analyze dependencies across all services
2. Identify shared dependencies
3. Find version conflicts between services
4. Recommend a dependency consolidation strategy
5. Highlight any security concerns
6. Suggest a migration plan to align all services on compatible versions

Generate a report I can share with the team.
```

## AzuraCast Testing Example

```
I want to learn about the AzuraCast codebase at ~/test/azuracast:

1. What are the main dependencies?
2. How is the autoloading structured?
3. Are there any PSR-4 violations?
4. What namespaces are used?
5. Generate a visual dependency graph
6. Check for security issues
7. What licenses are used?

Create documentation in ANALYSIS.md
```

## Code Review Assistant

```
For this PR that adds a new service dependency:

1. Analyze how this new package fits into our dependency tree
2. Check if it conflicts with existing packages
3. Review the license compatibility
4. Check for known security issues
5. Verify PSR-4 compliance if it affects autoloading

Provide a summary for code review.
```

## Tips for Best Results

- Always use absolute paths
- Be specific about which packages you want to focus on
- Request visualizations (graphs) for complex dependency trees
- Save important analyses to markdown files for documentation
- Use multi-repo analysis for microservices architectures
- Combine multiple tools for comprehensive insights
