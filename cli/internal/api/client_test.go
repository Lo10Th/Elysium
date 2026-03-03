package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elysium/elysium/cli/internal/config"
)

func init() {
	// Initialize config for tests
	config.Init()
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Error("NewClient() returned nil")
	}
	if client.baseURL == "" {
		t.Error("NewClient() client has empty baseURL")
	}
}

func TestNewClientWithBaseURL(t *testing.T) {
	client := NewClientWithBaseURL("http://localhost:8080")
	if client == nil {
		t.Error("NewClientWithBaseURL() returned nil")
	}
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("NewClientWithBaseURL() baseURL = %v, want http://localhost:8080", client.baseURL)
	}
}

func TestSetToken(t *testing.T) {
	client := NewClientWithBaseURL("http://localhost:8080")
	client.SetToken("test-token")
	// Token is set internally, verified through subsequent requests
}

func TestSetBaseURL(t *testing.T) {
	client := NewClientWithBaseURL("http://localhost:8080")
	client.SetBaseURL("http://newhost:9090")
	if client.baseURL != "http://newhost:9090" {
		t.Errorf("SetBaseURL() baseURL = %v, want http://newhost:9090", client.baseURL)
	}
}

func TestListEmblems(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		category   string
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/emblems" {
					t.Errorf("ListEmblems() path = %v, want /api/emblems", r.URL.Path)
				}
				emblems := []Emblem{
					{ID: "1", Name: "test-api", Description: "Test API"},
				}
				json.NewEncoder(w).Encode(emblems)
			},
			wantErr: false,
		},
		{
			name: "success with category",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("category") != "payments" {
					t.Errorf("ListEmblems() category = %v, want payments", r.URL.Query().Get("category"))
				}
				emblems := []Emblem{
					{ID: "1", Name: "stripe", Category: "payments"},
				}
				json.NewEncoder(w).Encode(emblems)
			},
			category: "payments",
			wantErr:  false,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "internal error"})
			},
			wantErr:    true,
			errContain: "500",
		},
		{
			name: "malformed json",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClientWithBaseURL(server.URL)
			emblems, err := client.ListEmblems(tt.category, 20, 0)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListEmblems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(emblems) == 0 {
				// Some tests may return empty list legitimately
				if emblems == nil {
					t.Error("ListEmblems() returned nil")
				}
			}
		})
	}
}

func TestSearchEmblems(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		query      string
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/emblems/search" {
					t.Errorf("SearchEmblems() path = %v, want /api/emblems/search", r.URL.Path)
				}
				if r.URL.Query().Get("q") != "api" {
					t.Errorf("SearchEmblems() query = %v, want 'api'", r.URL.Query().Get("q"))
				}
				emblems := []Emblem{
					{ID: "1", Name: "test-api"},
				}
				json.NewEncoder(w).Encode(emblems)
			},
			query:   "api",
			wantErr: false,
		},
		{
			name: "with category and sort",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("category") != "payments" {
					t.Errorf("SearchEmblems() category = %v, want payments", r.URL.Query().Get("category"))
				}
				if r.URL.Query().Get("sort") != "downloads" {
					t.Errorf("SearchEmblems() sort = %v, want downloads", r.URL.Query().Get("sort"))
				}
				emblems := []Emblem{}
				json.NewEncoder(w).Encode(emblems)
			},
			query:   "api",
			wantErr: false,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "not found"})
			},
			query:   "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClientWithBaseURL(server.URL)
			category := ""
			sort := ""
			if tt.name == "with category and sort" {
				category = "payments"
				sort = "downloads"
			}
			emblems, err := client.SearchEmblems(tt.query, category, sort, 20, 0)

			if (err != nil) != tt.wantErr {
				t.Errorf("SearchEmblems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && emblems == nil {
				t.Error("SearchEmblems() returned nil")
			}
		})
	}
}

func TestGetEmblem(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		emblemName string
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/emblems/test-api" {
					t.Errorf("GetEmblem() path = %v, want /api/emblems/test-api", r.URL.Path)
				}
				emblem := Emblem{ID: "1", Name: "test-api", Description: "Test API"}
				json.NewEncoder(w).Encode(emblem)
			},
			emblemName: "test-api",
			wantErr:    false,
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "not found"})
			},
			emblemName: "nonexistent",
			wantErr:    true,
			errContain: "not found",
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "internal error"})
			},
			emblemName: "test-api",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClientWithBaseURL(server.URL)
			emblem, err := client.GetEmblem(tt.emblemName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetEmblem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if emblem == nil {
					t.Error("GetEmblem() returned nil")
				}
				if emblem != nil && emblem.Name != tt.emblemName && tt.name == "success" {
					t.Errorf("GetEmblem() Name = %v, want %v", emblem.Name, tt.emblemName)
				}
			}
		})
	}
}

