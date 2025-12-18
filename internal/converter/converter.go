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
