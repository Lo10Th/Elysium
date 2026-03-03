package executor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elysium/elysium/cli/internal/emblem"
)

func makeTestDef(baseURL string) *emblem.Definition {
	return &emblem.Definition{
		APIVersion:  "v1",
		Name:        "test",
		Version:     "1.0.0",
		Description: "test emblem",
		BaseURL:     baseURL,
		Auth:        emblem.Auth{Type: emblem.AuthNone},
		Actions: map[string]emblem.Action{
			"get-item": {
				Description: "Get an item",
				Method:      "GET",
				Path:        "/items/{id}",
			},
		},
	}
}

func TestBuildURL_PathEncoding(t *testing.T) {
	exec := New(makeTestDef("http://example.com"))

	tests := []struct {
		name     string
		path     string
		params   map[string]interface{}
		contains string
		notContains string
	}{
		{
			name:     "normal param",
			path:     "/items/{id}",
			params:   map[string]interface{}{"id": "123"},
			contains: "/items/123",
		},
		{
			name:        "path traversal attempt",
			path:        "/items/{id}",
			params:      map[string]interface{}{"id": "../secret"},
			notContains: "../secret",
			contains:    "%2F",
		},
		{
			name:     "param with spaces",
			path:     "/items/{id}",
			params:   map[string]interface{}{"id": "hello world"},
			contains: "hello%20world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.buildURL(tt.path, tt.params)
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("buildURL(%q, %v) = %q, expected to contain %q", tt.path, tt.params, result, tt.contains)
			}
			if tt.notContains != "" && strings.Contains(result, tt.notContains) {
				t.Errorf("buildURL(%q, %v) = %q, expected NOT to contain %q", tt.path, tt.params, result, tt.notContains)
			}
		})
	}
}

func TestExecute_URLSchemeValidation(t *testing.T) {
	def := &emblem.Definition{
		APIVersion:  "v1",
		Name:        "test",
		Version:     "1.0.0",
		Description: "test emblem",
		BaseURL:     "file:///etc",
		Auth:        emblem.Auth{Type: emblem.AuthNone},
		Actions: map[string]emblem.Action{
			"read": {
				Description: "read",
				Method:      "GET",
				Path:        "/passwd",
			},
		},
	}

	exec := New(def)
	_, err := exec.Execute("read", nil, FormatOptions{})
	if err == nil {
		t.Fatal("Execute() with file:// URL should return an error")
	}
	if !strings.Contains(err.Error(), "invalid URL scheme") {
		t.Errorf("Execute() error = %q, expected 'invalid URL scheme'", err.Error())
	}
}

func TestExecute_ResponseSizeLimit(t *testing.T) {
	// Create a server that returns a response larger than the limit
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write maxResponseBytes + 1 bytes
		chunk := strings.Repeat("x", 1024)
		for i := 0; i <= maxResponseBytes/len(chunk); i++ {
			w.Write([]byte(chunk))
		}
	}))
	defer server.Close()

	def := makeTestDef(server.URL)
	exec := New(def)
	_, err := exec.Execute("get-item", map[string]interface{}{"id": "1"}, FormatOptions{})
	if err == nil {
		t.Fatal("Execute() with oversized response should return an error")
	}
	if !strings.Contains(err.Error(), "response too large") {
		t.Errorf("Execute() error = %q, expected 'response too large'", err.Error())
	}
}
