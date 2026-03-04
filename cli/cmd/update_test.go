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
)

// newUpdateTestServer creates a test server that handles:
//   GET /api/emblems/<name>          → latest version
//   GET /api/emblems/<name>/<ver>    → EmblemVersion with yaml_content
func newUpdateTestServer(t *testing.T, name, latestVersion string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		versionPrefix := fmt.Sprintf("/api/emblems/%s/", name)
		if strings.HasPrefix(r.URL.Path, versionPrefix) {
			ver := strings.TrimPrefix(r.URL.Path, versionPrefix)
			versionResp := api.EmblemVersion{
				Name:        name,
				Version:     ver,
				YAMLContent: validEmblemYAML,
			}
			json.NewEncoder(w).Encode(versionResp)
			return
		}

		if r.URL.Path == fmt.Sprintf("/api/emblems/%s", name) {
			json.NewEncoder(w).Encode(api.Emblem{
				ID:            "e-" + name,
				Name:          name,
				LatestVersion: latestVersion,
			})
			return
		}
		http.NotFound(w, r)
	}))
}

func TestUpdateCmd_NoArgsNoAll(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// No --all flag, no args → error.
	err := updateCmd.RunE(updateCmd, []string{})
	if err == nil {
		t.Error("updateCmd.RunE() expected error with no args and no --all, got nil")
	}
}

func TestUpdateCmd_AllNoEmblems(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	_ = updateCmd.Flags().Set("all", "true")
	defer func() { _ = updateCmd.Flags().Set("all", "false") }()

	err := updateCmd.RunE(updateCmd, []string{})
	if err != nil {
		t.Errorf("updateCmd.RunE(--all, no emblems) unexpected error: %v", err)
	}
}

func TestUpdateCmd_SpecificEmblem_NotInstalled(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Passing an emblem name that is not installed should report failure.
	err := updateCmd.RunE(updateCmd, []string{"ghost-api"})
	if err == nil {
		t.Error("updateCmd.RunE() expected error for uninstalled emblem, got nil")
	}
}

func TestUpdateCmd_SpecificEmblem_AlreadyLatest(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("current-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := newUpdateTestServer(t, "current-api", "1.0.0")
	defer server.Close()
	config.Get().Registry = server.URL

	err := updateCmd.RunE(updateCmd, []string{"current-api"})
	if err != nil {
		t.Errorf("updateCmd.RunE() unexpected error for up-to-date emblem: %v", err)
	}
}

func TestUpdateCmd_SpecificEmblem_UpdateAvailable(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("upgrade-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	// Write an emblem into cache so the pull (inside update) can save a new version.
	cfg := config.Get()
	cacheOld := filepath.Join(cfg.CacheDir, "upgrade-api@1.0.0")
	if err := os.MkdirAll(cacheOld, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	server := newUpdateTestServer(t, "upgrade-api", "2.0.0")
	defer server.Close()
	config.Get().Registry = server.URL

	err := updateCmd.RunE(updateCmd, []string{"upgrade-api"})
	if err != nil {
		t.Errorf("updateCmd.RunE() unexpected error for available update: %v", err)
	}
}

func TestUpdateCmd_All_UpdatesAvailable(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("bulk-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := newUpdateTestServer(t, "bulk-api", "1.1.0")
	defer server.Close()
	config.Get().Registry = server.URL

	_ = updateCmd.Flags().Set("all", "true")
	defer func() { _ = updateCmd.Flags().Set("all", "false") }()

	err := updateCmd.RunE(updateCmd, []string{})
	if err != nil {
		t.Errorf("updateCmd.RunE(--all) unexpected error: %v", err)
	}
}

func TestUpdateCmd_EmblemNotFound(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("notfound-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
	}))
	defer server.Close()
	config.Get().Registry = server.URL

	err := updateCmd.RunE(updateCmd, []string{"notfound-api"})
	if err == nil {
		t.Error("updateCmd.RunE() expected error for 404, got nil")
	}
}
