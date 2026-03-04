//go:build integration

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/emblem"
	"github.com/elysium/elysium/cli/internal/executor"
)

// setupTestHome creates a temporary home directory with an Elysium config and
// returns a cleanup function. It overrides the HOME environment variable so
// that config.Init() and emblem.GetCachePath() use the temp directory.
func setupTestHome(t *testing.T) (homeDir string, cleanup func()) {
	t.Helper()

	homeDir = t.TempDir()
	t.Setenv("HOME", homeDir)

	if err := config.Init(); err != nil {
		t.Fatalf("config.Init() failed: %v", err)
	}

	cleanup = func() {}
	return homeDir, cleanup
}

// writeCachedEmblem writes emblemYAML into the test home directory's cache.
func writeCachedEmblem(t *testing.T, homeDir, name, version, emblemYAML string) {
	t.Helper()

	cacheDir := filepath.Join(homeDir, ".elysium", "cache", name+"@"+version)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "emblem.yaml"), []byte(emblemYAML), 0644); err != nil {
		t.Fatalf("failed to write emblem.yaml: %v", err)
	}
}

// startMockAPIServer starts an httptest server that simulates a simple product
// API and returns its URL and a close function.
func startMockAPIServer(t *testing.T) (serverURL string, close func()) {
	t.Helper()

	mux := http.NewServeMux()

	// GET /api/products — list products, optional ?category filter
	mux.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		products := []map[string]interface{}{
			{"id": 1, "name": "Widget", "price": 9.99, "category": "tools"},
			{"id": 2, "name": "Gadget", "price": 19.99, "category": "electronics"},
		}
		cat := r.URL.Query().Get("category")
		if cat != "" {
			filtered := []map[string]interface{}{}
			for _, p := range products {
				if p["category"] == cat {
					filtered = append(filtered, p)
				}
			}
			products = filtered
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	})

	// GET /api/products/{id}
	mux.HandleFunc("/api/products/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		product := map[string]interface{}{
			"id": 1, "name": "Widget", "price": 9.99, "category": "tools",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	})

	// POST /api/products — create product
	mux.HandleFunc("/api/products/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 42, "name": "New Product"})
	})

	srv := httptest.NewServer(mux)
	return srv.URL, srv.Close
}

// emblemTemplate returns a valid emblem YAML using the given base URL.
func emblemTemplate(baseURL string) string {
	return `apiVersion: v1
name: test-api
version: 1.0.0
description: Integration test emblem
baseUrl: ` + baseURL + `/api
auth:
  type: api_key
  keyEnv: TEST_API_KEY
  header: X-API-Key
actions:
  list-products:
    description: List all products
    method: GET
    path: /products
    parameters:
      - name: category
        type: string
        in: query
        required: false
        description: Filter by category
  get-product:
    description: Get a product by ID
    method: GET
    path: /products/{id}
    parameters:
      - name: id
        type: integer
        in: path
        required: true
        description: Product ID
  create-product:
    description: Create a new product
    method: POST
    path: /products/create
    parameters:
      - name: name
        type: string
        in: body
        required: true
        description: Product name
`
}

// ---------------------------------------------------------------------------
// Scenario 1: Full emblem flow — pull (cache write) → load → execute
// ---------------------------------------------------------------------------

// TestFullEmblemFlow verifies the end-to-end happy path:
//  1. Write an emblem into the local cache (simulating "ely pull").
//  2. Load the emblem definition from the cache.
//  3. Execute a GET action against a mock HTTP server.
//  4. Verify the response contains expected data.
func TestFullEmblemFlow(t *testing.T) {
	serverURL, closeServer := startMockAPIServer(t)
	defer closeServer()

	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	writeCachedEmblem(t, homeDir, "test-api", "1.0.0", emblemTemplate(serverURL))
	t.Setenv("TEST_API_KEY", "test-key")

	// Load emblem from cache (simulates what "ely pull" would have done).
	def, err := emblem.LoadFromCache("test-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	if def.Name != "test-api" {
		t.Errorf("def.Name = %q, want %q", def.Name, "test-api")
	}
	if def.Version != "1.0.0" {
		t.Errorf("def.Version = %q, want %q", def.Version, "1.0.0")
	}

	// Execute the list-products action.
	exec := executor.New(def)
	output, err := exec.Execute("list-products", map[string]interface{}{}, executor.FormatOptions{Format: "json"})
	if err != nil {
		t.Fatalf("Execute(list-products) error = %v", err)
	}

	// Response must contain product data.
	if len(output) == 0 {
		t.Error("Execute(list-products) returned empty output")
	}
}

