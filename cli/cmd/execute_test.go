package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsInstalledEmblem(t *testing.T) {
	t.Run("emblem installed", func(t *testing.T) {
		home, _ := os.UserHomeDir()
		cacheDir := filepath.Join(home, ".elysium", "cache", "test-emblem@1.0.0")
		os.MkdirAll(cacheDir, 0755)

		emblemContent := `apiVersion: v1
name: test-emblem
version: 1.0.0
description: Test emblem
baseUrl: http://localhost:5000/api
auth:
  type: none
actions:
  list-products:
    description: List products
    method: GET
    path: /products
`
		os.WriteFile(filepath.Join(cacheDir, "emblem.yaml"), []byte(emblemContent), 0644)
		defer os.RemoveAll(filepath.Join(home, ".elysium", "cache", "test-emblem@1.0.0"))

		configPath := filepath.Join(home, ".elysium", "config.yaml")
		configContent := `registry: https://ely.karlharrenga.com
cache_dir: ~/.elysium/cache
installed:
  test-emblem: 1.0.0
`
		os.WriteFile(configPath, []byte(configContent), 0644)
		defer os.Remove(configPath)

		result := isInstalledEmblem("test-emblem")
		if !result {
			t.Errorf("isInstalledEmblem(test-emblem) = false, want true")
		}
	})

	t.Run("emblem not installed", func(t *testing.T) {
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".elysium", "config.yaml")
		configContent := `registry: https://ely.karlharrenga.com
cache_dir: ~/.elysium/cache
installed:
  test-emblem: 1.0.0
`
		os.WriteFile(configPath, []byte(configContent), 0644)
		defer os.Remove(configPath)

		result := isInstalledEmblem("nonexistent-emblem")
		if result {
			t.Errorf("isInstalledEmblem(nonexistent-emblem) = true, want false")
		}
	})
}

func TestParseParams(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		filePath string
		args     []string
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "empty params",
			args:     []string{},
			expected: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "simple flag",
			args:     []string{"--category", "shoes"},
			expected: map[string]interface{}{"category": "shoes"},
			wantErr:  false,
		},
		{
			name:     "flag with equals",
			args:     []string{"category=shoes"},
			expected: map[string]interface{}{"category": "shoes"},
			wantErr:  false,
		},
		{
			name:     "json params",
			jsonStr:  `{"category": "shoes", "limit": 10}`,
			args:     []string{},
			expected: map[string]interface{}{"category": "shoes", "limit": float64(10)},
			wantErr:  false,
		},
		{
			name:     "multiple flags",
			args:     []string{"--category", "shoes", "--limit", "10"},
			expected: map[string]interface{}{"category": "shoes", "limit": "10"},
			wantErr:  false,
		},
		{
			name:     "short flag",
			args:     []string{"-c", "shoes"},
			expected: map[string]interface{}{"c": "shoes"},
			wantErr:  false,
		},
		{
			name:     "boolean flag",
			args:     []string{"--verbose"},
			expected: map[string]interface{}{"verbose": "true"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paramsJSON = tt.jsonStr
			paramsFile = tt.filePath

			result, err := parseParams(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseParams() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseParams() unexpected error: %v", err)
				return
			}

			for key, expectedVal := range tt.expected {
				actualVal, exists := result[key]
				if !exists {
					t.Errorf("parseParams() missing key %s", key)
					return
				}
				if actualVal != expectedVal {
					t.Errorf("parseParams()[%s] = %v, want %v", key, actualVal, expectedVal)
				}
			}

			paramsJSON = ""
			paramsFile = ""
		})
	}
}

func TestIsJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"{\"key\": \"value\"}", true},
		{"[1, 2, 3]", true},
		{"plain text", false},
		{"", false},
		{"   {", true},
		{"   [", true},
	}

	for _, tt := range tests {
		result := isJSON(tt.input)
		if result != tt.expected {
			t.Errorf("isJSON(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
