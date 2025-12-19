// Package detector identifies which release tool is configured in a project.
package detector

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Tool represents a release management tool.
type Tool string

// Tool constants for supported release management tools.
const (
	ToolNone            Tool = "none"
	ToolSemanticRelease Tool = "semantic-release"
	ToolReleaseIt       Tool = "release-it"
	ToolStandardVersion Tool = "standard-version"
	ToolGoReleaser      Tool = "goreleaser"
)

// Result contains detection results.
type Result struct {
	Tool       Tool
	ConfigFile string
	ConfigData map[string]any
	Details    map[string]any
}

// Detect identifies the release tool configuration in the given directory.
func Detect(dir string) (*Result, error) {
	// Try each tool in order of specificity
	detectors := []func(string) (*Result, error){
		detectSemanticRelease,
		detectReleaseIt,
		detectStandardVersion,
		detectGoReleaser,
	}

	for _, detect := range detectors {
		result, err := detect(dir)
		if err != nil {
			continue // Try next detector
		}
		if result != nil && result.Tool != ToolNone {
			return result, nil
		}
	}

	return &Result{Tool: ToolNone}, nil
}

// detectSemanticRelease looks for semantic-release configuration.
func detectSemanticRelease(dir string) (*Result, error) {
	// Check dedicated config files first
	configFiles := []string{
		".releaserc",
		".releaserc.json",
		".releaserc.yaml",
		".releaserc.yml",
		"release.config.js",
		"release.config.cjs",
	}

	for _, file := range configFiles {
		path := filepath.Join(dir, file)
		if data, err := readConfigFile(path); err == nil {
			return &Result{
				Tool:       ToolSemanticRelease,
				ConfigFile: path,
				ConfigData: data,
				Details:    extractSemanticReleaseDetails(data),
			}, nil
		}
	}

	// Check package.json for "release" key
	pkgPath := filepath.Join(dir, "package.json")
	if pkg, err := readPackageJSON(pkgPath); err == nil {
		if release, ok := pkg["release"].(map[string]any); ok {
			return &Result{
				Tool:       ToolSemanticRelease,
				ConfigFile: pkgPath + " (release key)",
				ConfigData: release,
				Details:    extractSemanticReleaseDetails(release),
			}, nil
		}
	}

	return nil, nil
}

// detectReleaseIt looks for release-it configuration.
func detectReleaseIt(dir string) (*Result, error) {
	configFiles := []string{
		".release-it.json",
		".release-it.yaml",
		".release-it.yml",
		".release-it.js",
		".release-it.cjs",
		".release-it.ts",
	}

	for _, file := range configFiles {
		path := filepath.Join(dir, file)
		if data, err := readConfigFile(path); err == nil {
			return &Result{
				Tool:       ToolReleaseIt,
				ConfigFile: path,
				ConfigData: data,
				Details:    extractReleaseItDetails(data),
			}, nil
		}
	}

	// Check package.json for "release-it" key
	pkgPath := filepath.Join(dir, "package.json")
	if pkg, err := readPackageJSON(pkgPath); err == nil {
		if releaseIt, ok := pkg["release-it"].(map[string]any); ok {
			return &Result{
				Tool:       ToolReleaseIt,
				ConfigFile: pkgPath + " (release-it key)",
				ConfigData: releaseIt,
				Details:    extractReleaseItDetails(releaseIt),
			}, nil
		}
	}

	return nil, nil
}

// detectStandardVersion looks for standard-version configuration.
func detectStandardVersion(dir string) (*Result, error) {
	configFiles := []string{
		".versionrc",
		".versionrc.json",
		".versionrc.js",
		".versionrc.cjs",
	}

	for _, file := range configFiles {
		path := filepath.Join(dir, file)
		if data, err := readConfigFile(path); err == nil {
			return &Result{
				Tool:       ToolStandardVersion,
				ConfigFile: path,
				ConfigData: data,
				Details:    extractStandardVersionDetails(data),
			}, nil
		}
	}

	// Check package.json for "standard-version" key
	pkgPath := filepath.Join(dir, "package.json")
	if pkg, err := readPackageJSON(pkgPath); err == nil {
		if sv, ok := pkg["standard-version"].(map[string]any); ok {
			return &Result{
				Tool:       ToolStandardVersion,
				ConfigFile: pkgPath + " (standard-version key)",
				ConfigData: sv,
				Details:    extractStandardVersionDetails(sv),
			}, nil
		}
	}

	return nil, nil
}

