package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elysium/elysium/cli/internal/api"
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
