package converter

import (
	"testing"

	"github.com/relicta-tech/migrate/internal/detector"
)

func TestConvert_SemanticRelease(t *testing.T) {
	tests := []struct {
		name       string
		configData map[string]any
		wantPrefix string
		wantPlugins int
	}{
		{
			name: "basic config",
			configData: map[string]any{
				"branches": []any{"main"},
			},
			wantPrefix:  "",
			wantPlugins: 0,
		},
		{
			name: "with tag format",
			configData: map[string]any{
				"tagFormat": "v${version}",
				"branches":  []any{"main"},
			},
			wantPrefix:  "v",
			wantPlugins: 0,
		},
		{
			name: "with github plugin",
			configData: map[string]any{
				"plugins": []any{
					"@semantic-release/github",
				},
			},
			wantPrefix:  "",
			wantPlugins: 1,
		},
		{
			name: "with multiple plugins",
			configData: map[string]any{
				"plugins": []any{
					"@semantic-release/commit-analyzer",
					"@semantic-release/release-notes-generator",
					"@semantic-release/changelog",
					"@semantic-release/github",
					"@semantic-release/npm",
				},
			},
			wantPrefix:  "",
			wantPlugins: 2, // Only github and npm are converted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &detector.Result{
				Tool:       detector.ToolSemanticRelease,
				ConfigFile: ".releaserc.json",
				ConfigData: tt.configData,
			}

			config, err := Convert(result)
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			if config.Versioning.TagPrefix != tt.wantPrefix {
				t.Errorf("TagPrefix = %v, want %v", config.Versioning.TagPrefix, tt.wantPrefix)
			}

			if len(config.Plugins) != tt.wantPlugins {
				t.Errorf("Plugins count = %v, want %v", len(config.Plugins), tt.wantPlugins)
			}
		})
	}
}

func TestConvert_ReleaseIt(t *testing.T) {
	tests := []struct {
		name             string
		configData       map[string]any
		wantPrefix       string
		wantGitHub       bool
		wantNPM          bool
		wantCommitMsg    string
	}{
		{
			name: "basic git config",
			configData: map[string]any{
				"git": map[string]any{
					"tagName": "v${version}",
				},
			},
			wantPrefix: "v",
		},
		{
			name: "with github release",
			configData: map[string]any{
				"github": map[string]any{
					"release": true,
				},
			},
			wantGitHub: true,
		},
		{
			name: "with npm publish",
			configData: map[string]any{
				"npm": map[string]any{
					"publish": true,
				},
			},
			wantNPM: true,
		},
		{
			name: "with commit message",
			configData: map[string]any{
				"git": map[string]any{
					"commitMessage": "chore(release): ${version}",
				},
			},
			wantCommitMsg: "chore(release): {{.Version}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &detector.Result{
				Tool:       detector.ToolReleaseIt,
				ConfigFile: ".release-it.json",
				ConfigData: tt.configData,
			}

			config, err := Convert(result)
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			if config.Versioning.TagPrefix != tt.wantPrefix {
				t.Errorf("TagPrefix = %v, want %v", config.Versioning.TagPrefix, tt.wantPrefix)
			}

			hasGitHub := false
			hasNPM := false
			for _, p := range config.Plugins {
				if p.Name == "github" && p.Enabled {
					hasGitHub = true
				}
				if p.Name == "npm" && p.Enabled {
					hasNPM = true
				}
			}

			if hasGitHub != tt.wantGitHub {
				t.Errorf("GitHub plugin = %v, want %v", hasGitHub, tt.wantGitHub)
			}
			if hasNPM != tt.wantNPM {
				t.Errorf("NPM plugin = %v, want %v", hasNPM, tt.wantNPM)
			}
			if tt.wantCommitMsg != "" && config.Git.CommitMessage != tt.wantCommitMsg {
				t.Errorf("CommitMessage = %v, want %v", config.Git.CommitMessage, tt.wantCommitMsg)
			}
		})
	}
}

func TestConvert_StandardVersion(t *testing.T) {
	tests := []struct {
		name           string
		configData     map[string]any
		wantPrefix     string
		wantChangelog  bool
		wantCreateTag  bool
	}{
		{
			name: "basic config",
			configData: map[string]any{
				"tagPrefix": "v",
			},
			wantPrefix:    "v",
			wantChangelog: true,
			wantCreateTag: true,
		},
		{
			name: "skip changelog",
			configData: map[string]any{
				"skip": map[string]any{
					"changelog": true,
				},
			},
			wantChangelog: false,
			wantCreateTag: true,
		},
		{
			name: "skip tag",
			configData: map[string]any{
				"skip": map[string]any{
					"tag": true,
				},
			},
			wantChangelog: true,
			wantCreateTag: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &detector.Result{
				Tool:       detector.ToolStandardVersion,
				ConfigFile: ".versionrc.json",
				ConfigData: tt.configData,
			}

			config, err := Convert(result)
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			if config.Versioning.TagPrefix != tt.wantPrefix {
				t.Errorf("TagPrefix = %v, want %v", config.Versioning.TagPrefix, tt.wantPrefix)
			}
			if config.Changelog.Enabled != tt.wantChangelog {
				t.Errorf("Changelog.Enabled = %v, want %v", config.Changelog.Enabled, tt.wantChangelog)
			}
			if config.Git.CreateTag != tt.wantCreateTag {
				t.Errorf("Git.CreateTag = %v, want %v", config.Git.CreateTag, tt.wantCreateTag)
			}
		})
	}
}

