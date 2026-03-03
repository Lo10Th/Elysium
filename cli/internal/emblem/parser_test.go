package emblem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid emblem",
			content: `apiVersion: v1
name: test-api
version: 1.0.0
description: A test API
baseUrl: https://api.example.com
actions:
  get-item:
    method: GET
    path: /items/{id}
    description: Get an item by ID
`,
			wantErr: false,
		},
		{
			name:    "missing file",
			content: "",
			wantErr: true,
		},
		{
			name: "invalid yaml",
			content: `apiVersion: v1
name: test-api
version: 1.0.0
description: A test API
baseUrl: https://api.example.com
actions:
  - this is not a valid action map
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "missing file" {
				_, err := Load("/nonexistent/path/emblem.yaml")
				if err == nil {
					t.Error("Load() expected error for missing file")
				}
				return
			}

			tmpDir := os.TempDir()
			tmpFile := filepath.Join(tmpDir, "test-emblem.yaml")
			defer os.Remove(tmpFile)

			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			def, err := Load(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && def == nil {
				t.Error("Load() returned nil definition")
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid emblem",
			yaml: `apiVersion: v1
name: test-api
version: 1.0.0
description: A test API
baseUrl: https://api.example.com
actions:
  get-item:
    method: GET
    path: /items/{id}
    description: Get an item
`,
			wantErr: false,
		},
		{
			name: "missing apiVersion",
			yaml: `name: test-api
version: 1.0.0
baseUrl: https://api.example.com
actions:
  get-item:
    method: GET
    path: /items
    description: Get items
`,
			wantErr: true,
		},
		{
			name: "invalid apiVersion",
			yaml: `apiVersion: v2
name: test-api
version: 1.0.0
baseUrl: https://api.example.com
actions:
  get-item:
    method: GET
    path: /items
    description: Get items
`,
			wantErr: true,
		},
		{
			name: "missing name",
			yaml: `apiVersion: v1
version: 1.0.0
baseUrl: https://api.example.com
actions:
  get-item:
    method: GET
    path: /items
    description: Get items
`,
			wantErr: true,
		},
		{
			name: "missing version",
			yaml: `apiVersion: v1
name: test-api
baseUrl: https://api.example.com
actions:
  get-item:
    method: GET
    path: /items
    description: Get items
`,
			wantErr: true,
		},
		{
			name: "missing baseUrl",
			yaml: `apiVersion: v1
name: test-api
version: 1.0.0
actions:
  get-item:
    method: GET
    path: /items
    description: Get items
`,
			wantErr: true,
		},
		{
			name: "missing actions",
			yaml: `apiVersion: v1
name: test-api
version: 1.0.0
baseUrl: https://api.example.com
`,
			wantErr: true,
		},
		{
			name: "action missing method",
			yaml: `apiVersion: v1
name: test-api
version: 1.0.0
baseUrl: https://api.example.com
actions:
  get-item:
    path: /items/{id}
    description: Get item
`,
			wantErr: true,
		},
		{
			name: "action missing path",
			yaml: `apiVersion: v1
name: test-api
version: 1.0.0
baseUrl: https://api.example.com
actions:
  get-item:
    method: GET
    description: Get item
`,
			wantErr: true,
		},
		{
			name: "action missing description",
			yaml: `apiVersion: v1
name: test-api
version: 1.0.0
baseUrl: https://api.example.com
actions:
  get-item:
    method: GET
    path: /items/{id}
`,
			wantErr: true,
		},
		{
			name: "full emblem with auth",
			yaml: `apiVersion: v1
name: secure-api
version: 2.0.0
description: A secure API
baseUrl: https://secure.example.com
author: Test Author
license: MIT
repository: https://github.com/test/secure-api
auth:
  type: bearer
  keyEnv: API_TOKEN
tags:
  - api
  - secure
category: utilities
actions:
  list-items:
    method: GET
    path: /items
    description: List all items
  get-item:
    method: GET
    path: /items/{id}
    description: Get item by ID
  create-item:
    method: POST
    path: /items
    description: Create a new item
`,
			wantErr: false,
		},
		{
			name: "invalid yaml syntax",
			yaml: `apiVersion: v1
name: test-api
version: 1.0.0
baseUrl: https://api.example.com
actions:
  get-item: {invalid yaml syntax
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := Parse([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if def == nil {
					t.Error("Parse() returned nil definition")
					return
				}
				if def.Name == "" && tt.name == "valid emblem" {
					t.Error("Parse() returned definition with empty name")
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		def     *Definition
		wantErr bool
	}{
		{
			name: "valid definition",
			def: &Definition{
				APIVersion: "v1",
				Name:       "test-api",
				Version:    "1.0.0",
				BaseURL:    "https://api.example.com",
				Actions: map[string]Action{
					"get-item": {
						Method:      "GET",
						Path:        "/items/{id}",
						Description: "Get an item",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid API version",
			def: &Definition{
				APIVersion: "v2",
				Name:       "test-api",
				Version:    "1.0.0",
				BaseURL:    "https://api.example.com",
				Actions: map[string]Action{
					"get": {Method: "GET", Path: "/", Description: "Get"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty name",
			def: &Definition{
				APIVersion: "v1",
				Name:       "",
				Version:    "1.0.0",
				BaseURL:    "https://api.example.com",
				Actions: map[string]Action{
					"get": {Method: "GET", Path: "/", Description: "Get"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty version",
			def: &Definition{
				APIVersion: "v1",
				Name:       "test-api",
				Version:    "",
				BaseURL:    "https://api.example.com",
				Actions: map[string]Action{
					"get": {Method: "GET", Path: "/", Description: "Get"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty baseURL",
			def: &Definition{
				APIVersion: "v1",
				Name:       "test-api",
				Version:    "1.0.0",
				BaseURL:    "",
				Actions: map[string]Action{
					"get": {Method: "GET", Path: "/", Description: "Get"},
				},
			},
			wantErr: true,
		},
		{
			name: "no actions",
			def: &Definition{
				APIVersion: "v1",
				Name:       "test-api",
				Version:    "1.0.0",
				BaseURL:    "https://api.example.com",
				Actions:    map[string]Action{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.def)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetAction(t *testing.T) {
	def := &Definition{
		APIVersion: "v1",
		Name:       "test-api",
		Version:    "1.0.0",
		BaseURL:    "https://api.example.com",
		Actions: map[string]Action{
			"get-item": {
				Method:      "GET",
				Path:        "/items/{id}",
				Description: "Get an item",
			},
			"list-items": {
				Method:      "GET",
				Path:        "/items",
				Description: "List all items",
			},
		},
	}

	t.Run("action exists", func(t *testing.T) {
		action, err := def.GetAction("get-item")
		if err != nil {
			t.Errorf("GetAction() error = %v", err)
		}
		if action == nil {
			t.Error("GetAction() returned nil")
		}
		if action != nil && action.Method != "GET" {
			t.Errorf("GetAction() Method = %v, want GET", action.Method)
		}
	})

	t.Run("action not found", func(t *testing.T) {
		_, err := def.GetAction("nonexistent")
		if err == nil {
			t.Error("GetAction() expected error for nonexistent action")
		}
	})
}

func TestListActions(t *testing.T) {
	def := &Definition{
		APIVersion: "v1",
		Name:       "test-api",
		Version:    "1.0.0",
		BaseURL:    "https://api.example.com",
		Actions: map[string]Action{
			"get-item":    {Method: "GET", Path: "/items/{id}", Description: "Get"},
			"list-items":  {Method: "GET", Path: "/items", Description: "List"},
			"create-item": {Method: "POST", Path: "/items", Description: "Create"},
		},
	}

	actions := def.ListActions()
	if len(actions) != 3 {
		t.Errorf("ListActions() returned %d actions, want 3", len(actions))
	}
}

func TestGetAuthCredentials(t *testing.T) {
	tests := []struct {
		name      string
		auth      Auth
		envKey    string
		envValue  string
		wantErr   bool
		wantCreds map[string]string
	}{
		{
			name:      "no auth",
			auth:      Auth{Type: AuthNone},
			wantCreds: map[string]string{},
			wantErr:   false,
		},
		{
			name:      "api key auth",
			auth:      Auth{Type: AuthAPIKey, KeyEnv: "API_KEY", Header: "X-Custom-Key"},
			envKey:    "API_KEY",
			envValue:  "test-key-123",
			wantCreds: map[string]string{"value": "test-key-123", "header": "X-Custom-Key"},
			wantErr:   false,
		},
		{
			name:      "api key auth default header",
			auth:      Auth{Type: AuthAPIKey, KeyEnv: "API_KEY"},
			envKey:    "API_KEY",
			envValue:  "test-key-123",
			wantCreds: map[string]string{"value": "test-key-123", "header": "X-API-Key"},
			wantErr:   false,
		},
		{
			name:      "bearer auth",
			auth:      Auth{Type: AuthBearer, KeyEnv: "BEARER_TOKEN"},
			envKey:    "BEARER_TOKEN",
			envValue:  "my-token",
			wantCreds: map[string]string{"value": "my-token", "header": "Authorization", "prefix": "Bearer "},
			wantErr:   false,
		},
		{
			name:      "basic auth",
			auth:      Auth{Type: AuthBasic, KeyEnv: "BASIC_CREDS"},
			envKey:    "BASIC_CREDS",
			envValue:  "base64creds",
			wantCreds: map[string]string{"value": "base64creds", "header": "Authorization", "prefix": "Basic "},
			wantErr:   false,
		},
		{
			name:     "missing env var",
			auth:     Auth{Type: AuthAPIKey, KeyEnv: "MISSING_KEY"},
			envKey:   "",
			envValue: "",
			wantErr:  true,
		},
		{
			name:     "auth type requires keyEnv",
			auth:     Auth{Type: AuthAPIKey, KeyEnv: ""},
			envKey:   "",
			envValue: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			def := &Definition{
				APIVersion: "v1",
				Name:       "test",
				Version:    "1.0.0",
				BaseURL:    "https://api.example.com",
				Auth:       tt.auth,
				Actions:    map[string]Action{"get": {Method: "GET", Path: "/", Description: "Test"}},
			}

			creds, err := def.GetAuthCredentials()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAuthCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for k, v := range tt.wantCreds {
					if creds[k] != v {
						t.Errorf("GetAuthCredentials()[%s] = %v, want %v", k, creds[k], v)
					}
				}
			}
		})
	}
}

func TestSaveToCacheAndLoadFromCache(t *testing.T) {
	tmpDir := os.TempDir()
	cacheDir := filepath.Join(tmpDir, ".elysium-cache-test")
	defer os.RemoveAll(cacheDir)

	os.Setenv("HOME", cacheDir)

	validYAML := `apiVersion: v1
name: cache-test
version: 1.0.0
description: Cache test
baseUrl: https://api.example.com
actions:
  test:
    method: GET
    path: /test
    description: Test action
`

	err := SaveToCache("cache-test", "1.0.0", []byte(validYAML))
	if err != nil {
		t.Errorf("SaveToCache() error = %v", err)
		return
	}

	def, err := LoadFromCache("cache-test", "1.0.0")
	if err != nil {
		t.Errorf("LoadFromCache() error = %v", err)
		return
	}

	if def == nil {
		t.Error("LoadFromCache() returned nil")
		return
	}

	if def.Name != "cache-test" {
		t.Errorf("LoadFromCache() Name = %v, want 'cache-test'", def.Name)
	}

	_, err = LoadFromCache("nonexistent", "1.0.0")
	if err == nil {
		t.Error("LoadFromCache() expected error for nonexistent cache")
	}
}

func TestParseVersionConstraint(t *testing.T) {
	tests := []struct {
		name        string
		constraint  string
		wantName    string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "name only",
			constraint:  "my-emblem",
			wantName:    "my-emblem",
			wantVersion: "latest",
			wantErr:     false,
		},
		{
			name:        "name with version",
			constraint:  "my-emblem@1.2.3",
			wantName:    "my-emblem",
			wantVersion: "1.2.3",
			wantErr:     false,
		},
		{
			name:        "too many parts",
			constraint:  "my@emblem@version",
			wantName:    "",
			wantVersion: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, version, err := ParseVersionConstraint(tt.constraint)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersionConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if name != tt.wantName {
					t.Errorf("ParseVersionConstraint() name = %v, want %v", name, tt.wantName)
				}
				if version != tt.wantVersion {
					t.Errorf("ParseVersionConstraint() version = %v, want %v", version, tt.wantVersion)
				}
			}
		})
	}
}

func TestGetCachePath(t *testing.T) {
	path, err := GetCachePath("test-emblem", "1.0.0")
	if err != nil {
		t.Errorf("GetCachePath() error = %v", err)
		return
	}

	if path == "" {
		t.Error("GetCachePath() returned empty path")
		return
	}

	if !filepath.IsAbs(path) {
		t.Errorf("GetCachePath() returned relative path: %v", path)
	}
}
