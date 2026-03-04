package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
)

// newCheckUpdatesServer creates a test server returning emblem info for the
// given name→latest mapping, with an optional security advisory.
func newCheckUpdatesServer(t *testing.T, emblems []api.Emblem) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		for _, e := range emblems {
			if r.URL.Path == "/api/emblems/"+e.Name {
				json.NewEncoder(w).Encode(e)
				return
			}
		}
		http.NotFound(w, r)
	}))
}

func TestCheckUpdatesCmd_NoEmblems(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	err := checkUpdatesCmd.RunE(checkUpdatesCmd, []string{})
	if err != nil {
		t.Errorf("checkUpdatesCmd.RunE() unexpected error: %v", err)
	}
}

func TestCheckUpdatesCmd_AllUpToDate(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("stable-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := newCheckUpdatesServer(t, []api.Emblem{
		{ID: "e1", Name: "stable-api", LatestVersion: "1.0.0"},
	})
	defer server.Close()
	config.Get().Registry = server.URL

	err := checkUpdatesCmd.RunE(checkUpdatesCmd, []string{})
	if err != nil {
		t.Errorf("checkUpdatesCmd.RunE() unexpected error: %v", err)
	}
}

func TestCheckUpdatesCmd_OutdatedEmblem(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("outdated-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := newCheckUpdatesServer(t, []api.Emblem{
		{ID: "e1", Name: "outdated-api", LatestVersion: "2.0.0"},
	})
	defer server.Close()
	config.Get().Registry = server.URL

	err := checkUpdatesCmd.RunE(checkUpdatesCmd, []string{})
	if err != nil {
		t.Errorf("checkUpdatesCmd.RunE() unexpected error for outdated: %v", err)
	}
}

func TestCheckUpdatesCmd_SecurityAdvisory(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("vuln-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := newCheckUpdatesServer(t, []api.Emblem{
		{
			ID:               "e1",
			Name:             "vuln-api",
			LatestVersion:    "1.0.1",
			SecurityAdvisory: "CVE-2026-9999",
			SecuritySeverity: "critical",
		},
	})
	defer server.Close()
	config.Get().Registry = server.URL

	err := checkUpdatesCmd.RunE(checkUpdatesCmd, []string{})
	if err != nil {
		t.Errorf("checkUpdatesCmd.RunE() unexpected error for security advisory: %v", err)
	}
}

func TestCheckUpdatesCmd_AdvisoryWithoutSeverity(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("advisory-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := newCheckUpdatesServer(t, []api.Emblem{
		{
			ID:               "e1",
			Name:             "advisory-api",
			LatestVersion:    "1.0.1",
			SecurityAdvisory: "GHSA-xxxx-yyyy-zzzz",
			SecuritySeverity: "", // no severity
		},
	})
	defer server.Close()
	config.Get().Registry = server.URL

	err := checkUpdatesCmd.RunE(checkUpdatesCmd, []string{})
	if err != nil {
		t.Errorf("checkUpdatesCmd.RunE() unexpected error for advisory without severity: %v", err)
	}
}

func TestCheckUpdatesCmd_APIError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("error-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"server error"}`, http.StatusInternalServerError)
	}))
	defer server.Close()
	config.Get().Registry = server.URL

	// API errors are printed as warnings, not returned.
	err := checkUpdatesCmd.RunE(checkUpdatesCmd, []string{})
	if err != nil {
		t.Errorf("checkUpdatesCmd.RunE() unexpected error for API 500: %v", err)
	}
}

func TestCheckUpdatesCmd_ConnectionError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("conn-api2", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()

	config.Get().Registry = closedURL

	err := checkUpdatesCmd.RunE(checkUpdatesCmd, []string{})
	if err == nil {
		t.Error("checkUpdatesCmd.RunE() expected connection error, got nil")
	}
	if !strings.Contains(err.Error(), "refused") && !strings.Contains(err.Error(), "connection") {
		t.Errorf("checkUpdatesCmd.RunE() error = %q, want connection-related error", err.Error())
	}
}
