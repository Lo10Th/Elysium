package executor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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

// ─── ListActions ──────────────────────────────────────────────────────────────

func TestListActions(t *testing.T) {
	exec := New(makeTestDef("http://example.com"))
	actions := exec.ListActions()
	if len(actions) != 1 {
		t.Fatalf("ListActions() returned %d actions, want 1", len(actions))
	}
	if actions[0] != "get-item" {
		t.Errorf("ListActions()[0] = %q, want %q", actions[0], "get-item")
	}
}

// ─── Execute error paths ───────────────────────────────────────────────────────

func TestExecute_ActionNotFound(t *testing.T) {
	exec := New(makeTestDef("http://example.com"))
	_, err := exec.Execute("nonexistent", nil, FormatOptions{})
	if err == nil {
		t.Fatal("Execute() with unknown action should return an error")
	}
	if !strings.Contains(err.Error(), "action not found") {
		t.Errorf("Execute() error = %q, want 'action not found'", err.Error())
	}
}

func TestExecute_AuthError(t *testing.T) {
	envKey := "ELYSIUM_TEST_KEY_NOT_SET_XYZ123"
	t.Setenv(envKey, "") // ensure env var is empty; t.Setenv restores it on cleanup

	def := &emblem.Definition{
		APIVersion:  "v1",
		Name:        "test",
		Version:     "1.0.0",
		Description: "test",
		BaseURL:     "http://example.com",
		Auth:        emblem.Auth{Type: emblem.AuthAPIKey, KeyEnv: envKey},
		Actions: map[string]emblem.Action{
			"list": {Description: "List items", Method: "GET", Path: "/items"},
		},
	}
	exec := New(def)
	_, err := exec.Execute("list", nil, FormatOptions{})
	if err == nil {
		t.Fatal("Execute() with missing auth env var should return an error")
	}
	if !strings.Contains(err.Error(), "authentication error") {
		t.Errorf("Execute() error = %q, want 'authentication error'", err.Error())
	}
}

func TestExecute_UnsupportedMethod(t *testing.T) {
	def := &emblem.Definition{
		APIVersion:  "v1",
		Name:        "test",
		Version:     "1.0.0",
		Description: "test",
		BaseURL:     "http://example.com",
		Auth:        emblem.Auth{Type: emblem.AuthNone},
		Actions: map[string]emblem.Action{
			"trace": {Description: "Trace action", Method: "TRACE", Path: "/"},
		},
	}
	exec := New(def)
	_, err := exec.Execute("trace", nil, FormatOptions{})
	if err == nil {
		t.Fatal("Execute() with unsupported HTTP method should return an error")
	}
	if !strings.Contains(err.Error(), "unsupported HTTP method") {
		t.Errorf("Execute() error = %q, want 'unsupported HTTP method'", err.Error())
	}
}

// ─── Execute error responses ───────────────────────────────────────────────────

func TestExecute_ErrorResponse_ErrorField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	exec := New(makeTestDef(server.URL))
	_, err := exec.Execute("get-item", map[string]interface{}{"id": "1"}, FormatOptions{})
	if err == nil {
		t.Fatal("Execute() should return an error for 400 response")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("Execute() error = %q, want to contain 'bad request'", err.Error())
	}
}

func TestExecute_ErrorResponse_MessageField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "invalid input"}`))
	}))
	defer server.Close()

	exec := New(makeTestDef(server.URL))
	_, err := exec.Execute("get-item", map[string]interface{}{"id": "1"}, FormatOptions{})
	if err == nil {
		t.Fatal("Execute() should return an error for 400 response with message field")
	}
	if !strings.Contains(err.Error(), "invalid input") {
		t.Errorf("Execute() error = %q, want to contain 'invalid input'", err.Error())
	}
}

func TestExecute_ErrorResponse_StatusText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("not json at all"))
	}))
	defer server.Close()

	exec := New(makeTestDef(server.URL))
	_, err := exec.Execute("get-item", map[string]interface{}{"id": "1"}, FormatOptions{})
	if err == nil {
		t.Fatal("Execute() should return an error for 500 response")
	}
}

func TestExecute_404Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	exec := New(makeTestDef(server.URL))
	_, err := exec.Execute("get-item", map[string]interface{}{"id": "999"}, FormatOptions{})
	if err == nil {
		t.Fatal("Execute() should return an error for 404 response")
	}
}

// ─── Execute success paths ─────────────────────────────────────────────────────

func TestExecute_Success_GET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": "1", "name": "Widget"}`))
	}))
	defer server.Close()

	exec := New(makeTestDef(server.URL))
	out, err := exec.Execute("get-item", map[string]interface{}{"id": "1"}, FormatOptions{Format: FormatJSON})
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "Widget") {
		t.Errorf("Execute() output = %q, want to contain 'Widget'", out)
	}
}