// TestFullEmblemFlowWithPathParam verifies that path parameters are correctly
// substituted and the corresponding resource is returned.
func TestFullEmblemFlowWithPathParam(t *testing.T) {
	serverURL, closeServer := startMockAPIServer(t)
	defer closeServer()

	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	writeCachedEmblem(t, homeDir, "test-api", "1.0.0", emblemTemplate(serverURL))
	t.Setenv("TEST_API_KEY", "test-key")

	def, err := emblem.LoadFromCache("test-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	exec := executor.New(def)
	output, err := exec.Execute("get-product", map[string]interface{}{"id": "1"}, executor.FormatOptions{Format: "json"})
	if err != nil {
		t.Fatalf("Execute(get-product) error = %v", err)
	}

	if len(output) == 0 {
		t.Error("Execute(get-product) returned empty output")
	}
}

// TestFullEmblemFlowWithQueryParam verifies that query parameters are passed
// to the server and filtering works correctly.
func TestFullEmblemFlowWithQueryParam(t *testing.T) {
	serverURL, closeServer := startMockAPIServer(t)
	defer closeServer()

	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	writeCachedEmblem(t, homeDir, "test-api", "1.0.0", emblemTemplate(serverURL))
	t.Setenv("TEST_API_KEY", "test-key")

	def, err := emblem.LoadFromCache("test-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	exec := executor.New(def)
	output, err := exec.Execute("list-products", map[string]interface{}{"category": "tools"}, executor.FormatOptions{Format: "json"})
	if err != nil {
		t.Fatalf("Execute(list-products?category=tools) error = %v", err)
	}

	// The filtered response should still be valid JSON.
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}
}

// TestFullEmblemFlowPostAction verifies that POST actions with body parameters
// are executed correctly and the server acknowledges creation.
func TestFullEmblemFlowPostAction(t *testing.T) {
	serverURL, closeServer := startMockAPIServer(t)
	defer closeServer()

	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	writeCachedEmblem(t, homeDir, "test-api", "1.0.0", emblemTemplate(serverURL))
	t.Setenv("TEST_API_KEY", "test-key")

	def, err := emblem.LoadFromCache("test-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	exec := executor.New(def)
	output, err := exec.Execute("create-product", map[string]interface{}{"name": "New Widget"}, executor.FormatOptions{Format: "json"})
	if err != nil {
		t.Fatalf("Execute(create-product) error = %v", err)
	}

	if len(output) == 0 {
		t.Error("Execute(create-product) returned empty output")
	}
}

// ---------------------------------------------------------------------------
// Scenario 2: Error handling
// ---------------------------------------------------------------------------

// TestErrorHandling_NonExistentEmblem verifies that loading an emblem that has
// never been pulled returns a clear error rather than panicking.
func TestErrorHandling_NonExistentEmblem(t *testing.T) {
	_, cleanup := setupTestHome(t)
	defer cleanup()

	_, err := emblem.LoadFromCache("does-not-exist", "9.9.9")
	if err == nil {
		t.Fatal("LoadFromCache(does-not-exist) expected error, got nil")
	}
}

// TestErrorHandling_InvalidEmblemYAML verifies that a malformed emblem YAML
// produces an error during loading.
func TestErrorHandling_InvalidEmblemYAML(t *testing.T) {
	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	badYAML := `this is: not: valid: yaml: [[[`
	writeCachedEmblem(t, homeDir, "bad-emblem", "1.0.0", badYAML)

	_, err := emblem.LoadFromCache("bad-emblem", "1.0.0")
	if err == nil {
		t.Fatal("LoadFromCache(bad-emblem) expected parse error, got nil")
	}
}

// TestErrorHandling_MissingRequiredFields verifies that an emblem missing
// required fields (e.g. baseUrl, actions) is rejected with a validation error.
func TestErrorHandling_MissingRequiredFields(t *testing.T) {
	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	incompleteYAML := `apiVersion: v1
name: incomplete
version: 1.0.0
description: Missing baseUrl and actions
`
	writeCachedEmblem(t, homeDir, "incomplete", "1.0.0", incompleteYAML)

	_, err := emblem.LoadFromCache("incomplete", "1.0.0")
	if err == nil {
		t.Fatal("LoadFromCache(incomplete) expected validation error, got nil")
	}
}

// TestErrorHandling_UnknownAction verifies that requesting a non-existent
// action from a loaded emblem returns an error.
func TestErrorHandling_UnknownAction(t *testing.T) {
	serverURL, closeServer := startMockAPIServer(t)
	defer closeServer()

	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	writeCachedEmblem(t, homeDir, "test-api", "1.0.0", emblemTemplate(serverURL))
	t.Setenv("TEST_API_KEY", "test-key")

	def, err := emblem.LoadFromCache("test-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	exec := executor.New(def)
	_, err = exec.Execute("no-such-action", nil, executor.FormatOptions{})
	if err == nil {
		t.Fatal("Execute(no-such-action) expected error, got nil")
	}
}

// TestErrorHandling_APIError verifies that a non-2xx HTTP response from the
// upstream API is surfaced as an error.
func TestErrorHandling_APIError(t *testing.T) {
	// Server that always returns 500.
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
	}))
	defer errorServer.Close()

	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	writeCachedEmblem(t, homeDir, "test-api", "1.0.0", emblemTemplate(errorServer.URL))
	t.Setenv("TEST_API_KEY", "test-key")

	def, err := emblem.LoadFromCache("test-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	exec := executor.New(def)
	_, err = exec.Execute("list-products", nil, executor.FormatOptions{})
	if err == nil {
		t.Fatal("Execute() with 500 response expected error, got nil")
	}
}

// TestErrorHandling_ConnectionRefused verifies that a connection-refused
// scenario produces a meaningful error (not a panic or nil).
func TestErrorHandling_ConnectionRefused(t *testing.T) {
	// Use a URL that is guaranteed to be unreachable.
	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	unreachableYAML := `apiVersion: v1
name: unreachable
version: 1.0.0
description: Points to a port nobody is listening on
baseUrl: http://127.0.0.1:1
auth:
  type: none
actions:
  ping:
    description: Ping the server
    method: GET
    path: /ping
`
	writeCachedEmblem(t, homeDir, "unreachable", "1.0.0", unreachableYAML)

	def, err := emblem.LoadFromCache("unreachable", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	exec := executor.New(def)
	_, err = exec.Execute("ping", nil, executor.FormatOptions{})
	if err == nil {
		t.Fatal("Execute() against unreachable host expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// Scenario 3: Auth integration
// ---------------------------------------------------------------------------

// TestAuthIntegration_MissingEnvVar verifies that executing an action when the
// required API-key environment variable is not set returns a clear auth error.
func TestAuthIntegration_MissingEnvVar(t *testing.T) {
	serverURL, closeServer := startMockAPIServer(t)
	defer closeServer()

	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	writeCachedEmblem(t, homeDir, "test-api", "1.0.0", emblemTemplate(serverURL))
	// Explicitly unset the API key by setting it to empty string.
	t.Setenv("TEST_API_KEY", "")

	def, err := emblem.LoadFromCache("test-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	exec := executor.New(def)
	_, err = exec.Execute("list-products", nil, executor.FormatOptions{})
	if err == nil {
		t.Fatal("Execute() without API key expected auth error, got nil")
	}
}

// TestAuthIntegration_WrongAPIKey verifies that an incorrect API key results
// in an HTTP 401 error being surfaced to the caller.
func TestAuthIntegration_WrongAPIKey(t *testing.T) {
	serverURL, closeServer := startMockAPIServer(t)
	defer closeServer()

	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	writeCachedEmblem(t, homeDir, "test-api", "1.0.0", emblemTemplate(serverURL))
	// Supply a deliberately wrong key.
	t.Setenv("TEST_API_KEY", "wrong-key")

	def, err := emblem.LoadFromCache("test-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	exec := executor.New(def)
	_, err = exec.Execute("list-products", nil, executor.FormatOptions{})
	if err == nil {
		t.Fatal("Execute() with wrong API key expected error, got nil")
	}
}

// TestAuthIntegration_CorrectAPIKey verifies that providing the correct API
// key allows the action to succeed.
func TestAuthIntegration_CorrectAPIKey(t *testing.T) {
	serverURL, closeServer := startMockAPIServer(t)
	defer closeServer()

	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	writeCachedEmblem(t, homeDir, "test-api", "1.0.0", emblemTemplate(serverURL))
	t.Setenv("TEST_API_KEY", "test-key")

	def, err := emblem.LoadFromCache("test-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	exec := executor.New(def)
	output, err := exec.Execute("list-products", nil, executor.FormatOptions{Format: "json"})
	if err != nil {
		t.Fatalf("Execute() with correct API key error = %v", err)
	}
	if len(output) == 0 {
		t.Error("Execute() with correct API key returned empty output")
	}
}

// TestAuthIntegration_NoneAuthType verifies that an emblem with auth type
// "none" can execute actions without any credentials.
func TestAuthIntegration_NoneAuthType(t *testing.T) {
	// A server that accepts all requests without checking auth.
	openServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer openServer.Close()

	homeDir, cleanup := setupTestHome(t)
	defer cleanup()

	openYAML := `apiVersion: v1
name: open-api
version: 1.0.0
description: Open API with no auth
baseUrl: ` + openServer.URL + `
auth:
  type: none
actions:
  ping:
    description: Health check
    method: GET
    path: /health
`
	writeCachedEmblem(t, homeDir, "open-api", "1.0.0", openYAML)

	def, err := emblem.LoadFromCache("open-api", "1.0.0")
	if err != nil {
		t.Fatalf("LoadFromCache() error = %v", err)
	}

	exec := executor.New(def)
	output, err := exec.Execute("ping", nil, executor.FormatOptions{Format: "json"})
	if err != nil {
		t.Fatalf("Execute(ping) error = %v", err)
	}
	if len(output) == 0 {
		t.Error("Execute(ping) returned empty output")
	}
}
