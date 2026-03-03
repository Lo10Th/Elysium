package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elysium/elysium/cli/internal/api"
)

func TestListKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/keys" {
			t.Errorf("Expected /api/keys, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		keys := []api.Key{
			{
				ID:        "key-123",
				Name:      "test-key",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			{
				ID:        "key-456",
				Name:      "prod-key",
				CreatedAt: time.Date(2024, 1, 16, 9, 15, 0, 0, time.UTC),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(keys)
	}))
	defer server.Close()

	t.Run("lists keys successfully", func(t *testing.T) {
		client := api.NewClientWithBaseURL(server.URL)

		keys, err := client.ListKeys()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(keys) != 2 {
			t.Errorf("Expected 2 keys, got %d", len(keys))
		}

		if keys[0].Name != "test-key" {
			t.Errorf("Expected 'test-key', got '%s'", keys[0].Name)
		}
	})
}

func TestCreateKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/keys" {
			t.Errorf("Expected /api/keys, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		if req["name"] != "new-key" {
			t.Errorf("Expected name 'new-key', got '%v'", req["name"])
		}

		key := api.Key{
			ID:        "key-new",
			Name:      "new-key",
			Key:       "ely_secretkey123",
			CreatedAt: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(key)
	}))
	defer server.Close()

	t.Run("creates key successfully", func(t *testing.T) {
		client := api.NewClientWithBaseURL(server.URL)

		key, err := client.CreateKey("new-key", nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if key.Name != "new-key" {
			t.Errorf("Expected 'new-key', got '%s'", key.Name)
		}

		if key.Key != "ely_secretkey123" {
			t.Errorf("Expected key to be returned, got empty")
		}
	})
}

func TestDeleteKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/keys/key-123" {
			t.Errorf("Expected /api/keys/key-123, got %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Run("deletes key successfully", func(t *testing.T) {
		client := api.NewClientWithBaseURL(server.URL)

		err := client.DeleteKey("key-123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}

func TestGetKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/keys/key-123" {
			t.Errorf("Expected /api/keys/key-123, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		key := api.Key{
			ID:        "key-123",
			Name:      "test-key",
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(key)
	}))
	defer server.Close()

	t.Run("gets key details successfully", func(t *testing.T) {
		client := api.NewClientWithBaseURL(server.URL)

		key, err := client.GetKey("key-123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if key.Name != "test-key" {
			t.Errorf("Expected 'test-key', got '%s'", key.Name)
		}
	})
}

func TestCreateKeyWithExpiration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		if req["expires_days"] == nil {
			t.Error("Expected expires_days to be set")
		}

		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		key := api.Key{
			ID:        "key-exp",
			Name:      "expiring-key",
			CreatedAt: time.Now(),
			ExpiresAt: &expiresAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(key)
	}))
	defer server.Close()

	t.Run("creates key with expiration", func(t *testing.T) {
		client := api.NewClientWithBaseURL(server.URL)

		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		key, err := client.CreateKey("expiring-key", &expiresAt)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if key.Name != "expiring-key" {
			t.Errorf("Expected 'expiring-key', got '%s'", key.Name)
		}
	})
}

func TestUnauthorizedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
	}))
	defer server.Close()

	t.Run("returns error on unauthorized", func(t *testing.T) {
		client := api.NewClientWithBaseURL(server.URL)

		_, err := client.ListKeys()
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if err.Error() != "API error: Unauthorized" {
			t.Errorf("Expected 'API error: Unauthorized', got '%s'", err.Error())
		}
	})
}

func TestNotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Key not found",
		})
	}))
	defer server.Close()

	t.Run("returns error on not found", func(t *testing.T) {
		client := api.NewClientWithBaseURL(server.URL)

		_, err := client.GetKey("nonexistent")
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if err.Error() != "API error: Key not found" {
			t.Errorf("Expected 'API error: Key not found', got '%s'", err.Error())
		}
	})
}

func TestPrintKeysTable(t *testing.T) {
	keys := []api.Key{
		{
			ID:        "key-123",
			Name:      "test-key",
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			ID:        "key-456",
			Name:      "prod-key",
			CreatedAt: time.Date(2024, 1, 16, 9, 15, 0, 0, time.UTC),
		},
	}

	t.Run("print keys table does not panic", func(t *testing.T) {
		printKeysTable(keys)
	})
}
