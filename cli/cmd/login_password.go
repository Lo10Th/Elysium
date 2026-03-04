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

	"github.com/elysium/elysium/cli/internal/errfmt"
	"github.com/elysium/elysium/cli/internal/httpclient"
	"golang.org/x/term"
)

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
	answer, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "yes" {
		return fmt.Errorf("login cancelled")
	}

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