func TestExecute_HTTPMethods(t *testing.T) {
	for _, method := range []string{"POST", "PUT", "DELETE", "PATCH"} {
		method := method
		t.Run(method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					t.Errorf("expected %s, got %s", method, r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status": "ok"}`))
			}))
			defer server.Close()

			def := &emblem.Definition{
				APIVersion:  "v1",
				Name:        "test",
				Version:     "1.0.0",
				Description: "test",
				BaseURL:     server.URL,
				Auth:        emblem.Auth{Type: emblem.AuthNone},
				Actions: map[string]emblem.Action{
					"action": {Description: "Test action", Method: method, Path: "/resource"},
				},
			}
			exec := New(def)
			out, err := exec.Execute("action", nil, FormatOptions{Format: FormatJSON})
			if err != nil {
				t.Fatalf("Execute(%s) unexpected error: %v", method, err)
			}
			if !strings.Contains(string(out), "ok") {
				t.Errorf("Execute(%s) output = %q, want to contain 'ok'", method, out)
			}
		})
	}
}

func TestExecute_Success_WithQueryParams(t *testing.T) {
	var receivedQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id": 1}]`))
	}))
	defer server.Close()

	def := &emblem.Definition{
		APIVersion:  "v1",
		Name:        "test",
		Version:     "1.0.0",
		Description: "test",
		BaseURL:     server.URL,
		Auth:        emblem.Auth{Type: emblem.AuthNone},
		Actions: map[string]emblem.Action{
			"list": {
				Description: "List items",
				Method:      "GET",
				Path:        "/items",
				Parameters: []emblem.Parameter{
					{Name: "limit", In: "query", Type: "integer"},
					{Name: "page", In: "query", Type: "integer", Default: 1},
				},
			},
		},
	}
	exec := New(def)
	_, err := exec.Execute("list", map[string]interface{}{"limit": "10"}, FormatOptions{Format: FormatJSON})
	if err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if receivedQuery.Get("limit") != "10" {
		t.Errorf("query param limit = %q, want %q", receivedQuery.Get("limit"), "10")
	}
	if receivedQuery.Get("page") != "1" {
		t.Errorf("query param page (default) = %q, want %q", receivedQuery.Get("page"), "1")
	}
}

func TestExecute_Success_POST_WithBodyParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": "new-1"}`))
	}))
	defer server.Close()

	def := &emblem.Definition{
		APIVersion:  "v1",
		Name:        "test",
		Version:     "1.0.0",
		Description: "test",
		BaseURL:     server.URL,
		Auth:        emblem.Auth{Type: emblem.AuthNone},
		Actions: map[string]emblem.Action{
			"create": {
				Description: "Create item",
				Method:      "POST",
				Path:        "/items",
				Parameters: []emblem.Parameter{
					{Name: "name", In: "body", Type: "string"},
					{Name: "price", In: "body", Type: "number", Default: 9.99},
				},
			},
		},
	}
	exec := New(def)
	out, err := exec.Execute("create", map[string]interface{}{"name": "Widget"}, FormatOptions{Format: FormatJSON})
	if err != nil {
		t.Fatalf("Execute() POST unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "new-1") {
		t.Errorf("Execute() POST output = %q, want to contain 'new-1'", out)
	}
}

// ─── Auth header injection ─────────────────────────────────────────────────────

func TestExecute_WithBearerAuth(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	envKey := "ELYSIUM_TEST_BEARER_TOKEN_ABC"
	t.Setenv(envKey, "my-test-token")

	def := &emblem.Definition{
		APIVersion:  "v1",
		Name:        "test",
		Version:     "1.0.0",
		Description: "test",
		BaseURL:     server.URL,
		Auth:        emblem.Auth{Type: emblem.AuthBearer, KeyEnv: envKey},
		Actions: map[string]emblem.Action{
			"get": {Description: "Get resource", Method: "GET", Path: "/resource"},
		},
	}
	exec := New(def)
	_, err := exec.Execute("get", nil, FormatOptions{Format: FormatJSON})
	if err != nil {
		t.Fatalf("Execute() with bearer auth unexpected error: %v", err)
	}
	if receivedAuth != "Bearer my-test-token" {
		t.Errorf("Authorization header = %q, want %q", receivedAuth, "Bearer my-test-token")
	}
}

