package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect_SemanticRelease(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		wantTool   Tool
		wantConfig string
	}{
		{
			name: "releaserc.json",
			files: map[string]string{
				".releaserc.json": `{"branches": ["main"], "plugins": ["@semantic-release/github"]}`,
			},
			wantTool:   ToolSemanticRelease,
			wantConfig: ".releaserc.json",
		},
		{
			name: "releaserc.yaml",
			files: map[string]string{
				".releaserc.yaml": "branches:\n  - main\nplugins:\n  - '@semantic-release/github'",
			},
			wantTool:   ToolSemanticRelease,
			wantConfig: ".releaserc.yaml",
		},
		{
			name: "package.json with release key",
			files: map[string]string{
				"package.json": `{"name": "test", "release": {"branches": ["main"]}}`,
			},
			wantTool:   ToolSemanticRelease,
			wantConfig: "package.json (release key)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for filename, content := range tt.files {
				path := filepath.Join(dir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			result, err := Detect(dir)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if result.Tool != tt.wantTool {
				t.Errorf("Detect() tool = %v, want %v", result.Tool, tt.wantTool)
			}

			if !contains(result.ConfigFile, tt.wantConfig) {
				t.Errorf("Detect() configFile = %v, want to contain %v", result.ConfigFile, tt.wantConfig)
			}
		})
	}
}

func TestDetect_ReleaseIt(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		wantTool   Tool
		wantConfig string
	}{
		{
			name: "release-it.json",
			files: map[string]string{
				".release-it.json": `{"git": {"tagName": "v${version}"}, "github": {"release": true}}`,
			},
			wantTool:   ToolReleaseIt,
			wantConfig: ".release-it.json",
		},
		{
			name: "release-it.yaml",
			files: map[string]string{
				".release-it.yaml": "git:\n  tagName: v${version}\ngithub:\n  release: true",
			},
			wantTool:   ToolReleaseIt,
			wantConfig: ".release-it.yaml",
		},
		{
			name: "package.json with release-it key",
			files: map[string]string{
				"package.json": `{"name": "test", "release-it": {"git": {"tagName": "v${version}"}}}`,
			},
			wantTool:   ToolReleaseIt,
			wantConfig: "package.json (release-it key)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for filename, content := range tt.files {
				path := filepath.Join(dir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			result, err := Detect(dir)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if result.Tool != tt.wantTool {
				t.Errorf("Detect() tool = %v, want %v", result.Tool, tt.wantTool)
			}

			if !contains(result.ConfigFile, tt.wantConfig) {
				t.Errorf("Detect() configFile = %v, want to contain %v", result.ConfigFile, tt.wantConfig)
			}
		})
	}
}

func TestDetect_StandardVersion(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		wantTool   Tool
		wantConfig string
	}{
		{
			name: "versionrc.json",
			files: map[string]string{
				".versionrc.json": `{"tagPrefix": "v", "skip": {"changelog": false}}`,
			},
			wantTool:   ToolStandardVersion,
			wantConfig: ".versionrc.json",
		},
		{
			name: "versionrc (no extension)",
			files: map[string]string{
				".versionrc": `{"tagPrefix": "v"}`,
			},
			wantTool:   ToolStandardVersion,
			wantConfig: ".versionrc",
		},
		{
			name: "package.json with standard-version key",
			files: map[string]string{
				"package.json": `{"name": "test", "standard-version": {"tagPrefix": "v"}}`,
			},
			wantTool:   ToolStandardVersion,
			wantConfig: "package.json (standard-version key)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for filename, content := range tt.files {
				path := filepath.Join(dir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			result, err := Detect(dir)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if result.Tool != tt.wantTool {
				t.Errorf("Detect() tool = %v, want %v", result.Tool, tt.wantTool)
			}

			if !contains(result.ConfigFile, tt.wantConfig) {
				t.Errorf("Detect() configFile = %v, want to contain %v", result.ConfigFile, tt.wantConfig)
			}
		})
	}
}

func TestDetect_NoConfig(t *testing.T) {
	dir := t.TempDir()

	result, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if result.Tool != ToolNone {
		t.Errorf("Detect() tool = %v, want %v", result.Tool, ToolNone)
	}
}

func TestDetect_Priority(t *testing.T) {
	// When multiple configs exist, semantic-release should take priority
	dir := t.TempDir()

	files := map[string]string{
		".releaserc.json":   `{"branches": ["main"]}`,
		".release-it.json":  `{"git": {"tagName": "v${version}"}}`,
		".versionrc.json":   `{"tagPrefix": "v"}`,
	}

	for filename, content := range files {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
	}

	result, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if result.Tool != ToolSemanticRelease {
		t.Errorf("Detect() tool = %v, want %v (should prioritize semantic-release)", result.Tool, ToolSemanticRelease)
	}
}

func contains(s, substr string) bool {
	return filepath.Base(s) == substr || s == substr ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr)
}
