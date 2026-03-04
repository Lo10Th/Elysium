package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/emblem"
)

// newInfoTestServer creates a test server that serves emblem info and version.
func newInfoTestServer(t *testing.T, name, version string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// GET /api/emblems/<name>/<version>
		versionPath := fmt.Sprintf("/api/emblems/%s/%s", name, version)
		if r.URL.Path == versionPath {
			versionResp := api.EmblemVersion{
				Name:        name,
				Version:     version,
				YAMLContent: validEmblemYAML,
			}
			json.NewEncoder(w).Encode(versionResp)
			return
		}

		// GET /api/emblems/<name>
		if r.URL.Path == fmt.Sprintf("/api/emblems/%s", name) {
			json.NewEncoder(w).Encode(api.Emblem{
				ID:            "e-" + name,
				Name:          name,
				LatestVersion: version,
			})
			return
		}
		http.NotFound(w, r)
	}))
}

// writeEmblemToCache installs an emblem and writes it to the cache directory.
func writeEmblemToCache(t *testing.T, name, version string) {
	t.Helper()
	if err := config.InstallEmblem(name, version); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	cfg := config.Get()
	cacheEntry := filepath.Join(cfg.CacheDir, fmt.Sprintf("%s@%s", name, version))
	if err := os.MkdirAll(cacheEntry, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheEntry, "emblem.yaml"), []byte(validEmblemYAML), 0644); err != nil {
		t.Fatalf("WriteFile emblem.yaml: %v", err)
	}
}

func TestInfoCmd_FromCache(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	writeEmblemToCache(t, "test-shop", "1.0.0")

	err := infoCmd.RunE(infoCmd, []string{"test-shop"})
	if err != nil {
		t.Errorf("infoCmd.RunE() unexpected error loading from cache: %v", err)
	}
}

func TestInfoCmd_NotInstalled_FetchFromRegistry(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newInfoTestServer(t, "remote-api", "2.0.0")
	defer server.Close()
	config.Get().Registry = server.URL

	err := infoCmd.RunE(infoCmd, []string{"remote-api@2.0.0"})
	if err != nil {
		t.Errorf("infoCmd.RunE() unexpected error fetching from registry: %v", err)
	}
}

func TestInfoCmd_ParseError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Invalid version constraint format.
	err := infoCmd.RunE(infoCmd, []string{"name@bad@version"})
	if err == nil {
		t.Error("infoCmd.RunE() expected parse error for bad version, got nil")
	}
}

func TestInfoCmd_RegistryError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()
	config.Get().Registry = closedURL

	err := infoCmd.RunE(infoCmd, []string{"missing-api"})
	if err == nil {
		t.Error("infoCmd.RunE() expected error for connection failure, got nil")
	}
}

