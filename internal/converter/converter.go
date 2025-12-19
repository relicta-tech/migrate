// Package converter transforms release tool configs to Relicta format.
package converter

import (
	"fmt"
	"strings"

	"github.com/relicta-tech/migrate/internal/detector"
)

// RelictaConfig represents a Relicta release.config.yaml structure.
type RelictaConfig struct {
	Versioning VersioningConfig `yaml:"versioning"`
	Changelog  ChangelogConfig  `yaml:"changelog,omitempty"`
	Git        GitConfig        `yaml:"git,omitempty"`
	Plugins    []PluginConfig   `yaml:"plugins,omitempty"`
	AI         *AIConfig        `yaml:"ai,omitempty"`
}

// VersioningConfig holds versioning settings.
type VersioningConfig struct {
	Strategy  string `yaml:"strategy"`
	TagPrefix string `yaml:"tag_prefix,omitempty"`
}

// ChangelogConfig holds changelog settings.
type ChangelogConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Template string `yaml:"template,omitempty"`
	File     string `yaml:"file,omitempty"`
}

// GitConfig holds git settings.
type GitConfig struct {
	RequireCleanTree   bool   `yaml:"require_clean_tree"`
	PushTags           bool   `yaml:"push_tags"`
	CreateTag          bool   `yaml:"create_tag"`
	CommitMessage      string `yaml:"commit_message,omitempty"`
	TagMessage         string `yaml:"tag_message,omitempty"`
	RequireUpToDate    bool   `yaml:"require_up_to_date,omitempty"`
	AllowedBranches    []string `yaml:"allowed_branches,omitempty"`
}

// PluginConfig holds plugin settings.
type PluginConfig struct {
	Name    string         `yaml:"name"`
	Enabled bool           `yaml:"enabled"`
	Config  map[string]any `yaml:"config,omitempty"`
}

// AIConfig holds AI settings.
type AIConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider,omitempty"`
}

// Convert transforms a detected config to Relicta format.
func Convert(result *detector.Result) (*RelictaConfig, error) {
	switch result.Tool {
	case detector.ToolSemanticRelease:
		return convertSemanticRelease(result)
	case detector.ToolReleaseIt:
		return convertReleaseIt(result)
	case detector.ToolStandardVersion:
		return convertStandardVersion(result)
	case detector.ToolGoReleaser:
		return convertGoReleaser(result)
	default:
		return nil, fmt.Errorf("unsupported tool: %s", result.Tool)
	}
}

// convertSemanticRelease converts semantic-release config to Relicta.
func convertSemanticRelease(result *detector.Result) (*RelictaConfig, error) {
	data := result.ConfigData
	config := &RelictaConfig{
		Versioning: VersioningConfig{
			Strategy: "conventional",
		},
		Changelog: ChangelogConfig{
			Enabled: true,
			File:    "CHANGELOG.md",
		},
		Git: GitConfig{
			RequireCleanTree: true,
			PushTags:         true,
			CreateTag:        true,
		},
	}

	// Extract tag format
	if tagFormat, ok := data["tagFormat"].(string); ok {
		// semantic-release uses "${version}" syntax
		// Extract prefix (e.g., "v${version}" -> "v")
		prefix := strings.TrimSuffix(tagFormat, "${version}")
		if prefix != "" {
			config.Versioning.TagPrefix = prefix
		}
	}

	// Extract branches
	if branches, ok := data["branches"].([]any); ok {
		config.Git.AllowedBranches = extractBranches(branches)
	}

	// Convert plugins
	if plugins, ok := data["plugins"].([]any); ok {
		config.Plugins = convertSemanticReleasePlugins(plugins)
	}

	return config, nil
}

