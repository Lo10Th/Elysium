package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func loginWithBrowser(registry string) error {
	fmt.Println()
	fmt.Println("  Opening browser for authentication...")
	fmt.Println()

	// Request device code
	deviceResp, err := requestDeviceCode(registry)
	if err != nil {
		return fmt.Errorf("failed to start device login: %w", err)
	}

	// Display user code and URL
	fmt.Printf("  ┌─────────────────────────────────────────────────────┐\n")
	fmt.Printf("  │                                                     │\n")
	fmt.Printf("  │   Your device code:  \033[1;36m%s\033[0m              │\n", deviceResp.UserCode)
	fmt.Printf("  │                                                     │\n")
	fmt.Printf("  │   Open this URL:                                    │\n")
	fmt.Printf("  │   \033[4;34m%s\033[0m                    │\n", deviceResp.VerificationURI)
	fmt.Printf("  │                                                     │\n")
	fmt.Printf("  └─────────────────────────────────────────────────────┘\n")
	fmt.Println()

	// Try to open browser
	urlWithCode := fmt.Sprintf("%s?code=%s", deviceResp.VerificationURI, deviceResp.UserCode)
	if err := openBrowser(urlWithCode); err != nil {
		fmt.Println("  Could not open browser automatically. Please open the URL manually.")
	} else {
		fmt.Println("  Browser opened. Please authorize the CLI...")
	}

	fmt.Println()
	fmt.Printf("  Waiting for authorization")

	// Poll for token
	interval := deviceResp.Interval
	if interval == 0 {
		interval = 5
	}
	expiresAt := time.Now().Add(time.Duration(deviceResp.ExpiresIn) * time.Second)

	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIdx := 0

	for time.Now().Before(expiresAt) {
		time.Sleep(time.Duration(interval) * time.Second)

		// Print spinner
		fmt.Printf("\r  %s Waiting for authorization...", spinner[spinnerIdx])
		spinnerIdx = (spinnerIdx + 1) % len(spinner)

		tokenResp, err := pollForToken(registry, deviceResp.DeviceCode)
		if err != nil {
			// If "authorization pending", continue polling
			if strings.Contains(err.Error(), "pending") {
				continue
			}
			return fmt.Errorf("failed to check authorization: %w", err)
		}

		if tokenResp.AccessToken != "" {
			fmt.Printf("\r  ✓ Authorization successful!                    \n")
			fmt.Println()

			authResp := &authResponse{
				AccessToken:  tokenResp.AccessToken,
				RefreshToken: tokenResp.RefreshToken,
				TokenType:    tokenResp.TokenType,
			}
			authResp.User.Email = tokenResp.User.Email
			authResp.User.Username = tokenResp.User.Username

			return saveTokenAndSuccess(authResp)
		}
	}

	fmt.Printf("\r  ✗ Authorization timed out                      \n")
	return fmt.Errorf("device code expired - please try again")
}

func requestDeviceCode(registry string) (*deviceCodeResponse, error) {
	url := fmt.Sprintf("%s/api/auth/device/code", registry)
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := loginHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	var deviceResp deviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if deviceResp.Error != "" {
		return nil, fmt.Errorf("%s: %s", deviceResp.Error, deviceResp.Message)
	}

	return &deviceResp, nil
}

func pollForToken(registry, deviceCode string) (*deviceTokenResponse, error) {
	reqBody := map[string]string{"device_code": deviceCode}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/auth/device/token", registry)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := loginHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to poll for token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp deviceTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for pending status
	if resp.StatusCode == 400 {
		if tokenResp.Detail == "Authorization pending" || tokenResp.Message == "Authorization pending" {
			return nil, fmt.Errorf("authorization pending")
		}
		return nil, fmt.Errorf("%s", tokenResp.Detail)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("%s", tokenResp.Error)
	}

	return &tokenResp, nil
}