func TestExecute_WithAPIKeyAuth_CustomHeader(t *testing.T) {
	var receivedAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.Header.Get("X-Custom-Key")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	envKey := "ELYSIUM_TEST_API_KEY_DEF"
	t.Setenv(envKey, "secret-key")

	def := &emblem.Definition{
		APIVersion:  "v1",
		Name:        "test",
		Version:     "1.0.0",
		Description: "test",
		BaseURL:     server.URL,
		Auth:        emblem.Auth{Type: emblem.AuthAPIKey, KeyEnv: envKey, Header: "X-Custom-Key"},
		Actions: map[string]emblem.Action{
			"get": {Description: "Get resource", Method: "GET", Path: "/resource"},
		},
	}
	exec := New(def)
	_, err := exec.Execute("get", nil, FormatOptions{Format: FormatJSON})
	if err != nil {
		t.Fatalf("Execute() with API key auth unexpected error: %v", err)
	}
	if receivedAPIKey != "secret-key" {
		t.Errorf("X-Custom-Key header = %q, want %q", receivedAPIKey, "secret-key")
	}
}

func TestExecute_WithDefaultAPIKeyHeader(t *testing.T) {
	var receivedAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.Header.Get("X-API-Key")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	envKey := "ELYSIUM_TEST_API_KEY_GHI"
	t.Setenv(envKey, "default-header-key")

	def := &emblem.Definition{
		APIVersion:  "v1",
		Name:        "test",
		Version:     "1.0.0",
		Description: "test",
		BaseURL:     server.URL,
		// No Header field → defaults to "X-API-Key"
		Auth: emblem.Auth{Type: emblem.AuthAPIKey, KeyEnv: envKey},
		Actions: map[string]emblem.Action{
			"get": {Description: "Get resource", Method: "GET", Path: "/resource"},
		},
	}
	exec := New(def)
	_, err := exec.Execute("get", nil, FormatOptions{Format: FormatJSON})
	if err != nil {
		t.Fatalf("Execute() with default API key header unexpected error: %v", err)
	}
	if receivedAPIKey != "default-header-key" {
		t.Errorf("X-API-Key header = %q, want %q", receivedAPIKey, "default-header-key")
	}
}

// ─── extractBodyParams ─────────────────────────────────────────────────────────

func TestExtractBodyParams(t *testing.T) {
	exec := New(makeTestDef("http://example.com"))
	action := &emblem.Action{
		Method: "POST",
		Path:   "/items",
		Parameters: []emblem.Parameter{
			{Name: "name", In: "body"},
			{Name: "price", In: "body", Default: 9.99},
			{Name: "limit", In: "query"}, // must be ignored
		},
	}

	t.Run("provided param present in body", func(t *testing.T) {
		params := map[string]interface{}{"name": "Widget", "limit": 10}
		body := exec.extractBodyParams(action, params)
		if body["name"] != "Widget" {
			t.Errorf("body[name] = %v, want 'Widget'", body["name"])
		}
		if _, ok := body["limit"]; ok {
			t.Error("body should not contain 'limit' (it is a query param)")
		}
	})

	t.Run("default param applied when param absent", func(t *testing.T) {
		body := exec.extractBodyParams(action, map[string]interface{}{})
		if body["price"] != 9.99 {
			t.Errorf("body[price] = %v, want default 9.99", body["price"])
		}
	})

	t.Run("no parameters returns empty map", func(t *testing.T) {
		empty := &emblem.Action{Method: "POST", Path: "/"}
		body := exec.extractBodyParams(empty, nil)
		if len(body) != 0 {
			t.Errorf("expected empty body, got %v", body)
		}
	})
}

// ─── formatTable ──────────────────────────────────────────────────────────────