// convertReleaseIt converts release-it config to Relicta.
func convertReleaseIt(result *detector.Result) (*RelictaConfig, error) {
	data := result.ConfigData
	config := &RelictaConfig{
		Versioning: VersioningConfig{
			Strategy: "conventional",
		},
		Changelog: ChangelogConfig{
			Enabled: true,
			File:    "CHANGELOG.md",
		},
		Git: GitConfig{
			RequireCleanTree: true,
			PushTags:         true,
			CreateTag:        true,
		},
	}

	// Extract git config
	if git, ok := data["git"].(map[string]any); ok {
		if tagName, ok := git["tagName"].(string); ok {
			// release-it uses "${version}" syntax
			prefix := strings.TrimSuffix(tagName, "${version}")
			if prefix != "" {
				config.Versioning.TagPrefix = prefix
			}
		}
		if commitMessage, ok := git["commitMessage"].(string); ok {
			config.Git.CommitMessage = convertTemplate(commitMessage)
		}
		if tagAnnotation, ok := git["tagAnnotation"].(string); ok {
			config.Git.TagMessage = convertTemplate(tagAnnotation)
		}
		if requireCleanWorkingDir, ok := git["requireCleanWorkingDir"].(bool); ok {
			config.Git.RequireCleanTree = requireCleanWorkingDir
		}
		if push, ok := git["push"].(bool); ok {
			config.Git.PushTags = push
		}
	}

	// Extract npm config
	if npm, ok := data["npm"].(map[string]any); ok {
		if publish, ok := npm["publish"].(bool); ok && publish {
			config.Plugins = append(config.Plugins, PluginConfig{
				Name:    "npm",
				Enabled: true,
			})
		}
	}

	// Extract github config
	if github, ok := data["github"].(map[string]any); ok {
		if release, ok := github["release"].(bool); ok && release {
			ghConfig := PluginConfig{
				Name:    "github",
				Enabled: true,
				Config:  make(map[string]any),
			}
			if draft, ok := github["draft"].(bool); ok {
				ghConfig.Config["draft"] = draft
			}
			if preRelease, ok := github["preRelease"].(bool); ok {
				ghConfig.Config["prerelease"] = preRelease
			}
			config.Plugins = append(config.Plugins, ghConfig)
		}
	}

	// Extract gitlab config
	if gitlab, ok := data["gitlab"].(map[string]any); ok {
		if release, ok := gitlab["release"].(bool); ok && release {
			config.Plugins = append(config.Plugins, PluginConfig{
				Name:    "gitlab",
				Enabled: true,
			})
		}
	}

	return config, nil
}

// convertStandardVersion converts standard-version config to Relicta.
func convertStandardVersion(result *detector.Result) (*RelictaConfig, error) {
	data := result.ConfigData
	config := &RelictaConfig{
		Versioning: VersioningConfig{
			Strategy: "conventional",
		},
		Changelog: ChangelogConfig{
			Enabled: true,
			File:    "CHANGELOG.md",
		},
		Git: GitConfig{
			RequireCleanTree: true,
			PushTags:         true,
			CreateTag:        true,
		},
	}

	// Extract tag prefix
	if tagPrefix, ok := data["tagPrefix"].(string); ok {
		config.Versioning.TagPrefix = tagPrefix
	}

	// Extract skip options
	if skip, ok := data["skip"].(map[string]any); ok {
		if skipChangelog, ok := skip["changelog"].(bool); ok && skipChangelog {
			config.Changelog.Enabled = false
		}
		if skipTag, ok := skip["tag"].(bool); ok && skipTag {
			config.Git.CreateTag = false
		}
	}

	// Extract commit message
	if releaseCommitMessageFormat, ok := data["releaseCommitMessageFormat"].(string); ok {
		config.Git.CommitMessage = convertTemplate(releaseCommitMessageFormat)
	}

	// Extract changelog file path
	if infile, ok := data["infile"].(string); ok {
		config.Changelog.File = infile
	}

	return config, nil
}

