package emblem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type AuthType string

const (
	AuthNone   AuthType = "none"
	AuthAPIKey AuthType = "api_key"
	AuthBearer AuthType = "bearer"
	AuthBasic  AuthType = "basic"
	AuthOAuth2 AuthType = "oauth2"
)

type Auth struct {
	Type   AuthType `yaml:"type"`
	KeyEnv string   `yaml:"keyEnv"`
	Header string   `yaml:"header,omitempty"`
	Prefix string   `yaml:"prefix,omitempty"`
}

type Property struct {
	Type        string              `yaml:"type"`
	Description string              `yaml:"description,omitempty"`
	Required    bool                `yaml:"required,omitempty"`
	Default     interface{}         `yaml:"default,omitempty"`
	Items       *Property           `yaml:"items,omitempty"`
	Properties  map[string]Property `yaml:"properties,omitempty"`
	Enum        []interface{}       `yaml:"enum,omitempty"`
}

type TypeDefinition struct {
	Description string              `yaml:"description,omitempty"`
	Properties  map[string]Property `yaml:"properties"`
}

type Parameter struct {
	Name        string              `yaml:"name"`
	Type        string              `yaml:"type"`
	In          string              `yaml:"in"`
	Required    bool                `yaml:"required"`
	Description string              `yaml:"description,omitempty"`
	Default     interface{}         `yaml:"default,omitempty"`
	Items       *Property           `yaml:"items,omitempty"`
	Properties  map[string]Property `yaml:"properties,omitempty"`
	Enum        []interface{}       `yaml:"enum,omitempty"`
}

type Response struct {
	Description string      `yaml:"description"`
	Schema      interface{} `yaml:"schema,omitempty"`
}

type Action struct {
	Description string           `yaml:"description"`
	Method      string           `yaml:"method"`
	Path        string           `yaml:"path"`
	Parameters  []Parameter      `yaml:"parameters,omitempty"`
	RequestBody interface{}      `yaml:"requestBody,omitempty"`
	Responses   map[int]Response `yaml:"responses,omitempty"`
	Errors      []interface{}    `yaml:"errors,omitempty"`
}

type Definition struct {
	APIVersion  string                    `yaml:"apiVersion"`
	Name        string                    `yaml:"name"`
	Version     string                    `yaml:"version"`
	Description string                    `yaml:"description"`
	Author      string                    `yaml:"author,omitempty"`
	License     string                    `yaml:"license,omitempty"`
	Repository  string                    `yaml:"repository,omitempty"`
	Homepage    string                    `yaml:"homepage,omitempty"`
	BaseURL     string                    `yaml:"baseUrl"`
	Auth        Auth                      `yaml:"auth,omitempty"`
	Tags        []string                  `yaml:"tags,omitempty"`
	Category    string                    `yaml:"category,omitempty"`
	Types       map[string]TypeDefinition `yaml:"types,omitempty"`
	Actions     map[string]Action         `yaml:"actions"`
}

func Load(path string) (*Definition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read emblem file: %w", err)
	}

	return Parse(data)
}

func Parse(data []byte) (*Definition, error) {
	var def Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse emblem YAML: %w", err)
	}

	if err := Validate(&def); err != nil {
		return nil, fmt.Errorf("invalid emblem: %w", err)
	}

	return &def, nil
}

func Validate(def *Definition) error {
	if def.APIVersion != "v1" {
		return fmt.Errorf("unsupported API version: %s", def.APIVersion)
	}

	if def.Name == "" {
		return fmt.Errorf("emblem name is required")
	}

	if def.Version == "" {
		return fmt.Errorf("emblem version is required")
	}

	if def.BaseURL == "" {
		return fmt.Errorf("baseUrl is required")
	}

	if len(def.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}

	for actionName, action := range def.Actions {
		if action.Method == "" {
			return fmt.Errorf("action %s: method is required", actionName)
		}
		if action.Path == "" {
			return fmt.Errorf("action %s: path is required", actionName)
		}
		if action.Description == "" {
			return fmt.Errorf("action %s: description is required", actionName)
		}
	}

	return nil
}

func (d *Definition) GetAction(name string) (*Action, error) {
	action, ok := d.Actions[name]
	if !ok {
		return nil, fmt.Errorf("action %s not found", name)
	}
	return &action, nil
}

func (d *Definition) ListActions() []string {
	actions := make([]string, 0, len(d.Actions))
	for name := range d.Actions {
		actions = append(actions, name)
	}
	return actions
}

func (d *Definition) GetAuthCredentials() (map[string]string, error) {
	creds := make(map[string]string)

	if d.Auth.Type == AuthNone {
		return creds, nil
	}

	if d.Auth.KeyEnv == "" {
		return nil, fmt.Errorf("auth type %s requires keyEnv to be set", d.Auth.Type)
	}

	value := os.Getenv(d.Auth.KeyEnv)
	if value == "" {
		return nil, fmt.Errorf("environment variable %s is not set", d.Auth.KeyEnv)
	}

	creds["value"] = value

	switch d.Auth.Type {
	case AuthAPIKey:
		if d.Auth.Header != "" {
			creds["header"] = d.Auth.Header
		} else {
			creds["header"] = "X-API-Key"
		}
	case AuthBearer:
		creds["header"] = "Authorization"
		if d.Auth.Prefix != "" {
			creds["prefix"] = d.Auth.Prefix
		} else {
			creds["prefix"] = "Bearer "
		}
	case AuthBasic:
		creds["header"] = "Authorization"
		creds["prefix"] = "Basic "
	}

	return creds, nil
}

func GetCachePath(name, version string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".elysium", "cache", fmt.Sprintf("%s@%s", name, version), "emblem.yaml"), nil
}

func SaveToCache(name, version string, data []byte) error {
	cacheFile, err := GetCachePath(name, version)
	if err != nil {
		return err
	}

	cacheDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write emblem to cache: %w", err)
	}

	return nil
}

func LoadFromCache(name, version string) (*Definition, error) {
	cacheFile, err := GetCachePath(name, version)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("emblem %s@%s not found in cache", name, version)
	}

	return Load(cacheFile)
}

func ParseVersionConstraint(constraint string) (name, version string, err error) {
	parts := strings.Split(constraint, "@")
	if len(parts) == 1 {
		return parts[0], "latest", nil
	}
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return "", "", fmt.Errorf("invalid version constraint: %s", constraint)
}