func TestFormatTable(t *testing.T) {
	t.Run("nil items returns no-results message", func(t *testing.T) {
		out, err := formatTable(nil)
		if err != nil {
			t.Fatalf("formatTable(nil) unexpected error: %v", err)
		}
		if string(out) != "No results\n" {
			t.Errorf("formatTable(nil) = %q, want %q", out, "No results\n")
		}
	})

	t.Run("empty items returns no-results message", func(t *testing.T) {
		out, err := formatTable([]interface{}{})
		if err != nil {
			t.Fatalf("formatTable([]) unexpected error: %v", err)
		}
		if string(out) != "No results\n" {
			t.Errorf("formatTable([]) = %q, want %q", out, "No results\n")
		}
	})

	t.Run("map items renders table with headers and rows", func(t *testing.T) {
		items := []interface{}{
			map[string]interface{}{"id": "1", "name": "Widget"},
			map[string]interface{}{"id": "2", "name": "Gadget"},
		}
		out, err := formatTable(items)
		if err != nil {
			t.Fatalf("formatTable() unexpected error: %v", err)
		}
		s := string(out)
		if !strings.Contains(s, "Widget") {
			t.Errorf("formatTable() output = %q, want to contain 'Widget'", s)
		}
		if !strings.Contains(s, "Gadget") {
			t.Errorf("formatTable() output = %q, want to contain 'Gadget'", s)
		}
	})
}

// ─── formatObject ─────────────────────────────────────────────────────────────

func TestFormatObject(t *testing.T) {
	obj := map[string]interface{}{"name": "Widget", "price": 9.99}
	out, err := formatObject(obj)
	if err != nil {
		t.Fatalf("formatObject() unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "Widget") {
		t.Errorf("formatObject() output = %q, want to contain 'Widget'", s)
	}
	if !strings.Contains(s, "{") || !strings.Contains(s, "}") {
		t.Errorf("formatObject() output should contain braces: %q", s)
	}
}

// ─── PrintRaw / PrintJSON ─────────────────────────────────────────────────────

func TestPrintRaw(t *testing.T) {
	if err := PrintRaw([]byte("hello world")); err != nil {
		t.Errorf("PrintRaw() error = %v, want nil", err)
	}
}

func TestPrintJSON(t *testing.T) {
	t.Run("valid JSON is formatted", func(t *testing.T) {
		if err := PrintJSON([]byte(`{"name":"Alice"}`)); err != nil {
			t.Errorf("PrintJSON() valid JSON error = %v, want nil", err)
		}
	})

	t.Run("invalid JSON falls back to raw output", func(t *testing.T) {
		if err := PrintJSON([]byte("not json")); err != nil {
			t.Errorf("PrintJSON() invalid JSON error = %v, want nil", err)
		}
	})
}

// ─── ParseParams ──────────────────────────────────────────────────────────────

func TestParseParams(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]string
		key      string
		wantStr  string
		wantKind string // "string", "array", "map"
	}{
		{
			name:     "plain string value passes through",
			flags:    map[string]string{"q": "hello"},
			key:      "q",
			wantStr:  "hello",
			wantKind: "string",
		},
		{
			name:     "JSON array is parsed",
			flags:    map[string]string{"ids": "[1,2,3]"},
			key:      "ids",
			wantKind: "array",
		},
		{
			name:     "JSON object is parsed",
			flags:    map[string]string{"filter": `{"status": "active"}`},
			key:      "filter",
			wantKind: "map",
		},
		{
			name:     "invalid JSON array falls back to string",
			flags:    map[string]string{"val": "[invalid"},
			key:      "val",
			wantStr:  "[invalid",
			wantKind: "string",
		},
		{
			name:     "invalid JSON object falls back to string",
			flags:    map[string]string{"val": "{invalid"},
			key:      "val",
			wantStr:  "{invalid",
			wantKind: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := ParseParams(tt.flags)
			val := params[tt.key]

			switch tt.wantKind {
			case "string":
				if val != tt.wantStr {
					t.Errorf("ParseParams[%q] = %v (%T), want %q", tt.key, val, val, tt.wantStr)
				}
			case "array":
				if _, ok := val.([]interface{}); !ok {
					t.Errorf("ParseParams[%q] = %T, want []interface{}", tt.key, val)
				}
			case "map":
				if _, ok := val.(map[string]interface{}); !ok {
					t.Errorf("ParseParams[%q] = %T, want map[string]interface{}", tt.key, val)
				}
			}
		})
	}
}
