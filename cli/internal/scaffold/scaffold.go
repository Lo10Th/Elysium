package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

type EmblemTemplate struct {
	Name        string
	Category    string
	Description string
	Version     string
	BaseURL     string
	Actions     []ActionTemplate
}

type ActionTemplate struct {
	Name        string
	Method      string
	Path        string
	Description string
}

func GetCategoryTemplate(category string) []ActionTemplate {
	switch category {
	case "payments":
		return []ActionTemplate{
			{"create-payment", "POST", "/payments", "Create a new payment"},
			{"list-payments", "GET", "/payments", "List all payments"},
			{"get-payment", "GET", "/payments/{id}", "Get payment by ID"},
			{"delete-payment", "DELETE", "/payments/{id}", "Delete a payment"},
		}
	case "ecommerce":
		return []ActionTemplate{
			{"list-products", "GET", "/products", "List all products"},
			{"create-product", "POST", "/products", "Create a product"},
			{"get-product", "GET", "/products/{id}", "Get product by ID"},
			{"update-product", "PUT", "/products/{id}", "Update a product"},
			{"delete-product", "DELETE", "/products/{id}", "Delete a product"},
		}
	case "auth":
		return []ActionTemplate{
			{"login", "POST", "/auth/login", "User login"},
			{"logout", "POST", "/auth/logout", "User logout"},
			{"register", "POST", "/auth/register", "Register new user"},
			{"refresh", "POST", "/auth/refresh", "Refresh authentication token"},
		}
	default:
		return []ActionTemplate{
			{"list", "GET", "/items", "List items"},
			{"create", "POST", "/items", "Create an item"},
			{"get", "GET", "/items/{id}", "Get item by ID"},
			{"update", "PUT", "/items/{id}", "Update an item"},
			{"delete", "DELETE", "/items/{id}", "Delete an item"},
		}
	}
}

func ValidateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("name cannot exceed 63 characters")
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return fmt.Errorf("name must be lowercase alphanumeric with dashes (found: %c)", c)
		}
	}
	return nil
}

func CreateDirectories(name string) error {
	dirs := []string{
		name,
		filepath.Join(name, "examples"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func GenerateEmblem(tmpl EmblemTemplate, outputPath string) error {
	tmplContent := `apiVersion: v1
name: {{.Name}}
version: {{.Version}}
description: "{{.Description}}"
baseUrl: {{.BaseURL}}

auth:
  type: api_key
  location: header
  key_name: X-API-KEY

actions:
{{range .Actions}}
  {{.Name}}:
    method: {{.Method}}
    path: {{.Path}}
    description: "{{.Description}}"
{{end}}
`

	t, err := template.New("emblem").Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if err := t.Execute(f, tmpl); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func GenerateREADME(tmpl EmblemTemplate, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	content := fmt.Sprintf(`# %s

%s

## Installation

`+"```"+`
ely pull %s
`+"```"+`

## Usage

### Available Actions

`, tmpl.Name, tmpl.Description, tmpl.Name)

	for _, action := range tmpl.Actions {
		content += fmt.Sprintf(`#### %s

%s

`+"```"+`
ely %s %s
`+"```"+`

`, action.Name, action.Description, tmpl.Name, action.Name)
	}

	content += `## Development

1. Edit ` + "`emblem.yaml`" + ` to customize
2. Validate: ` + "`ely validate emblem.yaml`" + `
3. Test: ` + "`ely run . --local`" + `
`

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func GenerateExamples(name, category string, outputDir string) error {
	examplePath := filepath.Join(outputDir, "example.json")

	content := fmt.Sprintf(`{
  "id": 1,
  "name": "Example %s",
  "created_at": "%s"
}`, category, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(examplePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write example: %w", err)
	}

	return nil
}

func GenerateIgnoreFile(name string) error {
	content := `# Dependencies
node_modules/
vendor/

# Build outputs
dist/
build/
*.exe

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db

# Temp
*.tmp
*.log
`

	if err := os.WriteFile(filepath.Join(name, ".emblemignore"), []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .emblemignore: %w", err)
	}

	return nil
}
