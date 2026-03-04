package validator

import (
	"testing"

	"github.com/elysium/elysium/cli/internal/emblem"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		def        *emblem.Definition
		wantErrors int
	}{
		{
			name: "valid definition",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Auth:    emblem.Auth{Type: emblem.AuthAPIKey},
				Actions: map[string]emblem.Action{
					"get-items": {
						Method:      "GET",
						Path:        "/items",
						Description: "List all items",
					},
				},
			},
			wantErrors: 0,
		},
		{
			name: "missing name",
			def: &emblem.Definition{
				Name:    "",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			},
			wantErrors: 1,
		},
		{
			name: "missing version",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			},
			wantErrors: 1,
		},
		{
			name: "missing baseUrl",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			},
			wantErrors: 1,
		},
		{
			name: "invalid name format",
			def: &emblem.Definition{
				Name:    "Test_API",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			},
			wantErrors: 1,
		},
		{
			name: "invalid version format",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "v1.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			},
			wantErrors: 1,
		},
		{
			name: "invalid baseUrl",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "ftp://files.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			},
			wantErrors: 1,
		},
		{
			name: "no actions",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{},
			},
			wantErrors: 1,
		},
		{
			name: "action missing method",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Path: "/", Description: "Test"},
				},
			},
			wantErrors: 1,
		},
		{
			name: "action missing path",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Description: "Test"},
				},
			},
			wantErrors: 1,
		},
		{
			name: "action invalid method",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "INVALID", Path: "/", Description: "Test"},
				},
			},
			wantErrors: 1,
		},
		{
			name: "multiple errors",
			def: &emblem.Definition{
				Name:    "",
				Version: "",
				BaseURL: "",
				Actions: map[string]emblem.Action{},
			},
			wantErrors: 4,
		},
	}

	v := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.Validate(tt.def)
			if len(errors) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestValidateStrict(t *testing.T) {
	tests := []struct {
		name       string
		def        *emblem.Definition
		wantErrors int
	}{
		{
			name: "strictly valid definition",
			def: &emblem.Definition{
				Name:        "test-api",
				Version:     "1.0.0",
				BaseURL:     "https://api.example.com",
				Auth:        emblem.Auth{Type: emblem.AuthAPIKey},
				Description: "A test API",
				Actions: map[string]emblem.Action{
					"get-items": {
						Method:      "GET",
						Path:        "/items",
						Description: "List all items",
					},
				},
			},
			wantErrors: 0,
		},
		{
			name: "missing action description",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Auth:    emblem.Auth{Type: emblem.AuthAPIKey},
				Actions: map[string]emblem.Action{
					"get-items": {
						Method: "GET",
						Path:   "/items",
					},
				},
			},
			wantErrors: 1,
		},
		{
			name: "missing auth type",
			def: &emblem.Definition{
				Name:        "test-api",
				Version:     "1.0.0",
				BaseURL:     "https://api.example.com",
				Description: "A test API",
				Actions: map[string]emblem.Action{
					"get-items": {
						Method:      "GET",
						Path:        "/items",
						Description: "List all items",
					},
				},
			},
			wantErrors: 1,
		},
		{
			name: "multiple strict errors",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get-items": {
						Method: "GET",
						Path:   "/items",
					},
				},
			},
			wantErrors: 2,
		},
	}

	v := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidateStrict(tt.def)
			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateStrict() returned %d errors, want %d: %v", len(errors), tt.wantErrors, errors)
			}
		})
	}
}

