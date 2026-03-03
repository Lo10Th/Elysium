package cmd

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/errfmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var loginEmail string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Elysium registry",
	Long: `Authenticate with the Elysium registry using email and password.

If you don't have an account, you'll be prompted to register.`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().StringVarP(&loginEmail, "email", "e", "", "Email address")
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

func runLogin(cmd *cobra.Command, args []string) error {
	registry := viper.GetString("registry")
	if registry == "" {
		registry = "https://ely.karlharrenga.com"
	}

	// Get email
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

	// Get password
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

	// Try to login first
	authResp, err := attemptLogin(registry, email, password)
	if err == nil && authResp.AccessToken != "" {
		// Login successful
		return saveTokenAndSuccess(authResp)
	}

	// Check if it was an authentication error vs network error
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
			return errfmt.ConnectionError(registry, err)
		}
		return errfmt.NetworkError(err)
	}

	// Login failed - check if user wants to register
	fmt.Println()
	fmt.Println("Authentication failed. Would you like to register a new account? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "yes" {
		return fmt.Errorf("login cancelled")
	}

	// Register
	return registerUser(registry, email, password)
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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for error response
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
	// Get username
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

	// Confirm password
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

	// Register
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

	client := &http.Client{}
	resp, err := client.Do(req)
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

	// Check for error response
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

	// Store token
	if err := config.SetToken(authResp.AccessToken); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Store refresh token if available
	if authResp.RefreshToken != "" {
		if err := config.SetRefreshToken(authResp.RefreshToken); err != nil {
			// Non-fatal, just log it
			fmt.Println("Warning: Could not store refresh token")
		}
	}

	// Store user info
	if authResp.User.Email != "" {
		config.SetUserEmail(authResp.User.Email)
	}
	if authResp.User.Username != "" {
		config.SetUsername(authResp.User.Username)
	}

	fmt.Println()
	if authResp.User.Username != "" {
		fmt.Printf("✓ Logged in as %s (%s)\n", authResp.User.Username, authResp.User.Email)
	} else {
		fmt.Printf("✓ Logged in as %s\n", authResp.User.Email)
	}
	fmt.Println("Your credentials have been stored securely.")

	return nil
}

// tokenResponse holds the OAuth token data received via callback.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error"`
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

// findAvailablePort finds an available TCP port in the range 8080–8090.
func findAvailablePort() (int, error) {
	for port := 8080; port <= 8090; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found in range 8080-8090")
}

// isCommandAvailable reports whether a command is available on the system PATH.
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// startLocalServer starts a local HTTP server on the given port to handle the
// OAuth callback. It sends the received token to tokenChan or an error to errChan.
func startLocalServer(port int, state string, tokenChan chan *tokenResponse, errChan chan error) *http.Server {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if errMsg := q.Get("error"); errMsg != "" {
			errChan <- fmt.Errorf("OAuth error: %s", errMsg)
			fmt.Fprintln(w, "Authentication failed. You can close this window.")
			return
		}

		if q.Get("state") != state {
			errChan <- fmt.Errorf("invalid OAuth state parameter")
			fmt.Fprintln(w, "Authentication failed: invalid state. You can close this window.")
			return
		}

		accessToken := q.Get("access_token")
		if accessToken == "" {
			errChan <- fmt.Errorf("no access token received in callback")
			fmt.Fprintln(w, "Authentication failed: no token received. You can close this window.")
			return
		}

		tokenChan <- &tokenResponse{
			AccessToken:  accessToken,
			RefreshToken: q.Get("refresh_token"),
			TokenType:    "bearer",
		}
		fmt.Fprintln(w, "Authentication successful! You can close this window and return to the CLI.")
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()
	return server
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