// extractBranches extracts branch names from semantic-release branches config.
func extractBranches(branches []any) []string {
	var result []string
	for _, b := range branches {
		switch branch := b.(type) {
		case string:
			result = append(result, branch)
		case map[string]any:
			if name, ok := branch["name"].(string); ok {
				result = append(result, name)
			}
		}
	}
	return result
}

// convertSemanticReleasePlugins converts semantic-release plugins to Relicta plugins.
func convertSemanticReleasePlugins(plugins []any) []PluginConfig {
	var result []PluginConfig

	for _, p := range plugins {
		var pluginName string
		var pluginConfig map[string]any

		switch plugin := p.(type) {
		case string:
			pluginName = plugin
		case []any:
			if len(plugin) > 0 {
				if name, ok := plugin[0].(string); ok {
					pluginName = name
				}
				if len(plugin) > 1 {
					if cfg, ok := plugin[1].(map[string]any); ok {
						pluginConfig = cfg
					}
				}
			}
		}

		// Map semantic-release plugins to Relicta plugins
		relictaPlugin := mapSemanticReleasePlugin(pluginName, pluginConfig)
		if relictaPlugin != nil {
			result = append(result, *relictaPlugin)
		}
	}

	return result
}

// mapSemanticReleasePlugin maps a semantic-release plugin to Relicta equivalent.
func mapSemanticReleasePlugin(name string, config map[string]any) *PluginConfig {
	// Normalize plugin name
	name = strings.TrimPrefix(name, "@semantic-release/")

	switch name {
	case "github":
		return &PluginConfig{
			Name:    "github",
			Enabled: true,
			Config:  config,
		}
	case "gitlab":
		return &PluginConfig{
			Name:    "gitlab",
			Enabled: true,
			Config:  config,
		}
	case "npm":
		return &PluginConfig{
			Name:    "npm",
			Enabled: true,
			Config:  config,
		}
	case "changelog", "release-notes-generator":
		// Handled by Relicta core, not a plugin
		return nil
	case "commit-analyzer":
		// Handled by Relicta core
		return nil
	case "git":
		// Handled by Relicta core
		return nil
	case "exec":
		// Custom commands - note this in config
		return &PluginConfig{
			Name:    "custom",
			Enabled: false,
			Config: map[string]any{
				"_note":    "Migrate custom exec commands manually",
				"_original": config,
			},
		}
	default:
		// Unknown plugin - preserve for manual migration
		return &PluginConfig{
			Name:    name,
			Enabled: false,
			Config: map[string]any{
				"_note":    "Unknown plugin - requires manual migration",
				"_original": config,
			},
		}
	}
}

// convertTemplate converts template syntax from other tools to Relicta format.
func convertTemplate(template string) string {
	// ${version} -> {{.Version}}
	template = strings.ReplaceAll(template, "${version}", "{{.Version}}")
	// ${nextRelease.version} -> {{.Version}}
	template = strings.ReplaceAll(template, "${nextRelease.version}", "{{.Version}}")
	// {{version}} -> {{.Version}}
	template = strings.ReplaceAll(template, "{{version}}", "{{.Version}}")

	return template
}

