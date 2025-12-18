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

func TestConvert_UnsupportedTool(t *testing.T) {
	result := &detector.Result{
		Tool: detector.ToolNone,
	}

	_, err := Convert(result)
	if err == nil {
		t.Error("Convert() should return error for unsupported tool")
	}
}
