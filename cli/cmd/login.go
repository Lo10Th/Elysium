package cmd

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/errfmt"
	"github.com/elysium/elysium/cli/internal/httpclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// loginHTTPClient is used for device-code and token-polling requests, where a
// 10-second client-level timeout prevents indefinite hangs.
var loginHTTPClient = httpclient.ClientWithTimeout(10 * time.Second)

var loginEmail string
var loginWeb bool

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Elysium registry",
	Long: `Authenticate with the Elysium registry.

By default, opens a browser for authentication (recommended).
Use --email flag for email/password authentication.`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().StringVarP(&loginEmail, "email", "e", "", "Email address (for password auth)")
	loginCmd.Flags().BoolVarP(&loginWeb, "web", "w", true, "Open browser for authentication")
	rootCmd.AddCommand(loginCmd)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	User         struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"user"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Error           string `json:"error"`
	Message         string `json:"message"`
}

type deviceTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	User         struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"user"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

type deviceStatusResponse struct {
	UserCode   string `json:"user_code"`
	Verified   bool   `json:"verified"`
	ClientName string `json:"client_name"`
	ExpiresAt  string `json:"expires_at"`
}

func runLogin(cmd *cobra.Command, args []string) error {
	registry := viper.GetString("registry")
	if registry == "" {
		registry = "https://ely.karlharrenga.com"
	}

	// If email is provided, use password auth
	if loginEmail != "" {
		return loginWithEmailPassword(registry)
	}

	// Otherwise, use browser-based device flow
	return loginWithBrowser(registry)
}

func loginWithEmailPassword(registry string) error {
	email := loginEmail
	if email == "" {
		fmt.Print("Email: ")
		reader := bufio.NewReader(os.Stdin)
		var err error
		email, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
		email = strings.TrimSpace(email)
	}

	if email == "" {
		return fmt.Errorf("email is required")
	}

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)

	if password == "" {
		return fmt.Errorf("password is required")
	}

	authResp, err := attemptLogin(registry, email, password)
	if err == nil && authResp.AccessToken != "" {
		return saveTokenAndSuccess(authResp)
	}

	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
			return errfmt.ConnectionError(registry, err)
		}
		return errfmt.NetworkError(err)
	}

	fmt.Println()
	fmt.Println("Authentication failed. Would you like to register a new account? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "yes" {
		return fmt.Errorf("login cancelled")
	}

	return registerUser(registry, email, password)
}

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
	bodyBytes, _ := json.Marshal(reqBody)

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

func attemptLogin(registry, email, password string) (*authResponse, error) {
	reqBody := loginRequest{
		Email:    email,
		Password: password,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/auth/login", registry)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpclient.DefaultClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != 200 {
		if authResp.Detail != "" {
			return nil, fmt.Errorf("%s", authResp.Detail)
		}
		if authResp.Message != "" {
			return nil, fmt.Errorf("%s", authResp.Message)
		}
		return nil, fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	return &authResp, nil
}

func registerUser(registry, email, password string) error {
	fmt.Println()
	fmt.Println("Creating a new account...")
	fmt.Print("Username: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	if username == "" {
		return fmt.Errorf("username is required")
	}

	fmt.Print("Confirm password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	confirmPassword := string(confirmBytes)

	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	reqBody := registerRequest{
		Email:    email,
		Password: password,
		Username: username,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/auth/register", registry)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpclient.DefaultClient().Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
			return errfmt.ConnectionError(registry, err)
		}
		return errfmt.NetworkError(err)
	}
	defer resp.Body.Close()

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		if authResp.Detail != "" {
			return fmt.Errorf("registration failed: %s", authResp.Detail)
		}
		if authResp.Message != "" {
			return fmt.Errorf("registration failed: %s", authResp.Message)
		}
		return fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	fmt.Println()
	fmt.Println("✓ Account created successfully!")
	return saveTokenAndSuccess(&authResp)
}

func saveTokenAndSuccess(authResp *authResponse) error {
	if authResp.AccessToken == "" {
		return fmt.Errorf("no access token received")
	}

	if err := config.SetToken(authResp.AccessToken); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	if authResp.RefreshToken != "" {
		if err := config.SetRefreshToken(authResp.RefreshToken); err != nil {
			fmt.Println("Warning: Could not store refresh token")
		}
	}

	if authResp.User.Email != "" {
		config.SetUserEmail(authResp.User.Email)
	}
	if authResp.User.Username != "" {
		config.SetUsername(authResp.User.Username)
	}

	fmt.Println()
	if authResp.User.Username != "" {
		fmt.Printf("✓ Logged in as %s (%s)\n", authResp.User.Username, authResp.User.Email)
	} else if authResp.User.Email != "" {
		fmt.Printf("✓ Logged in as %s\n", authResp.User.Email)
	} else {
		fmt.Println("✓ Logged in successfully!")
	}
	fmt.Println("Your credentials have been stored securely.")

	return nil
}

// generateRandomState creates a cryptographically random base64-encoded state string
// for use as CSRF protection in OAuth flows.
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// isCommandAvailable reports whether a command is available on the system PATH.
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// openBrowser attempts to open the given URL in the system's default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		for _, browser := range []string{"xdg-open", "google-chrome", "chromium", "firefox"} {
			if isCommandAvailable(browser) {
				cmd = exec.Command(browser, url)
				break
			}
		}
		if cmd == nil {
			return fmt.Errorf("no browser command available")
		}
	}
	return cmd.Start()
}
