package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/elysium/elysium/cli/internal/emblem"
)

func TestExecuteEmblemEndToEnd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		switch r.URL.Path {
		case "/api/products":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			products := []map[string]interface{}{
				{
					"id":          1,
					"name":        "Test Product",
					"description": "A test product",
					"price":       29.99,
					"category":    r.URL.Query().Get("category"),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(products)

		case "/api/products/1":
			if r.Method != "GET" {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			product := map[string]interface{}{
				"id":          1,
				"name":        "Test Product",
				"description": "A test product",
				"price":       29.99,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(product)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".elysium", "cache", "test-shop@1.0.0")
	os.MkdirAll(cacheDir, 0755)
	defer os.RemoveAll(filepath.Join(home, ".elysium", "cache", "test-shop@1.0.0"))

	emblemContent := `apiVersion: v1
name: test-shop
version: 1.0.0
description: Test emblem
baseUrl: ` + server.URL + `/api
auth:
  type: api_key
  keyEnv: TEST_SHOP_API_KEY
  header: X-API-Key
actions:
  list-products:
    description: List products
    method: GET
    path: /products
    parameters:
      - name: category
        type: string
        in: query
        required: false
        description: Filter by category
  get-product:
    description: Get a product
    method: GET
    path: /products/{id}
    parameters:
      - name: id
        type: integer
        in: path
        required: true
        description: Product ID
`
	os.WriteFile(filepath.Join(cacheDir, "emblem.yaml"), []byte(emblemContent), 0644)

	t.Setenv("TEST_SHOP_API_KEY", "test-api-key")

	configPath := filepath.Join(home, ".elysium", "config.yaml")
	configContent := `registry: https://ely.karlharrenga.com
cache_dir: ~/.elysium/cache
installed:
  test-shop: 1.0.0
`
	os.WriteFile(configPath, []byte(configContent), 0644)
	defer os.Remove(configPath)

	t.Run("list actions with no action", func(t *testing.T) {
		outputFormat = "table"
		err := executeEmblemAction("test-shop", []string{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})
}

func TestEmblemLoadAndExecute(t *testing.T) {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".elysium", "cache", "clothing-shop@1.0.0")
	os.MkdirAll(cacheDir, 0755)
	defer os.RemoveAll(filepath.Join(home, ".elysium", "cache", "clothing-shop@1.0.0"))

	emblemYAML := `apiVersion: v1
name: clothing-shop
version: 1.0.0
description: Clothing store API
baseUrl: http://localhost:5000/api
auth:
  type: api_key
  keyEnv: CLOTHING_SHOP_API_KEY
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
`

	err := os.WriteFile(filepath.Join(cacheDir, "emblem.yaml"), []byte(emblemYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write emblem: %v", err)
	}

	def, err := emblem.LoadFromCache("clothing-shop", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to load emblem: %v", err)
	}

	if def.Name != "clothing-shop" {
		t.Errorf("Expected name 'clothing-shop', got '%s'", def.Name)
	}

	if def.BaseURL != "http://localhost:5000/api" {
		t.Errorf("Expected baseUrl 'http://localhost:5000/api', got '%s'", def.BaseURL)
	}

	if len(def.Actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(def.Actions))
	}

	actions := def.ListActions()
	expectedActions := []string{"list-products", "get-product"}
	for _, expected := range expectedActions {
		found := false
		for _, action := range actions {
			if action == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing action: %s", expected)
		}
	}

	action, err := def.GetAction("list-products")
	if err != nil {
		t.Errorf("Failed to get action: %v", err)
	}
	if action.Method != "GET" {
		t.Errorf("Expected method GET, got %s", action.Method)
	}
	if action.Path != "/products" {
		t.Errorf("Expected path /products, got %s", action.Path)
	}
}

func TestParseParamsIntegration(t *testing.T) {
	t.Run("params with flags and JSON", func(t *testing.T) {
		paramsJSON = `{"limit": 10}`
		defer func() { paramsJSON = "" }()

		params, err := parseParams([]string{"--category", "shoes"})
		if err != nil {
			t.Fatalf("Failed to parse params: %v", err)
		}

		if params["category"] != "shoes" {
			t.Errorf("Expected category 'shoes', got '%v'", params["category"])
		}

		if params["limit"] != float64(10) {
			t.Errorf("Expected limit 10, got '%v'", params["limit"])
		}
	})

	t.Run("params with type conversion", func(t *testing.T) {
		params, err := parseParams([]string{"--id", "123", "--name", "test"})
		if err != nil {
			t.Fatalf("Failed to parse params: %v", err)
		}

		if params["id"] != "123" {
			t.Errorf("Expected id '123', got '%v'", params["id"])
		}

		if params["name"] != "test" {
			t.Errorf("Expected name 'test', got '%v'", params["name"])
		}
	})
}
