package cmd

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateRandomState(t *testing.T) {
	state1, err := generateRandomState()
	if err != nil {
		t.Fatalf("generateRandomState() returned error: %v", err)
	}

	if state1 == "" {
		t.Error("generateRandomState() returned empty string")
	}

	state2, err := generateRandomState()
	if err != nil {
		t.Fatalf("generateRandomState() returned error: %v", err)
	}

	if state1 == state2 {
		t.Error("generateRandomState() returned same state twice")
	}

	decoded, err := base64.URLEncoding.DecodeString(state1)
	if err != nil {
		t.Fatalf("state is not valid base64: %v", err)
	}

	if len(decoded) != 32 {
		t.Errorf("state decoded to wrong length: got %d, want 32", len(decoded))
	}
}

func TestIsCommandAvailable(t *testing.T) {
	lsAvailable := isCommandAvailable("ls")
	if !lsAvailable {
		t.Error("isCommandAvailable('ls') returned false, expected true")
	}

	fakeAvailable := isCommandAvailable("this-command-does-not-exist-12345")
	if fakeAvailable {
		t.Error("isCommandAvailable('this-command-does-not-exist-12345') returned true, expected false")
	}
}

func TestOpenBrowser(t *testing.T) {
	err := openBrowser("https://example.com")
	if err != nil {
		if strings.Contains(err.Error(), "no browser command available") {
			return
		}
	}
}

func BenchmarkGenerateRandomState(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := generateRandomState()
		if err != nil {
			b.Fatalf("generateRandomState returned error: %v", err)
		}
	}
}

// --- saveTokenAndSuccess tests ---

func TestSaveTokenAndSuccess_NoToken(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	resp := &authResponse{AccessToken: ""}
	err := saveTokenAndSuccess(resp)
	if err == nil {
		t.Error("saveTokenAndSuccess() expected error for empty token, got nil")
	}
	if !strings.Contains(err.Error(), "no access token") {
		t.Errorf("saveTokenAndSuccess() error = %q, want to contain 'no access token'", err.Error())
	}
}

func TestSaveTokenAndSuccess_WithToken(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	resp := &authResponse{
		AccessToken:  "test-access-token-abc",
		RefreshToken: "test-refresh-token-xyz",
	}
	resp.User.Email = "user@example.com"
	resp.User.Username = "testuser"

	err := saveTokenAndSuccess(resp)
	if err != nil {
		t.Errorf("saveTokenAndSuccess() unexpected error: %v", err)
	}
}

func TestSaveTokenAndSuccess_EmailOnly(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	resp := &authResponse{AccessToken: "tok-email-only"}
	resp.User.Email = "only@example.com"

	err := saveTokenAndSuccess(resp)
	if err != nil {
		t.Errorf("saveTokenAndSuccess(email only) unexpected error: %v", err)
	}
}

func TestSaveTokenAndSuccess_NoUserDetails(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	resp := &authResponse{AccessToken: "tok-no-user"}

	err := saveTokenAndSuccess(resp)
	if err != nil {
		t.Errorf("saveTokenAndSuccess(no user) unexpected error: %v", err)
	}
}

// --- attemptLogin tests ---

func TestAttemptLogin_Success(t *testing.T) {
	authResp := authResponse{
		AccessToken:  "returned-token",
		RefreshToken: "returned-refresh",
	}
	authResp.User.Email = "user@example.com"
	authResp.User.Username = "testuser"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/login" || r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(authResp)
	}))
	defer server.Close()

	result, err := attemptLogin(server.URL, "user@example.com", "password123")
	if err != nil {
		t.Fatalf("attemptLogin() unexpected error: %v", err)
	}
	if result.AccessToken != "returned-token" {
		t.Errorf("attemptLogin() AccessToken = %q, want %q", result.AccessToken, "returned-token")
	}
}

func TestAttemptLogin_Failure_Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{})
	}))
	defer server.Close()

	_, err := attemptLogin(server.URL, "user@example.com", "wrongpassword")
	if err == nil {
		t.Error("attemptLogin() expected error for 401, got nil")
	}
}

func TestAttemptLogin_Failure_Detail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"detail": "Invalid credentials"})
	}))
	defer server.Close()

	_, err := attemptLogin(server.URL, "user@example.com", "wrongpassword")
	if err == nil {
		t.Error("attemptLogin() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid credentials") {
		t.Errorf("attemptLogin() error = %q, want to contain 'Invalid credentials'", err.Error())
	}
}

func TestAttemptLogin_Failure_Message(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Account locked"})
	}))
	defer server.Close()

	_, err := attemptLogin(server.URL, "user@example.com", "wrongpassword")
	if err == nil {
		t.Error("attemptLogin() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Account locked") {
		t.Errorf("attemptLogin() error = %q, want to contain 'Account locked'", err.Error())
	}
}

func TestAttemptLogin_ConnectionError(t *testing.T) {
	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()

	_, err := attemptLogin(closedURL, "user@example.com", "password")
	if err == nil {
		t.Error("attemptLogin() expected connection error, got nil")
	}
}
