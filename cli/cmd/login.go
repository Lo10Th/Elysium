package cmd

import (
	"fmt"
	"time"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/httpclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