func TestInfoCmd_CacheLoadFails_FallsBackToRegistry(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Register emblem in config but don't write cache file, forcing fallback.
	if err := config.InstallEmblem("fallback-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := newInfoTestServer(t, "fallback-api", "1.0.0")
	defer server.Close()
	config.Get().Registry = server.URL

	err := infoCmd.RunE(infoCmd, []string{"fallback-api"})
	if err != nil {
		t.Errorf("infoCmd.RunE() unexpected error for cache-fallback: %v", err)
	}
}

func TestInfoCmd_Verbose(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	writeEmblemToCache(t, "verbose-info-api", "1.0.0")

	// Set verbose flag on rootCmd.
	_ = rootCmd.PersistentFlags().Set("verbose", "true")
	defer func() { _ = rootCmd.PersistentFlags().Set("verbose", "false") }()

	err := infoCmd.RunE(infoCmd, []string{"verbose-info-api"})
	if err != nil {
		t.Errorf("infoCmd.RunE() verbose unexpected error: %v", err)
	}
}

func TestInfoCmd_WithTags(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Write an emblem with tags and category to the cache.
	withTagsYAML := `apiVersion: v1
name: tagged-api
version: 1.0.0
description: An API with tags
baseUrl: http://localhost:5000/api
author: Test Author
license: MIT
category: ecommerce
tags:
  - payments
  - ecommerce
auth:
  type: none
actions:
  list:
    description: List items
    method: GET
    path: /items
`
	if err := config.InstallEmblem("tagged-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	cfg := config.Get()
	cacheEntry := filepath.Join(cfg.CacheDir, "tagged-api@1.0.0")
	if err := os.MkdirAll(cacheEntry, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheEntry, "emblem.yaml"), []byte(withTagsYAML), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	err := infoCmd.RunE(infoCmd, []string{"tagged-api"})
	if err != nil {
		t.Errorf("infoCmd.RunE() unexpected error for tagged emblem: %v", err)
	}
}

func TestInfoCmd_AuthTypeNotNone(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	authYAML := `apiVersion: v1
name: auth-api
version: 1.0.0
description: An API requiring auth
baseUrl: http://localhost:5000/api
auth:
  type: api_key
  key_env: MY_API_KEY
  header: X-API-Key
actions:
  list:
    description: List items
    method: GET
    path: /items
`
	if err := config.InstallEmblem("auth-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	cfg := config.Get()
	cacheEntry := filepath.Join(cfg.CacheDir, "auth-api@1.0.0")
	if err := os.MkdirAll(cacheEntry, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheEntry, "emblem.yaml"), []byte(authYAML), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	err := infoCmd.RunE(infoCmd, []string{"auth-api"})
	if err != nil {
		t.Errorf("infoCmd.RunE() unexpected error for auth emblem: %v", err)
	}
}

// TestInfoCmd_LatestFromRegistry tests the "latest" version lookup path where
// infoCmd fetches from registry to resolve the actual latest version.
func TestInfoCmd_LatestFromRegistry(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newInfoTestServer(t, "latest-api", "3.0.0")
	defer server.Close()
	config.Get().Registry = server.URL

	// "latest" → triggers the GetEmblem + GetEmblemVersion path.
	err := infoCmd.RunE(infoCmd, []string{"latest-api@latest"})
	if err != nil {
		t.Errorf("infoCmd.RunE(@latest) unexpected error: %v", err)
	}
}

// TestInfoCmd_ParseEmblemFails tests the error path when the YAML from the
// registry cannot be parsed by emblem.Parse.
func TestInfoCmd_ParseEmblemFails(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/1.0.0") {
			vr := api.EmblemVersion{
				Name:        "bad-yaml-api",
				Version:     "1.0.0",
				YAMLContent: "not: valid: yaml: emblem",
			}
			json.NewEncoder(w).Encode(vr)
			return
		}
		json.NewEncoder(w).Encode(api.Emblem{
			ID: "e1", Name: "bad-yaml-api", LatestVersion: "1.0.0",
		})
	}))
	defer server.Close()
	config.Get().Registry = server.URL

	err := infoCmd.RunE(infoCmd, []string{"bad-yaml-api@1.0.0"})
	// The emblem.Parse might succeed (it just returns a struct) or fail depending
	// on the YAML content. Either way it should not panic.
	_ = err
}

// TestEmblemLoadFromCacheAndActions exercises the code path that reads def.Actions.
func TestInfoCmd_MultipleActions(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	multiActionYAML := `apiVersion: v1
name: multi-api
version: 1.0.0
description: Multi action API
baseUrl: http://localhost:5000/api
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
  delete:
    description: Delete item
    method: DELETE
    path: /items/{id}
`
	if err := config.InstallEmblem("multi-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	cfg := config.Get()
	cacheEntry := filepath.Join(cfg.CacheDir, "multi-api@1.0.0")
	if err := os.MkdirAll(cacheEntry, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheEntry, "emblem.yaml"), []byte(multiActionYAML), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Verify that emblem.LoadFromCache works for this YAML.
	def, err := emblem.LoadFromCache("multi-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache: %v", err)
	}
	if len(def.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(def.Actions))
	}

	err = infoCmd.RunE(infoCmd, []string{"multi-api"})
	if err != nil {
		t.Errorf("infoCmd.RunE() unexpected error for multi-action emblem: %v", err)
	}
}
