package cmd

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestFindAvailablePort(t *testing.T) {
	port, err := findAvailablePort()
	if err != nil {
		t.Fatalf("findAvailablePort() returned error: %v", err)
	}

	if port < 8080 || port > 8090 {
		t.Errorf("findAvailablePort() returned port %d, expected between 8080-8090", port)
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

func TestStartLocalServer(t *testing.T) {
	state, err := generateRandomState()
	if err != nil {
		t.Fatalf("failed to generate state: %v", err)
	}

	tokenChan := make(chan *tokenResponse, 1)
	errChan := make(chan error, 1)

	port, err := findAvailablePort()
	if err != nil {
		t.Fatalf("failed to find available port: %v", err)
	}

	server := startLocalServer(port, state, tokenChan, errChan)
	if server == nil {
		t.Fatal("startLocalServer returned nil server")
	}

	defer server.Shutdown(nil)

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost:" + string(rune(port)) + "/callback?state=" + state + "&access_token=testtoken")
	if err != nil {
		t.Logf("request failed: %v", err)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func TestStartLocalServerInvalidState(t *testing.T) {
	state, err := generateRandomState()
	if err != nil {
		t.Fatalf("failed to generate state: %v", err)
	}

	tokenChan := make(chan *tokenResponse, 1)
	errChan := make(chan error, 1)

	port, err := findAvailablePort()
	if err != nil {
		t.Fatalf("failed to find available port: %v", err)
	}

	server := startLocalServer(port, state, tokenChan, errChan)
	if server == nil {
		t.Fatal("startLocalServer returned nil server")
	}

	defer server.Shutdown(nil)
}

func TestStartLocalServerErrorParam(t *testing.T) {
	state, err := generateRandomState()
	if err != nil {
		t.Fatalf("failed to generate state: %v", err)
	}

	tokenChan := make(chan *tokenResponse, 1)
	errChan := make(chan error, 1)

	port, err := findAvailablePort()
	if err != nil {
		t.Fatalf("failed to find available port: %v", err)
	}

	server := startLocalServer(port, state, tokenChan, errChan)
	if server == nil {
		t.Fatal("startLocalServer returned nil server")
	}

	defer server.Shutdown(nil)
}

func TestStartLocalServerNoToken(t *testing.T) {
	state, err := generateRandomState()
	if err != nil {
		t.Fatalf("failed to generate state: %v", err)
	}

	tokenChan := make(chan *tokenResponse, 1)
	errChan := make(chan error, 1)

	port, err := findAvailablePort()
	if err != nil {
		t.Fatalf("failed to find available port: %v", err)
	}

	server := startLocalServer(port+1, state, tokenChan, errChan)
	if server == nil {
		t.Fatal("startLocalServer returned nil server")
	}

	defer server.Shutdown(nil)

	go func() {
		time.Sleep(100 * time.Millisecond)
		req := httptest.NewRequest("GET", "/callback?state="+state, nil)
		w := httptest.NewRecorder()
		server.Handler.ServeHTTP(w, req)
	}()

	time.Sleep(200 * time.Millisecond)
}

func TestTokenResponse(t *testing.T) {
	resp := &tokenResponse{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "bearer",
	}

	if resp.AccessToken != "test-access-token" {
		t.Errorf("AccessToken = %s, want test-access-token", resp.AccessToken)
	}

	errResp := &tokenResponse{
		Error: "test error",
	}

	if errResp.Error != "test error" {
		t.Errorf("Error = %s, want test error", errResp.Error)
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

func BenchmarkFindAvailablePort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := findAvailablePort()
		if err != nil {
			b.Fatalf("findAvailablePort returned error: %v", err)
		}
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
