package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/elysium/elysium/cli/internal/config"
)

// setupEmblemInCache installs an emblem into the temp config and writes a
// valid emblem.yaml file to the cache directory.
func setupEmblemInCache(t *testing.T, name, version, baseURL string) {
	t.Helper()

	// Register in config.
	if err := config.InstallEmblem(name, version); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	// Write emblem.yaml to cache.
	cfg := config.Get()
	cacheEntry := filepath.Join(cfg.CacheDir, fmt.Sprintf("%s@%s", name, version))
	if err := os.MkdirAll(cacheEntry, 0755); err != nil {
		t.Fatalf("MkdirAll cache: %v", err)
	}

	yamlContent := fmt.Sprintf(`apiVersion: v1
name: %s
version: %s
description: Test emblem for execute tests
baseUrl: %s
auth:
  type: none
actions:
  list:
    description: List items
    method: GET
    path: /items
  create:
    description: Create item
    method: POST
    path: /items
`, name, version, baseURL)

	if err := os.WriteFile(filepath.Join(cacheEntry, "emblem.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("WriteFile emblem.yaml: %v", err)
	}
}

func TestExecuteEmblemAction_EmblemNotInstalled(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Don't install anything – the config file exists but has no "my-emblem" entry.
	err := executeEmblemAction("my-emblem", []string{})
	if err == nil {
		t.Error("executeEmblemAction() expected error for uninstalled emblem, got nil")
	}
}

func TestExecuteEmblemAction_NoConfigFile(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Remove the config file so reading it fails.
	cfg := config.Get()
	configPath := cfg.CacheDir[:len(cfg.CacheDir)-len("/cache")] + "/config.yaml"
	os.Remove(configPath)

	err := executeEmblemAction("any-emblem", []string{})
	if err == nil {
		t.Error("executeEmblemAction() expected error when config missing, got nil")
	}
}

func TestExecuteEmblemAction_ListActions_NoAction(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	setupEmblemInCache(t, "shop", "1.0.0", "http://localhost:9999/api")

	// No action name → should list available actions and return nil.
	err := executeEmblemAction("shop", []string{})
	if err != nil {
		t.Errorf("executeEmblemAction(list actions) unexpected error: %v", err)
	}
}

func TestExecuteEmblemAction_UnknownAction(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	setupEmblemInCache(t, "shop", "1.0.0", "http://localhost:9999/api")

	err := executeEmblemAction("shop", []string{"no-such-action"})
	if err == nil {
		t.Error("executeEmblemAction() expected error for unknown action, got nil")
	}
}

func TestExecuteEmblemAction_ValidAction_WithServer(t *testing.T) {
	// Start a mock API server for the "list" action.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "name": "Widget"},
		})
	}))
	defer server.Close()

	cleanup := initTestConfig(t)
	defer cleanup()

	setupEmblemInCache(t, "shop", "1.0.0", server.URL)

	err := executeEmblemAction("shop", []string{"list"})
	if err != nil {
		t.Errorf("executeEmblemAction(list with server) unexpected error: %v", err)
	}
}

func TestExecuteEmblemAction_NoCacheFile(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Install emblem in config but don't write emblem.yaml to cache.
	if err := config.InstallEmblem("ghost", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	err := executeEmblemAction("ghost", []string{"list"})
	if err == nil {
		t.Error("executeEmblemAction() expected error when cache file missing, got nil")
	}
}

// --- parseParams extra coverage ---

func TestParseParams_ParamsFile_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	paramsPath := filepath.Join(tmpDir, "params.json")

	if err := os.WriteFile(paramsPath, []byte(`{"category":"electronics","limit":5}`), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	oldParamsFile := paramsFile
	paramsFile = paramsPath
	defer func() { paramsFile = oldParamsFile }()

	result, err := parseParams([]string{})
	if err != nil {
		t.Fatalf("parseParams() unexpected error: %v", err)
	}
	if result["category"] != "electronics" {
		t.Errorf("parseParams() category = %v, want 'electronics'", result["category"])
	}
}

func TestParseParams_ParamsFile_NotFound(t *testing.T) {
	oldParamsFile := paramsFile
	paramsFile = "/nonexistent/params.json"
	defer func() { paramsFile = oldParamsFile }()

	_, err := parseParams([]string{})
	if err == nil {
		t.Error("parseParams() expected error for missing params file, got nil")
	}
}

func TestParseParams_ParamsFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	paramsPath := filepath.Join(tmpDir, "params.json")

	if err := os.WriteFile(paramsPath, []byte(`not valid json`), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	oldParamsFile := paramsFile
	paramsFile = paramsPath
	defer func() { paramsFile = oldParamsFile }()

	_, err := parseParams([]string{})
	if err == nil {
		t.Error("parseParams() expected error for invalid JSON file, got nil")
	}
}

func TestParseParams_InvalidParamsJSON(t *testing.T) {
	oldParamsJSON := paramsJSON
	paramsJSON = `{bad json`
	defer func() { paramsJSON = oldParamsJSON }()

	_, err := parseParams([]string{})
	if err == nil {
		t.Error("parseParams() expected error for invalid --params JSON, got nil")
	}
}

func TestParseParams_FlagWithJSONValue(t *testing.T) {
	result, err := parseParams([]string{"--filter", `{"key":"val"}`})
	if err != nil {
		t.Fatalf("parseParams() unexpected error: %v", err)
	}
	if result["filter"] == nil {
		t.Error("parseParams() filter key missing")
	}
}

func TestParseParams_ShortFlagNoValue(t *testing.T) {
	result, err := parseParams([]string{"-v"})
	if err != nil {
		t.Fatalf("parseParams() unexpected error: %v", err)
	}
	if result["v"] != "true" {
		t.Errorf("parseParams() -v = %v, want 'true'", result["v"])
	}
}

func TestParseParams_FlagAtEnd(t *testing.T) {
	result, err := parseParams([]string{"--verbose", "--debug"})
	if err != nil {
		t.Fatalf("parseParams() unexpected error: %v", err)
	}
	if result["verbose"] != "true" {
		t.Errorf("parseParams() --verbose = %v, want 'true'", result["verbose"])
	}
}