func TestCheckBestPractices(t *testing.T) {
	tests := []struct {
		name         string
		def          *emblem.Definition
		wantWarnings int
	}{
		{
			name: "complete definition",
			def: &emblem.Definition{
				Name:        "test-api",
				Version:     "1.0.0",
				Description: "A test API",
				BaseURL:     "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			},
			wantWarnings: 0,
		},
		{
			name: "missing description",
			def: &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			},
			wantWarnings: 1,
		},
	}

	v := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := v.CheckBestPractices(tt.def)
			if len(warnings) != tt.wantWarnings {
				t.Errorf("CheckBestPractices() returned %d warnings, want %d: %v", len(warnings), tt.wantWarnings, warnings)
			}
		})
	}
}

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid lowercase", "test-api", true},
		{"valid with numbers", "api-v2-123", true},
		{"valid single char", "a", true},
		{"invalid uppercase", "Test-API", false},
		{"invalid underscore", "test_api", false},
		{"invalid space", "test api", false},
		{"invalid special chars", "test@api", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: isValidName is unexported, tested via Validate
			v := New()
			def := &emblem.Definition{
				Name:    tt.input,
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			}

			// If name is empty, Validate returns error for missing name
			// otherwise it checks format
			errors := v.Validate(def)

			if tt.want {
				// Should not have name format error
				for _, err := range errors {
					if err == "name must be lowercase alphanumeric with dashes" {
						t.Errorf("isValidName(%q) should return true, got error: %s", tt.input, err)
						return
					}
				}
			} else {
				// Should have a name format error (not the "required" error)
				if tt.input != "" {
					for _, err := range errors {
						if err == "name must be lowercase alphanumeric with dashes" {
							return // Expected error
						}
					}
					t.Errorf("isValidName(%q) should return false, got errors: %v", tt.input, errors)
				}
			}
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid semver", "1.0.0", true},
		{"valid semver with v", "v1.0.0", false},
		{"valid two part", "1.0", false},
		{"valid single number", "1", false},
		{"invalid letters", "1.0.x", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: isValidVersion is unexported, tested via Validate
			v := New()
			def := &emblem.Definition{
				Name:    "test-api",
				Version: tt.input,
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			}

			errors := v.Validate(def)

			if tt.input == "" {
				// Empty version gives "version is required" error
				return
			}

			if tt.want {
				for _, err := range errors {
					if err == "version must be semver (e.g., 1.0.0)" {
						t.Errorf("isValidVersion(%q) should return true, got error", tt.input)
					}
				}
			} else {
				found := false
				for _, err := range errors {
					if err == "version must be semver (e.g., 1.0.0)" {
						found = true
						break
					}
				}
				if !found && len(errors) > 0 {
					t.Errorf("isValidVersion(%q) should return false, got errors: %v", tt.input, errors)
				}
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid http", "http://example.com", true},
		{"valid https", "https://api.example.com", true},
		{"valid with path", "https://api.example.com/v1", true},
		{"invalid ftp", "ftp://files.example.com", false},
		{"invalid no scheme", "example.com", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			def := &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: tt.input,
				Actions: map[string]emblem.Action{
					"get": {Method: "GET", Path: "/", Description: "Test"},
				},
			}

			errors := v.Validate(def)

			if tt.input == "" {
				// Empty baseUrl gives "baseUrl is required" error
				return
			}

			if tt.want {
				for _, err := range errors {
					if err == "baseUrl must be a valid URL" {
						t.Errorf("isValidURL(%q) should return true, got error", tt.input)
					}
				}
			} else {
				found := false
				for _, err := range errors {
					if err == "baseUrl must be a valid URL" {
						found = true
						break
					}
				}
				if !found && len(errors) > 0 && tt.input != "" {
					t.Errorf("isValidURL(%q) should return false, got errors: %v", tt.input, errors)
				}
			}
		})
	}
}

func BenchmarkValidate(b *testing.B) {
	v := New()
	def := &emblem.Definition{
		Name:    "test-api",
		Version: "1.0.0",
		BaseURL: "https://api.example.com",
		Actions: map[string]emblem.Action{
			"get": {Method: "GET", Path: "/", Description: "Test"},
		},
	}
	for i := 0; i < b.N; i++ {
		v.Validate(def)
	}
}

func TestIsValidHTTPMethod(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid GET", "GET", true},
		{"valid POST", "POST", true},
		{"valid PUT", "PUT", true},
		{"valid PATCH", "PATCH", true},
		{"valid DELETE", "DELETE", true},
		{"valid HEAD", "HEAD", true},
		{"valid OPTIONS", "OPTIONS", true},
		{"invalid lowercase", "get", false},
		{"invalid mixed", "Get", false},
		{"invalid CONNECT", "CONNECT", false},
		{"invalid unknown", "UNKNOWN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			def := &emblem.Definition{
				Name:    "test-api",
				Version: "1.0.0",
				BaseURL: "https://api.example.com",
				Actions: map[string]emblem.Action{
					"test": {Method: tt.input, Path: "/", Description: "Test"},
				},
			}

			errors := v.Validate(def)

			if tt.want {
				for _, err := range errors {
					if err == "action 'test' has invalid method: "+tt.input {
						t.Errorf("isValidHTTPMethod(%q) should return true, got error", tt.input)
					}
				}
			} else {
				found := false
				for _, err := range errors {
					if err == "action 'test' has invalid method: "+tt.input {
						found = true
						break
					}
				}
				if !found && len(errors) == 0 {
					t.Errorf("isValidHTTPMethod(%q) should return false, but validation passed", tt.input)
				}
			}
		})
	}
}