func TestGetEmblemVersion(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		emblemName  string
		version     string
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/emblems/test-api/1.0.0" {
					t.Errorf("GetEmblemVersion() path = %v, want /api/emblems/test-api/1.0.0", r.URL.Path)
				}
				ver := EmblemVersion{
					Name:        "test-api",
					Version:     "1.0.0",
					YAMLContent: "apiVersion: v1\nname: test-api",
				}
				json.NewEncoder(w).Encode(ver)
			},
			emblemName: "test-api",
			version:    "1.0.0",
			wantErr:    false,
		},
		{
			name: "version not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "version not found"})
			},
			emblemName:  "test-api",
			version:     "99.0.0",
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClientWithBaseURL(server.URL)
			ver, err := client.GetEmblemVersion(tt.emblemName, tt.version)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetEmblemVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && ver == nil {
				t.Error("GetEmblemVersion() returned nil")
			}
		})
	}
}

func TestPublishEmblem(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("PublishEmblem() method = %v, want POST", r.Method)
				}
				if r.URL.Path != "/api/emblems" {
					t.Errorf("PublishEmblem() path = %v, want /api/emblems", r.URL.Path)
				}
				emblem := Emblem{ID: "1", Name: "test-api"}
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(emblem)
			},
			wantErr: false,
		},
		{
			name: "auth required",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
			},
			wantErr:     true,
			errContains: "401",
		},
		{
			name: "validation error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid name"})
			},
			wantErr:     true,
			errContains: "400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClientWithBaseURL(server.URL)
			emblem, err := client.PublishEmblem("test-api", "Test API", "yaml: content", "1.0.0", "general", []string{"test"})

			if (err != nil) != tt.wantErr {
				t.Errorf("PublishEmblem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && emblem == nil {
				t.Error("PublishEmblem() returned nil")
			}
		})
	}
}

func TestListKeys(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("ListKeys() method = %v, want GET", r.Method)
				}
				if r.URL.Path != "/api/keys" {
					t.Errorf("ListKeys() path = %v, want /api/keys", r.URL.Path)
				}
				keys := []Key{
					{ID: "1", Name: "test-key"},
				}
				json.NewEncoder(w).Encode(keys)
			},
			wantErr: false,
		},
		{
			name: "auth required",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
			},
			wantErr:     true,
			errContains: "401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClientWithBaseURL(server.URL)
			keys, err := client.ListKeys()

			if (err != nil) != tt.wantErr {
				t.Errorf("ListKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && keys == nil {
				t.Error("ListKeys() returned nil")
			}
		})
	}
}

func TestCreateKey(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		keyName     string
		expiresAt   *time.Time
		wantErr     bool
		errContains string
	}{
		{
			name: "success without expiration",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("CreateKey() method = %v, want POST", r.Method)
				}
				if r.URL.Path != "/api/keys" {
					t.Errorf("CreateKey() path = %v, want /api/keys", r.URL.Path)
				}
				key := Key{ID: "1", Name: "test-key", Key: "sk_test_123"}
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(key)
			},
			keyName: "test-key",
			wantErr: false,
		},
		{
			name: "success with expiration",
			handler: func(w http.ResponseWriter, r *http.Request) {
				key := Key{ID: "1", Name: "temp-key", Key: "sk_test_456"}
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(key)
			},
			keyName:   "temp-key",
			expiresAt: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
			wantErr:   false,
		},
		{
			name: "auth required",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
			},
			keyName:     "test-key",
			wantErr:     true,
			errContains: "401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClientWithBaseURL(server.URL)
			key, err := client.CreateKey(tt.keyName, tt.expiresAt)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && key == nil {
				t.Error("CreateKey() returned nil")
			}
		})
	}
}

func TestGetKey(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		keyID       string
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("GetKey() method = %v, want GET", r.Method)
				}
				if r.URL.Path != "/api/keys/key-123" {
					t.Errorf("GetKey() path = %v, want /api/keys/key-123", r.URL.Path)
				}
				key := Key{ID: "key-123", Name: "test-key"}
				json.NewEncoder(w).Encode(key)
			},
			keyID:   "key-123",
			wantErr: false,
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "key not found"})
			},
			keyID:       "nonexistent",
			wantErr:     true,
			errContains: "404",
		},
		{
			name: "auth required",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
			},
			keyID:       "key-123",
			wantErr:     true,
			errContains: "401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClientWithBaseURL(server.URL)
			key, err := client.GetKey(tt.keyID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && key == nil {
				t.Error("GetKey() returned nil")
			}
		})
	}
}

func TestDeleteKey(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		keyID       string
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					t.Errorf("DeleteKey() method = %v, want DELETE", r.Method)
				}
				if r.URL.Path != "/api/keys/key-123" {
					t.Errorf("DeleteKey() path = %v, want /api/keys/key-123", r.URL.Path)
				}
				w.WriteHeader(http.StatusNoContent)
			},
			keyID:   "key-123",
			wantErr: false,
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "key not found"})
			},
			keyID:       "nonexistent",
			wantErr:     true,
			errContains: "404",
		},
		{
			name: "auth required",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"})
			},
			keyID:       "key-123",
			wantErr:     true,
			errContains: "401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewClientWithBaseURL(server.URL)
			err := client.DeleteKey(tt.keyID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
