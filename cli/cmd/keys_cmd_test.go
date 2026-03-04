package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
)

// newKeysServer creates a test server that handles all key API endpoints.
func newKeysServer(t *testing.T) *httptest.Server {
	t.Helper()
	expiry := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	keys := []api.Key{
		{
			ID:        "key-aaa",
			Name:      "ci-key",
			CreatedAt: time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			ID:        "key-bbb",
			Name:      "prod-key",
			CreatedAt: time.Date(2024, 4, 15, 8, 0, 0, 0, time.UTC),
			ExpiresAt: &expiry,
		},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "GET" && r.URL.Path == "/api/keys":
			json.NewEncoder(w).Encode(keys)

		case r.Method == "POST" && r.URL.Path == "/api/keys":
			newKey := api.Key{
				ID:        "key-new",
				Name:      keyName,
				Key:       "ely_supersecretkey",
				CreatedAt: time.Now(),
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newKey)

		case r.Method == "DELETE" && r.URL.Path == "/api/keys/key-aaa":
			w.WriteHeader(http.StatusNoContent)

		case r.Method == "GET" && r.URL.Path == "/api/keys/key-aaa":
			json.NewEncoder(w).Encode(keys[0])

		case r.Method == "GET" && r.URL.Path == "/api/keys/key-bbb":
			json.NewEncoder(w).Encode(keys[1])

		default:
			http.NotFound(w, r)
		}
	}))
}

// newEmptyKeysServer returns 200 with an empty list for GET /api/keys.
func newEmptyKeysServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]api.Key{})
	}))
}

// --- runKeysList tests ---

func TestRunKeysList_TableOutput(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newKeysServer(t)
	defer server.Close()
	config.Get().Registry = server.URL

	err := runKeysList(keysListCmd, []string{})
	if err != nil {
		t.Errorf("runKeysList() unexpected error: %v", err)
	}
}

func TestRunKeysList_JSONOutput(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newKeysServer(t)
	defer server.Close()
	config.Get().Registry = server.URL

	// Set output flag to json.
	if err := keysListCmd.Flags().Set("output", "json"); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}
	defer keysListCmd.Flags().Set("output", "table") //nolint:errcheck

	err := runKeysList(keysListCmd, []string{})
	if err != nil {
		t.Errorf("runKeysList(json) unexpected error: %v", err)
	}
}

func TestRunKeysList_Empty(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newEmptyKeysServer(t)
	defer server.Close()
	config.Get().Registry = server.URL

	err := runKeysList(keysListCmd, []string{})
	if err != nil {
		t.Errorf("runKeysList(empty) unexpected error: %v", err)
	}
}

func TestRunKeysList_ConnectionError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()
	config.Get().Registry = closedURL

	err := runKeysList(keysListCmd, []string{})
	if err == nil {
		t.Error("runKeysList() expected connection error, got nil")
	}
}

// --- runKeysCreate tests ---

func TestRunKeysCreate_Success(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newKeysServer(t)
	defer server.Close()
	config.Get().Registry = server.URL

	oldKeyName := keyName
	keyName = "test-ci-key"
	defer func() { keyName = oldKeyName }()

	err := runKeysCreate(keysCreateCmd, []string{})
	if err != nil {
		t.Errorf("runKeysCreate() unexpected error: %v", err)
	}
}

func TestRunKeysCreate_WithExpiry(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newKeysServer(t)
	defer server.Close()
	config.Get().Registry = server.URL

	oldKeyName := keyName
	oldKeyExpires := keyExpires
	keyName = "expiry-key"
	keyExpires = "2025-12-31T23:59:59Z"
	defer func() {
		keyName = oldKeyName
		keyExpires = oldKeyExpires
	}()

	err := runKeysCreate(keysCreateCmd, []string{})
	if err != nil {
		t.Errorf("runKeysCreate(with expiry) unexpected error: %v", err)
	}
}

