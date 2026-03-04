package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRequestDeviceCode_Success verifies the happy path of requestDeviceCode.
func TestRequestDeviceCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/device/code" || r.Method != "POST" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		resp := deviceCodeResponse{
			DeviceCode:      "dev-code-123",
			UserCode:        "ABC-DEF",
			VerificationURI: "https://example.com/activate",
			ExpiresIn:       300,
			Interval:        5,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	result, err := requestDeviceCode(server.URL)
	if err != nil {
		t.Fatalf("requestDeviceCode() unexpected error: %v", err)
	}
	if result.DeviceCode != "dev-code-123" {
		t.Errorf("DeviceCode = %q, want %q", result.DeviceCode, "dev-code-123")
	}
	if result.UserCode != "ABC-DEF" {
		t.Errorf("UserCode = %q, want %q", result.UserCode, "ABC-DEF")
	}
}

// TestRequestDeviceCode_ErrorResponse verifies the path where the response body
// contains an Error field (server returns 200 but signals an error in JSON).
func TestRequestDeviceCode_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := deviceCodeResponse{
			Error:   "device_not_supported",
			Message: "Device flow is not supported for this client",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	_, err := requestDeviceCode(server.URL)
	if err == nil {
		t.Error("requestDeviceCode() expected error for error response, got nil")
	}
	if !strings.Contains(err.Error(), "device_not_supported") {
		t.Errorf("requestDeviceCode() error = %q, want to contain 'device_not_supported'", err.Error())
	}
}

// TestRequestDeviceCode_InvalidJSON verifies the decode-error path.
func TestRequestDeviceCode_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	_, err := requestDeviceCode(server.URL)
	if err == nil {
		t.Error("requestDeviceCode() expected decode error, got nil")
	}
}

// TestPollForToken_Success verifies that pollForToken returns a valid token when
// the server responds with an access token.
func TestPollForToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/device/token" || r.Method != "POST" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		resp := deviceTokenResponse{
			AccessToken:  "access-token-xyz",
			RefreshToken: "refresh-token-abc",
			TokenType:    "bearer",
		}
		resp.User.Email = "user@example.com"
		resp.User.Username = "testuser"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	result, err := pollForToken(server.URL, "test-device-code")
	if err != nil {
		t.Fatalf("pollForToken() unexpected error: %v", err)
	}
	if result.AccessToken != "access-token-xyz" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "access-token-xyz")
	}
}

// TestPollForToken_Pending verifies that pollForToken returns an error containing
// "pending" when the server signals authorization_pending.
func TestPollForToken_Pending(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := deviceTokenResponse{
			Detail: "Authorization pending",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	_, err := pollForToken(server.URL, "test-device-code")
	if err == nil {
		t.Error("pollForToken() expected error for pending authorization, got nil")
	}
	if !strings.Contains(err.Error(), "pending") {
		t.Errorf("pollForToken() error = %q, want to contain 'pending'", err.Error())
	}
}

// TestPollForToken_PendingMessage verifies the alternate "Authorization pending"
// message field path.
func TestPollForToken_PendingMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := deviceTokenResponse{
			Message: "Authorization pending",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	_, err := pollForToken(server.URL, "test-device-code")
	if err == nil {
		t.Error("pollForToken() expected error for pending (message field), got nil")
	}
	if !strings.Contains(err.Error(), "pending") {
		t.Errorf("pollForToken() error = %q, want to contain 'pending'", err.Error())
	}
}

// TestPollForToken_Error400 verifies that pollForToken returns an error when the
// server returns 400 with a non-pending detail message.
func TestPollForToken_Error400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := deviceTokenResponse{
			Detail: "Device code expired",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	_, err := pollForToken(server.URL, "expired-device-code")
	if err == nil {
		t.Error("pollForToken() expected error for 400 non-pending, got nil")
	}
}

// TestPollForToken_TokenError verifies the path where the response body has a
// non-empty Error field (indicating a protocol error).
func TestPollForToken_TokenError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := deviceTokenResponse{
			Error: "expired_token",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	_, err := pollForToken(server.URL, "test-device-code")
	if err == nil {
		t.Error("pollForToken() expected error for token error, got nil")
	}
	if !strings.Contains(err.Error(), "expired_token") {
		t.Errorf("pollForToken() error = %q, want to contain 'expired_token'", err.Error())
	}
}

// TestPollForToken_InvalidJSON verifies the decode-error path.
func TestPollForToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	_, err := pollForToken(server.URL, "test-device-code")
	if err == nil {
		t.Error("pollForToken() expected decode error, got nil")
	}
}

// TestLoginWithBrowser_DeviceCodeFails verifies that loginWithBrowser returns
// an error when requestDeviceCode fails (e.g. server returns invalid JSON).
func TestLoginWithBrowser_DeviceCodeFails(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	err := loginWithBrowser(server.URL)
	if err == nil {
		t.Error("loginWithBrowser() expected error when device code request fails, got nil")
	}
	if !strings.Contains(err.Error(), "failed to start device login") {
		t.Errorf("loginWithBrowser() error = %q, want to contain 'failed to start device login'", err.Error())
	}
}

// TestLoginWithBrowser_DeviceCodeError verifies loginWithBrowser handles the
// JSON error-field path (server returns 200 but with Error set in the body).
func TestLoginWithBrowser_DeviceCodeError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := deviceCodeResponse{
			Error:   "server_error",
			Message: "Unexpected server error",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	err := loginWithBrowser(server.URL)
	if err == nil {
		t.Error("loginWithBrowser() expected error for server_error, got nil")
	}
}

// TestLoginWithBrowser_Timeout verifies that loginWithBrowser returns a timeout
// error when the device code has already expired (ExpiresIn < 0), so the
// polling loop never executes.
func TestLoginWithBrowser_Timeout(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/device/code") && r.Method == "POST" {
			resp := deviceCodeResponse{
				DeviceCode:      "dev-code-timeout",
				UserCode:        "XYZ-123",
				VerificationURI: "https://example.com/activate",
				ExpiresIn:       -1, // already expired → loop is skipped immediately
				Interval:        5,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	err := loginWithBrowser(server.URL)
	if err == nil {
		t.Error("loginWithBrowser() expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("loginWithBrowser() error = %q, want to contain 'expired'", err.Error())
	}
}
