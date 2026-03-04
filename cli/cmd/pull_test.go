package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
)

// validEmblemYAML is a minimal valid emblem YAML for use as yaml_content in
// EmblemVersion mock responses.
const validEmblemYAML = `apiVersion: v1
name: test-shop
version: 1.0.0
description: Test emblem for unit tests
baseUrl: http://localhost:5000/api
auth:
  type: none
actions:
  list:
    description: List items
    method: GET
    path: /items
`

// newPullTestServer creates an httptest.Server that serves a single emblem and
// its version.  The server handles:
//
//	GET /api/emblems/<name>          → api.Emblem JSON
//	GET /api/emblems/<name>/<ver>    → api.EmblemVersion JSON
//
// status controls the HTTP status code for the emblem endpoint.
func newPullTestServer(t *testing.T, name, latestVer string, emblemStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// /api/emblems/<name>/<version>
		versionPrefix := fmt.Sprintf("/api/emblems/%s/", name)
		if strings.HasPrefix(path, versionPrefix) {
			ver := strings.TrimPrefix(path, versionPrefix)
			versionResp := api.EmblemVersion{
				Name:        name,
				Version:     ver,
				YAMLContent: validEmblemYAML,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(versionResp)
			return
		}

		// /api/emblems/<name>
		if path == fmt.Sprintf("/api/emblems/%s", name) {
			if emblemStatus != http.StatusOK {
				http.Error(w, `{"detail":"not found"}`, emblemStatus)
				return
			}
			emblemResp := api.Emblem{
				ID:            "emblem-1",
				Name:          name,
				LatestVersion: latestVer,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(emblemResp)
			return
		}

		http.NotFound(w, r)
	}))
}

func TestPullSingleEmblem_ParseError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// A constraint with two @ signs is invalid.
	err := pullSingleEmblem("name@bad@version", false)
	if err == nil {
		t.Error("pullSingleEmblem() expected parse error, got nil")
	}
}

func TestPullSingleEmblem_NotFound(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newPullTestServer(t, "missing-emblem", "1.0.0", http.StatusNotFound)
	defer server.Close()

	config.Get().Registry = server.URL

	err := pullSingleEmblem("missing-emblem", false)
	if err == nil {
		t.Error("pullSingleEmblem() expected error for 404, got nil")
	}
}

func TestPullSingleEmblem_ConnectionError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Start a server and immediately close it so the port is unreachable.
	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()

	config.Get().Registry = closedURL

	err := pullSingleEmblem("any-emblem", false)
	if err == nil {
		t.Error("pullSingleEmblem() expected connection error, got nil")
	}
}

func TestPullSingleEmblem_Success(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newPullTestServer(t, "test-shop", "1.0.0", http.StatusOK)
	defer server.Close()

	config.Get().Registry = server.URL

	err := pullSingleEmblem("test-shop", false)
	if err != nil {
		t.Errorf("pullSingleEmblem() unexpected error: %v", err)
	}
}

func TestPullSingleEmblem_Verbose(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newPullTestServer(t, "test-shop", "2.0.0", http.StatusOK)
	defer server.Close()

	config.Get().Registry = server.URL

	err := pullSingleEmblem("test-shop", true)
	if err != nil {
		t.Errorf("pullSingleEmblem(verbose) unexpected error: %v", err)
	}
}

func TestPullSingleEmblem_WithSpecificVersion(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newPullTestServer(t, "test-shop", "2.0.0", http.StatusOK)
	defer server.Close()

	config.Get().Registry = server.URL

	// Pull a specific older version - server returns it successfully.
	err := pullSingleEmblem("test-shop@1.0.0", false)
	if err != nil {
		t.Errorf("pullSingleEmblem(@1.0.0) unexpected error: %v", err)
	}
}

func TestPullSingleEmblem_VersionNotFound(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Serve the emblem, but return 404 for any version request.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// /api/emblems/<name>/<version> → 404
		if strings.Contains(path, "/api/emblems/test-shop/") {
			http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
			return
		}
		// /api/emblems/<name> → ok
		if path == "/api/emblems/test-shop" {
			emblemResp := api.Emblem{
				ID:            "emblem-1",
				Name:          "test-shop",
				LatestVersion: "1.0.0",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(emblemResp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	config.Get().Registry = server.URL

	err := pullSingleEmblem("test-shop", false)
	if err == nil {
		t.Error("pullSingleEmblem() expected error when version not found, got nil")
	}
}