func TestRunKeysCreate_InvalidExpiry(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	oldKeyName := keyName
	oldKeyExpires := keyExpires
	keyName = "bad-key"
	keyExpires = "not-a-date"
	defer func() {
		keyName = oldKeyName
		keyExpires = oldKeyExpires
	}()

	err := runKeysCreate(keysCreateCmd, []string{})
	if err == nil {
		t.Error("runKeysCreate() expected error for invalid date, got nil")
	}
}

func TestRunKeysCreate_ConnectionError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()
	config.Get().Registry = closedURL

	oldKeyName := keyName
	keyName = "test-key"
	defer func() { keyName = oldKeyName }()

	err := runKeysCreate(keysCreateCmd, []string{})
	if err == nil {
		t.Error("runKeysCreate() expected connection error, got nil")
	}
}

// --- runKeysDelete tests ---

func TestRunKeysDelete_Success(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newKeysServer(t)
	defer server.Close()
	config.Get().Registry = server.URL

	err := runKeysDelete(keysDeleteCmd, []string{"key-aaa"})
	if err != nil {
		t.Errorf("runKeysDelete() unexpected error: %v", err)
	}
}

func TestRunKeysDelete_ConnectionError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()
	config.Get().Registry = closedURL

	err := runKeysDelete(keysDeleteCmd, []string{"key-aaa"})
	if err == nil {
		t.Error("runKeysDelete() expected connection error, got nil")
	}
}

// --- runKeysShow tests ---

func TestRunKeysShow_TableOutput(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newKeysServer(t)
	defer server.Close()
	config.Get().Registry = server.URL

	err := runKeysShow(keysShowCmd, []string{"key-aaa"})
	if err != nil {
		t.Errorf("runKeysShow(table) unexpected error: %v", err)
	}
}

func TestRunKeysShow_TableOutput_WithExpiry(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newKeysServer(t)
	defer server.Close()
	config.Get().Registry = server.URL

	err := runKeysShow(keysShowCmd, []string{"key-bbb"})
	if err != nil {
		t.Errorf("runKeysShow(table+expiry) unexpected error: %v", err)
	}
}

func TestRunKeysShow_JSONOutput(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := newKeysServer(t)
	defer server.Close()
	config.Get().Registry = server.URL

	if err := keysShowCmd.Flags().Set("output", "json"); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}
	defer keysShowCmd.Flags().Set("output", "table") //nolint:errcheck

	err := runKeysShow(keysShowCmd, []string{"key-bbb"})
	if err != nil {
		t.Errorf("runKeysShow(json) unexpected error: %v", err)
	}
}

func TestRunKeysShow_ConnectionError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()
	config.Get().Registry = closedURL

	err := runKeysShow(keysShowCmd, []string{"key-aaa"})
	if err == nil {
		t.Error("runKeysShow() expected connection error, got nil")
	}
}

// --- print helper direct tests ---

func TestPrintKeysJSON_WithAndWithoutExpiry(t *testing.T) {
	expiry := time.Date(2025, 12, 31, 23, 59, 0, 0, time.UTC)
	keys := []api.Key{
		{
			ID:        "k1",
			Name:      "no-expiry",
			CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        "k2",
			Name:      "with-expiry",
			CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			ExpiresAt: &expiry,
		},
	}
	// Should not panic.
	printKeysJSON(keys)
}

func TestPrintKeyTable_NoExpiry(t *testing.T) {
	key := &api.Key{
		ID:        "k1",
		Name:      "no-expiry",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	printKeyTable(key) // Should not panic.
}

func TestPrintKeyTable_WithExpiry(t *testing.T) {
	expiry := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	key := &api.Key{
		ID:        "k2",
		Name:      "with-expiry",
		CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		ExpiresAt: &expiry,
	}
	printKeyTable(key) // Should not panic.
}

func TestPrintKeyJSON_NoExpiry(t *testing.T) {
	key := &api.Key{
		ID:        "k1",
		Name:      "no-expiry",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	printKeyJSON(key) // Should not panic.
}

func TestPrintKeyJSON_WithExpiry(t *testing.T) {
	expiry := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	key := &api.Key{
		ID:        "k2",
		Name:      "with-expiry",
		CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		ExpiresAt: &expiry,
	}
	printKeyJSON(key) // Should not panic.
}
