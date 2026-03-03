package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetCategoryTemplate(t *testing.T) {
	tests := []struct {
		name            string
		category        string
		wantActionCount int
		wantFirstAction string
	}{
		{"payments category", "payments", 4, "create-payment"},
		{"ecommerce category", "ecommerce", 5, "list-products"},
		{"auth category", "auth", 4, "login"},
		{"unknown category defaults to CRUD", "unknown", 5, "list"},
		{"empty category defaults to CRUD", "", 5, "list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := GetCategoryTemplate(tt.category)
			if len(actions) != tt.wantActionCount {
				t.Errorf("GetCategoryTemplate(%q) returned %d actions, want %d", tt.category, len(actions), tt.wantActionCount)
			}
			if len(actions) > 0 && actions[0].Name != tt.wantFirstAction {
				t.Errorf("GetCategoryTemplate(%q) first action = %q, want %q", tt.category, actions[0].Name, tt.wantFirstAction)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid lowercase", "test-api", false},
		{"valid with numbers", "api-v2-123", false},
		{"valid single char", "a", false},
		{"valid short", "my-api", false},
		{"valid max length", strings.Repeat("a", 63), false},
		{"invalid uppercase", "Test-API", true},
		{"invalid underscore", "test_api", true},
		{"invalid space", "test api", true},
		{"invalid special chars", "test@api", true},
		{"invalid dots", "test.api", true},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 64), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestCreateDirectories(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
	}{
		{
			name: "creates directories successfully",
			setup: func() string {
				dir := filepath.Join(os.TempDir(), "scaffold-test-"+strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-")))
				return dir
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup()
			defer os.RemoveAll(dir)

			err := CreateDirectories(dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDirectories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Error("CreateDirectories() did not create main directory")
				}
				if _, err := os.Stat(filepath.Join(dir, "examples")); os.IsNotExist(err) {
					t.Error("CreateDirectories() did not create examples subdirectory")
				}
			}
		})
	}
}

func TestGenerateEmblem(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    EmblemTemplate
		wantErr bool
	}{
		{
			name: "generates valid emblem",
			tmpl: EmblemTemplate{
				Name:        "test-api",
				Version:     "1.0.0",
				Description: "A test API",
				BaseURL:     "https://api.example.com",
				Actions: []ActionTemplate{
					{Name: "list", Method: "GET", Path: "/items", Description: "List items"},
				},
			},
			wantErr: false,
		},
		{
			name: "generates emblem with multiple actions",
			tmpl: EmblemTemplate{
				Name:        "complex-api",
				Version:     "2.0.0",
				Description: "Complex API",
				BaseURL:     "https://api.example.com",
				Actions: []ActionTemplate{
					{Name: "list", Method: "GET", Path: "/items", Description: "List"},
					{Name: "get", Method: "GET", Path: "/items/{id}", Description: "Get"},
					{Name: "create", Method: "POST", Path: "/items", Description: "Create"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := os.TempDir()
			outputPath := filepath.Join(tmpDir, "test-emblem.yaml")
			defer os.Remove(outputPath)

			err := GenerateEmblem(tt.tmpl, outputPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEmblem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Error("GenerateEmblem() did not create file")
					return
				}

				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatalf("Failed to read generated file: %v", err)
				}

				if !strings.Contains(string(content), "apiVersion: v1") {
					t.Error("GenerateEmblem() missing apiVersion in output")
				}
				if !strings.Contains(string(content), tt.tmpl.Name) {
					t.Errorf("GenerateEmblem() missing name %q in output", tt.tmpl.Name)
				}
			}
		})
	}
}

func TestGenerateREADME(t *testing.T) {
	tmpl := EmblemTemplate{
		Name:        "test-api",
		Version:     "1.0.0",
		Description: "A test API for testing",
		BaseURL:     "https://api.example.com",
		Actions: []ActionTemplate{
			{Name: "list", Method: "GET", Path: "/items", Description: "List all items"},
			{Name: "get", Method: "GET", Path: "/items/{id}", Description: "Get item by ID"},
		},
	}

	tmpDir := os.TempDir()
	outputPath := filepath.Join(tmpDir, "README.md")
	defer os.Remove(outputPath)

	err := GenerateREADME(tmpl, outputPath)
	if err != nil {
		t.Errorf("GenerateREADME() error = %v", err)
		return
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated README: %v", err)
	}

	contentStr := string(content)

	checks := []struct {
		name    string
		contain string
	}{
		{"name header", "# test-api"},
		{"description", "A test API for testing"},
		{"installation", "ely pull test-api"},
		{"list action", "list"},
		{"get action", "get"},
	}

	for _, check := range checks {
		if !strings.Contains(contentStr, check.contain) {
			t.Errorf("GenerateREADME() missing %s in output (expected %q)", check.name, check.contain)
		}
	}
}

func TestGenerateExamples(t *testing.T) {
	name := "test-emblem"
	category := "payments"
	outputDir := os.TempDir()

	err := GenerateExamples(name, category, outputDir)
	if err != nil {
		t.Errorf("GenerateExamples() error = %v", err)
		return
	}

	examplePath := filepath.Join(outputDir, "example.json")
	defer os.Remove(examplePath)

	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Error("GenerateExamples() did not create example.json")
		return
	}

	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("Failed to read example.json: %v", err)
	}

	if !strings.Contains(string(content), `"id":`) {
		t.Error("GenerateExamples() example.json missing id field")
	}
	if !strings.Contains(string(content), category) {
		t.Errorf("GenerateExamples() example.json missing category %q", category)
	}
}

func TestGenerateIgnoreFile(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "scaffold-ignore-test")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	err := GenerateIgnoreFile(tmpDir)
	if err != nil {
		t.Errorf("GenerateIgnoreFile() error = %v", err)
		return
	}

	ignorePath := filepath.Join(tmpDir, ".emblemignore")
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		t.Error("GenerateIgnoreFile() did not create .emblemignore")
		return
	}

	content, err := os.ReadFile(ignorePath)
	if err != nil {
		t.Fatalf("Failed to read .emblemignore: %v", err)
	}

	contentStr := string(content)

	checks := []string{
		"node_modules/",
		"vendor/",
		"dist/",
		".idea/",
		".vscode/",
		".DS_Store",
	}

	for _, check := range checks {
		if !strings.Contains(contentStr, check) {
			t.Errorf("GenerateIgnoreFile() missing %q in output", check)
		}
	}
}