func TestConvertTemplate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"${version}", "{{.Version}}"},
		{"v${version}", "v{{.Version}}"},
		{"chore(release): ${version}", "chore(release): {{.Version}}"},
		{"${nextRelease.version}", "{{.Version}}"},
		{"{{version}}", "{{.Version}}"},
		{"no template", "no template"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := convertTemplate(tt.input)
			if got != tt.want {
				t.Errorf("convertTemplate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvert_GoReleaser(t *testing.T) {
	tests := []struct {
		name           string
		configData     map[string]any
		wantPrefix     string
		wantChangelog  bool
		wantGitHub     bool
		wantDraft      bool
		wantPrerelease bool
		wantAssets     int
	}{
		{
			name:          "basic config",
			configData:    map[string]any{},
			wantPrefix:    "v",
			wantChangelog: true,
			wantGitHub:    true,
			wantAssets:    0,
		},
		{
			name: "with project name",
			configData: map[string]any{
				"project_name": "myapp",
			},
			wantPrefix:    "v",
			wantChangelog: true,
			wantGitHub:    true,
		},
		{
			name: "skip changelog",
			configData: map[string]any{
				"changelog": map[string]any{
					"skip": true,
				},
			},
			wantPrefix:    "v",
			wantChangelog: false,
			wantGitHub:    true,
		},
		{
			name: "with github release config",
			configData: map[string]any{
				"release": map[string]any{
					"github": map[string]any{
						"owner": "test-org",
						"name":  "test-repo",
					},
					"draft":      true,
					"prerelease": "auto",
				},
			},
			wantPrefix:     "v",
			wantChangelog:  true,
			wantGitHub:     true,
			wantDraft:      true,
			wantPrerelease: true,
		},
		{
			name: "with prerelease bool",
			configData: map[string]any{
				"release": map[string]any{
					"prerelease": true,
				},
			},
			wantPrefix:     "v",
			wantChangelog:  true,
			wantGitHub:     true,
			wantPrerelease: true,
		},
		{
			name: "with name template",
			configData: map[string]any{
				"release": map[string]any{
					"name_template": "{{.ProjectName}}-{{.Version}}",
				},
			},
			wantPrefix:    "v",
			wantChangelog: true,
			wantGitHub:    true,
		},
		{
			name: "with builds",
			configData: map[string]any{
				"project_name": "plugin-test",
				"builds": []any{
					map[string]any{
						"binary": "plugin-test",
						"goos":   []any{"linux", "darwin"},
						"goarch": []any{"amd64", "arm64"},
					},
				},
			},
			wantPrefix:    "v",
			wantChangelog: true,
			wantGitHub:    true,
			wantAssets:    5, // 2 OS * 2 arch + checksums
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &detector.Result{
				Tool:       detector.ToolGoReleaser,
				ConfigFile: ".goreleaser.yml",
				ConfigData: tt.configData,
			}

			config, err := Convert(result)
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			if config.Versioning.TagPrefix != tt.wantPrefix {
				t.Errorf("TagPrefix = %v, want %v", config.Versioning.TagPrefix, tt.wantPrefix)
			}

			if config.Changelog.Enabled != tt.wantChangelog {
				t.Errorf("Changelog.Enabled = %v, want %v", config.Changelog.Enabled, tt.wantChangelog)
			}

			hasGitHub := false
			var ghConfig map[string]any
			for _, p := range config.Plugins {
				if p.Name == "github" && p.Enabled {
					hasGitHub = true
					ghConfig = p.Config
				}
			}

			if hasGitHub != tt.wantGitHub {
				t.Errorf("GitHub plugin = %v, want %v", hasGitHub, tt.wantGitHub)
			}

			if tt.wantDraft && ghConfig != nil {
				if draft, ok := ghConfig["draft"].(bool); !ok || !draft {
					t.Errorf("draft = %v, want true", ghConfig["draft"])
				}
			}

			if tt.wantPrerelease && ghConfig != nil {
				if prerelease, ok := ghConfig["prerelease"].(bool); !ok || !prerelease {
					t.Errorf("prerelease = %v, want true", ghConfig["prerelease"])
				}
			}

			if tt.wantAssets > 0 && ghConfig != nil {
				if assets, ok := ghConfig["assets"].([]string); ok {
					if len(assets) != tt.wantAssets {
						t.Errorf("assets count = %v, want %v", len(assets), tt.wantAssets)
					}
				}
			}
		})
	}
}

func TestConvert_GoReleaser_Assets(t *testing.T) {
	result := &detector.Result{
		Tool:       detector.ToolGoReleaser,
		ConfigFile: ".goreleaser.yml",
		ConfigData: map[string]any{
			"project_name": "plugin-test",
			"builds": []any{
				map[string]any{
					"binary": "plugin-test",
					"goos":   []any{"linux", "darwin", "windows"},
					"goarch": []any{"amd64", "arm64"},
				},
			},
		},
	}

	config, err := Convert(result)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Find GitHub plugin
	var ghConfig map[string]any
	for _, p := range config.Plugins {
		if p.Name == "github" {
			ghConfig = p.Config
			break
		}
	}

	if ghConfig == nil {
		t.Fatal("GitHub plugin config not found")
	}

	assets, ok := ghConfig["assets"].([]string)
	if !ok {
		t.Fatal("assets should be []string")
	}

	// Should have 6 archives (3 OS * 2 arch) + checksums = 7
	if len(assets) != 7 {
		t.Errorf("assets count = %v, want 7", len(assets))
	}

	// Verify asset naming format
	expectedPatterns := []string{
		"release/plugin-test_linux_x86_64.tar.gz",
		"release/plugin-test_linux_aarch64.tar.gz",
		"release/plugin-test_darwin_x86_64.tar.gz",
		"release/plugin-test_darwin_aarch64.tar.gz",
		"release/plugin-test_windows_x86_64.zip",
		"release/plugin-test_windows_aarch64.zip",
		"release/checksums.txt",
	}

	for _, expected := range expectedPatterns {
		found := false
		for _, asset := range assets {
			if asset == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected asset %q not found in %v", expected, assets)
		}
	}
}

func TestConvert_GoReleaser_GitSettings(t *testing.T) {
	result := &detector.Result{
		Tool:       detector.ToolGoReleaser,
		ConfigFile: ".goreleaser.yml",
		ConfigData: map[string]any{},
	}

	config, err := Convert(result)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// Verify default git settings for GoReleaser
	if !config.Git.RequireCleanTree {
		t.Error("RequireCleanTree should be true")
	}
	if !config.Git.PushTags {
		t.Error("PushTags should be true")
	}
	if !config.Git.CreateTag {
		t.Error("CreateTag should be true")
	}
	if len(config.Git.AllowedBranches) != 1 || config.Git.AllowedBranches[0] != "main" {
		t.Errorf("AllowedBranches = %v, want [main]", config.Git.AllowedBranches)
	}
}

func TestExtractGoReleaserAssets(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]any
		projectName string
		wantCount   int
	}{
		{
			name:        "default targets",
			data:        map[string]any{},
			projectName: "myapp",
			wantCount:   7, // 3 OS * 2 arch + checksums
		},
		{
			name: "custom targets",
			data: map[string]any{
				"builds": []any{
					map[string]any{
						"goos":   []any{"linux"},
						"goarch": []any{"amd64"},
					},
				},
			},
			projectName: "myapp",
			wantCount:   2, // 1 OS * 1 arch + checksums
		},
		{
			name: "with binary name",
			data: map[string]any{
				"builds": []any{
					map[string]any{
						"binary": "custom-name",
						"goos":   []any{"linux", "darwin"},
						"goarch": []any{"amd64"},
					},
				},
			},
			projectName: "myapp",
			wantCount:   3, // 2 OS * 1 arch + checksums
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assets := extractGoReleaserAssets(tt.data, tt.projectName)
			if len(assets) != tt.wantCount {
				t.Errorf("extractGoReleaserAssets() count = %v, want %v", len(assets), tt.wantCount)
			}
		})
	}
}

func TestToStringSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []any
		want  []string
	}{
		{
			name:  "empty",
			input: []any{},
			want:  []string{},
		},
		{
			name:  "strings only",
			input: []any{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "mixed types",
			input: []any{"a", 123, "b", true, "c"},
			want:  []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toStringSlice(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("toStringSlice() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("toStringSlice()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestConvert_UnsupportedTool(t *testing.T) {
	result := &detector.Result{
		Tool: detector.ToolNone,
	}

	_, err := Convert(result)
	if err == nil {
		t.Error("Convert() should return error for unsupported tool")
	}
}
