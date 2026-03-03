package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
)

func TestSecurityIcon(t *testing.T) {
	tests := []struct {
		severity string
		want     string
	}{
		{"critical", "🔴 CRITICAL"},
		{"CRITICAL", "🔴 CRITICAL"},
		{"high", "🔴 HIGH"},
		{"medium", "⚠️  MEDIUM"},
		{"low", "⚠️  LOW"},
		{"", "✓"},
		{"none", "✓"},
		{"unknown", "✓"},
	}

	for _, tt := range tests {
		got := securityIcon(tt.severity)
		if got != tt.want {
			t.Errorf("securityIcon(%q) = %q, want %q", tt.severity, got, tt.want)
		}
	}
}

func TestCheckUpdatesQueryEmblem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		emblem := api.Emblem{
			ID:               "emblem-1",
			Name:             "clothing-shop",
			LatestVersion:    "1.1.0",
			SecurityAdvisory: "",
			SecuritySeverity: "",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emblem)
	}))
	defer server.Close()

	client := api.NewClientWithBaseURL(server.URL)

	emblem, err := client.GetEmblem("clothing-shop")
	if err != nil {
		t.Fatalf("GetEmblem() error = %v", err)
	}

	if emblem.LatestVersion != "1.1.0" {
		t.Errorf("LatestVersion = %q, want %q", emblem.LatestVersion, "1.1.0")
	}

	if emblem.SecurityAdvisory != "" {
		t.Errorf("SecurityAdvisory = %q, want empty", emblem.SecurityAdvisory)
	}
}

func TestCheckUpdatesQueryEmblemWithSecurity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		emblem := api.Emblem{
			ID:               "emblem-2",
			Name:             "stripe",
			LatestVersion:    "2.3.2",
			SecurityAdvisory: "CVE-2026-1234",
			SecuritySeverity: "critical",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emblem)
	}))
	defer server.Close()

	client := api.NewClientWithBaseURL(server.URL)

	emblem, err := client.GetEmblem("stripe")
	if err != nil {
		t.Fatalf("GetEmblem() error = %v", err)
	}

	if emblem.SecurityAdvisory != "CVE-2026-1234" {
		t.Errorf("SecurityAdvisory = %q, want %q", emblem.SecurityAdvisory, "CVE-2026-1234")
	}

	if emblem.SecuritySeverity != "critical" {
		t.Errorf("SecuritySeverity = %q, want %q", emblem.SecuritySeverity, "critical")
	}

	icon := securityIcon(emblem.SecuritySeverity)
	if icon != "🔴 CRITICAL" {
		t.Errorf("securityIcon(%q) = %q, want %q", emblem.SecuritySeverity, icon, "🔴 CRITICAL")
	}
}

// initTestConfig sets up a minimal config in a temp home directory and returns a cleanup func.
func initTestConfig(t *testing.T) func() {
	t.Helper()
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	os.MkdirAll(filepath.Join(tmpHome, ".elysium", "cache"), 0755)
	if err := config.Init(); err != nil {
		t.Fatalf("config.Init() error = %v", err)
	}
	return func() { os.Setenv("HOME", oldHome) }
}

func TestPrintUpdateNotification_NoCheck(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// With noCheck=true, nothing should be printed even if an update is available.
	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	config.InstallEmblem("clothing-shop", "1.0.0")
	config.SetVersionCache("clothing-shop", "1.1.0", "", "")

	PrintUpdateNotification("clothing-shop", true)

	w.Close()
	os.Stdout = old
	buf.Reset()
	buf.ReadFrom(r)

	if buf.Len() != 0 {
		t.Errorf("expected no output with noCheck=true, got %q", buf.String())
	}
}

func TestPrintUpdateNotification_Outdated(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	config.InstallEmblem("clothing-shop", "1.0.0")
	config.SetVersionCache("clothing-shop", "1.1.0", "", "")

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	PrintUpdateNotification("clothing-shop", false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "1.1.0") {
		t.Errorf("expected output to mention latest version 1.1.0, got %q", out)
	}
	if !strings.Contains(out, "ely update clothing-shop") {
		t.Errorf("expected output to contain 'ely update clothing-shop', got %q", out)
	}
}

func TestPrintUpdateNotification_SecurityAdvisory(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	config.InstallEmblem("stripe", "2.3.1")
	config.SetVersionCache("stripe", "2.3.2", "CVE-2026-1234", "critical")

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	PrintUpdateNotification("stripe", false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "CVE-2026-1234") {
		t.Errorf("expected output to mention CVE-2026-1234, got %q", out)
	}
	if !strings.Contains(out, "SECURITY UPDATE") {
		t.Errorf("expected output to contain 'SECURITY UPDATE', got %q", out)
	}
}

func TestPrintUpdateNotification_UpToDate(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	config.InstallEmblem("github", "3.0.1")
	config.SetVersionCache("github", "3.0.1", "", "")

	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	PrintUpdateNotification("github", false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if out != "" {
		t.Errorf("expected no output when up-to-date, got %q", out)
	}
}
