package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
)

// newOutdatedTestServer creates a test server that handles emblem lookup.
// latestVersions maps emblem name → latest version string.
func newOutdatedTestServer(t *testing.T, latestVersions map[string]string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Extract emblem name from path /api/emblems/<name>
		for name, latest := range latestVersions {
			if r.URL.Path == "/api/emblems/"+name {
				if statusCode != http.StatusOK {
					http.Error(w, `{"detail":"error"}`, statusCode)
					return
				}
				emblem := api.Emblem{
					ID:            "e-" + name,
					Name:          name,
					LatestVersion: latest,
				}
				json.NewEncoder(w).Encode(emblem)
				return
			}
		}
		http.NotFound(w, r)
	}))
}

func TestOutdatedCmd_NoEmblems(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	err := outdatedCmd.RunE(outdatedCmd, []string{})
	if err != nil {
		t.Errorf("outdatedCmd.RunE() unexpected error: %v", err)
	}
}

func TestOutdatedCmd_AllUpToDate(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("my-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := newOutdatedTestServer(t, map[string]string{"my-api": "1.0.0"}, http.StatusOK)
	defer server.Close()

	config.Get().Registry = server.URL

	err := outdatedCmd.RunE(outdatedCmd, []string{})
	if err != nil {
		t.Errorf("outdatedCmd.RunE() unexpected error: %v", err)
	}
}

func TestOutdatedCmd_SomeOutdated(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("my-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := newOutdatedTestServer(t, map[string]string{"my-api": "2.0.0"}, http.StatusOK)
	defer server.Close()

	config.Get().Registry = server.URL

	err := outdatedCmd.RunE(outdatedCmd, []string{})
	if err != nil {
		t.Errorf("outdatedCmd.RunE() unexpected error for outdated emblems: %v", err)
	}
}

func TestOutdatedCmd_APIError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("error-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := newOutdatedTestServer(t, map[string]string{"error-api": "1.0.0"}, http.StatusInternalServerError)
	defer server.Close()

	config.Get().Registry = server.URL

	// Should return nil (errors are printed as warnings, not returned).
	err := outdatedCmd.RunE(outdatedCmd, []string{})
	if err != nil {
		t.Errorf("outdatedCmd.RunE() unexpected error for API error: %v", err)
	}
}

func TestOutdatedCmd_ConnectionError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("conn-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()

	config.Get().Registry = closedURL

	err := outdatedCmd.RunE(outdatedCmd, []string{})
	// connection refused → error is returned immediately
	if err == nil {
		t.Error("outdatedCmd.RunE() expected connection error, got nil")
	}
}