// readConfigFile reads JSON or YAML config files.
func readConfigFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]any

	// Try JSON first
	if err := json.Unmarshal(data, &result); err == nil {
		return result, nil
	}

	// Try YAML
	if err := yaml.Unmarshal(data, &result); err == nil {
		return result, nil
	}

	// For JS/TS files, we can't parse them directly
	// Return empty map to indicate file exists
	ext := filepath.Ext(path)
	if ext == ".js" || ext == ".cjs" || ext == ".ts" {
		return map[string]any{"_jsConfig": true}, nil
	}

	return nil, os.ErrNotExist
}

// readPackageJSON reads and parses package.json.
func readPackageJSON(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// extractSemanticReleaseDetails extracts key details from semantic-release config.
func extractSemanticReleaseDetails(data map[string]any) map[string]any {
	details := make(map[string]any)

	if branches, ok := data["branches"]; ok {
		details["branches"] = branches
	}
	if plugins, ok := data["plugins"]; ok {
		details["plugins"] = countPlugins(plugins)
	}
	if tagFormat, ok := data["tagFormat"]; ok {
		details["tagFormat"] = tagFormat
	}

	return details
}

// extractReleaseItDetails extracts key details from release-it config.
func extractReleaseItDetails(data map[string]any) map[string]any {
	details := make(map[string]any)

	if git, ok := data["git"].(map[string]any); ok {
		if tagName, ok := git["tagName"]; ok {
			details["tagName"] = tagName
		}
	}
	if npm, ok := data["npm"].(map[string]any); ok {
		if publish, ok := npm["publish"]; ok {
			details["npmPublish"] = publish
		}
	}
	if github, ok := data["github"].(map[string]any); ok {
		if release, ok := github["release"]; ok {
			details["githubRelease"] = release
		}
	}

	return details
}

// extractStandardVersionDetails extracts key details from standard-version config.
func extractStandardVersionDetails(data map[string]any) map[string]any {
	details := make(map[string]any)

	if tagPrefix, ok := data["tagPrefix"]; ok {
		details["tagPrefix"] = tagPrefix
	}
	if skip, ok := data["skip"].(map[string]any); ok {
		details["skip"] = skip
	}
	if types, ok := data["types"]; ok {
		details["customTypes"] = true
		_ = types // Use the variable
	}

	return details
}

// countPlugins counts the number of plugins in a semantic-release config.
func countPlugins(plugins any) int {
	switch p := plugins.(type) {
	case []any:
		return len(p)
	case []string:
		return len(p)
	default:
		return 0
	}
}

// detectGoReleaser looks for GoReleaser configuration.
func detectGoReleaser(dir string) (*Result, error) {
	configFiles := []string{
		".goreleaser.yml",
		".goreleaser.yaml",
		"goreleaser.yml",
		"goreleaser.yaml",
	}

	for _, file := range configFiles {
		path := filepath.Join(dir, file)
		if data, err := readConfigFile(path); err == nil {
			return &Result{
				Tool:       ToolGoReleaser,
				ConfigFile: path,
				ConfigData: data,
				Details:    extractGoReleaserDetails(data),
			}, nil
		}
	}

	return nil, nil
}

// extractGoReleaserDetails extracts key details from GoReleaser config.
func extractGoReleaserDetails(data map[string]any) map[string]any {
	details := make(map[string]any)

	// Extract project name
	if projectName, ok := data["project_name"].(string); ok {
		details["projectName"] = projectName
	}

	// Count builds
	if builds, ok := data["builds"].([]any); ok {
		details["buildsCount"] = len(builds)
	}

	// Check for archives
	if archives, ok := data["archives"].([]any); ok {
		details["archivesCount"] = len(archives)
	}

	// Check release config
	if release, ok := data["release"].(map[string]any); ok {
		if github, ok := release["github"].(map[string]any); ok {
			details["github"] = github
		}
		if draft, ok := release["draft"].(bool); ok {
			details["draft"] = draft
		}
		if prerelease, ok := release["prerelease"]; ok {
			details["prerelease"] = prerelease
		}
	}

	// Check for changelog config
	if changelog, ok := data["changelog"].(map[string]any); ok {
		if skip, ok := changelog["skip"].(bool); ok {
			details["changelogSkip"] = skip
		}
		if sort, ok := changelog["sort"].(string); ok {
			details["changelogSort"] = sort
		}
	}

	// Extract GOOS/GOARCH targets
	if builds, ok := data["builds"].([]any); ok && len(builds) > 0 {
		if build, ok := builds[0].(map[string]any); ok {
			if goos, ok := build["goos"].([]any); ok {
				details["goos"] = goos
			}
			if goarch, ok := build["goarch"].([]any); ok {
				details["goarch"] = goarch
			}
		}
	}

	return details
}
