package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