// convertGoReleaser converts GoReleaser config to Relicta.
func convertGoReleaser(result *detector.Result) (*RelictaConfig, error) {
	data := result.ConfigData
	config := &RelictaConfig{
		Versioning: VersioningConfig{
			Strategy:  "conventional",
			TagPrefix: "v",
		},
		Changelog: ChangelogConfig{
			Enabled: true,
			File:    "CHANGELOG.md",
		},
		Git: GitConfig{
			RequireCleanTree: true,
			PushTags:         true,
			CreateTag:        true,
			AllowedBranches:  []string{"main"},
		},
	}

	// Extract project name for reference
	projectName := ""
	if pn, ok := data["project_name"].(string); ok {
		projectName = pn
	}

	// Extract changelog config
	if changelog, ok := data["changelog"].(map[string]any); ok {
		if skip, ok := changelog["skip"].(bool); ok && skip {
			config.Changelog.Enabled = false
		}
	}

	// Extract release config
	if release, ok := data["release"].(map[string]any); ok {
		ghConfig := PluginConfig{
			Name:    "github",
			Enabled: true,
			Config:  make(map[string]any),
		}

		// Extract GitHub owner/repo if present
		if github, ok := release["github"].(map[string]any); ok {
			if owner, ok := github["owner"].(string); ok {
				ghConfig.Config["owner"] = owner
			}
			if name, ok := github["name"].(string); ok {
				ghConfig.Config["repo"] = name
			}
		}

		// Extract draft setting
		if draft, ok := release["draft"].(bool); ok {
			ghConfig.Config["draft"] = draft
		}

		// Extract prerelease setting
		if prerelease, ok := release["prerelease"].(string); ok {
			ghConfig.Config["prerelease"] = prerelease == "auto"
		} else if prerelease, ok := release["prerelease"].(bool); ok {
			ghConfig.Config["prerelease"] = prerelease
		}

		// Extract name template
		if nameTemplate, ok := release["name_template"].(string); ok {
			ghConfig.Config["name_template"] = nameTemplate
		}

		config.Plugins = append(config.Plugins, ghConfig)
	} else {
		// Default GitHub plugin if no release config
		config.Plugins = append(config.Plugins, PluginConfig{
			Name:    "github",
			Enabled: true,
		})
	}

	// Extract build targets for assets config
	assets := extractGoReleaserAssets(data, projectName)
	if len(assets) > 0 {
		for i := range config.Plugins {
			if config.Plugins[i].Name == "github" {
				if config.Plugins[i].Config == nil {
					config.Plugins[i].Config = make(map[string]any)
				}
				config.Plugins[i].Config["assets"] = assets
				break
			}
		}
	}

	// Extract snapshot config for version template reference
	if snapshot, ok := data["snapshot"].(map[string]any); ok {
		if versionTemplate, ok := snapshot["version_template"].(string); ok {
			// Store for reference - Relicta handles this differently
			_ = versionTemplate
		}
	}

	return config, nil
}

// extractGoReleaserAssets generates asset patterns from GoReleaser build config.
func extractGoReleaserAssets(data map[string]any, projectName string) []string {
	var assets []string

	// Determine binary name
	binaryName := projectName
	if builds, ok := data["builds"].([]any); ok && len(builds) > 0 {
		if build, ok := builds[0].(map[string]any); ok {
			if binary, ok := build["binary"].(string); ok {
				binaryName = binary
			}
		}
	}

	if binaryName == "" {
		binaryName = "{{.ProjectName}}"
	}

	// Generate standard asset patterns based on common GoReleaser output
	goos := []string{"linux", "darwin", "windows"}
	goarch := []string{"amd64", "arm64"}

	// Try to extract actual targets from config
	if builds, ok := data["builds"].([]any); ok && len(builds) > 0 {
		if build, ok := builds[0].(map[string]any); ok {
			if g, ok := build["goos"].([]any); ok {
				goos = toStringSlice(g)
			}
			if g, ok := build["goarch"].([]any); ok {
				goarch = toStringSlice(g)
			}
		}
	}

	// Map to Relicta asset naming convention
	for _, os := range goos {
		for _, arch := range goarch {
			// Convert to Relicta naming: plugin-name_os_arch
			archName := arch
			if arch == "amd64" {
				archName = "x86_64"
			} else if arch == "arm64" {
				archName = "aarch64"
			}

			ext := ".tar.gz"
			if os == "windows" {
				ext = ".zip"
			}

			assets = append(assets, fmt.Sprintf("release/%s_%s_%s%s", binaryName, os, archName, ext))
		}
	}

	// Add checksums
	assets = append(assets, "release/checksums.txt")

	return assets
}

// toStringSlice converts []any to []string.
func toStringSlice(input []any) []string {
	result := make([]string, 0, len(input))
	for _, v := range input {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
