package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
)

func TestSearchCmd_NoResults(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]api.Emblem{})
	}))
	defer server.Close()

	config.Get().Registry = server.URL

	err := searchCmd.RunE(searchCmd, []string{"payment"})
	if err != nil {
		t.Errorf("searchCmd.RunE() unexpected error: %v", err)
	}
}

func TestSearchCmd_WithResults(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	emblems := []api.Emblem{
		{
			ID:            "e1",
			Name:          "stripe",
			Description:   "Stripe payment integration",
			LatestVersion: "1.2.0",
		},
		{
			ID:            "e2",
			Name:          "paypal",
			Description:   "PayPal payment API",
			LatestVersion: "2.0.0",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emblems)
	}))
	defer server.Close()

	config.Get().Registry = server.URL

	err := searchCmd.RunE(searchCmd, []string{"payment"})
	if err != nil {
		t.Errorf("searchCmd.RunE() unexpected error: %v", err)
	}
}

func TestSearchCmd_LongDescription(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	emblems := []api.Emblem{
		{
			ID:            "e1",
			Name:          "long-desc",
			Description:   "This description is very long and should be truncated at 42 characters in the output table when rendered to the terminal by the search command",
			LatestVersion: "1.0.0",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emblems)
	}))
	defer server.Close()

	config.Get().Registry = server.URL

	err := searchCmd.RunE(searchCmd, []string{"long"})
	if err != nil {
		t.Errorf("searchCmd.RunE() unexpected error for long description: %v", err)
	}
}

func TestSearchCmd_ConnectionError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()

	config.Get().Registry = closedURL

	err := searchCmd.RunE(searchCmd, []string{"anything"})
	if err == nil {
		t.Error("searchCmd.RunE() expected error for connection refused, got nil")
	}
}
