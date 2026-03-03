package cmd

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/errfmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Elysium registry",
	Long: `Authenticate with the Elysium registry by opening your browser
to the login page. A local server will receive the OAuth callback.`,
	RunE: runLogin,
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error"`
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	registry := viper.GetString("registry")
	if registry == "" {
		registry = "https://ely.karlharrenga.com"
	}

	state, err := generateRandomState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	port, err := findAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	if verbose {
		fmt.Printf("Starting local server on port %d...\n", port)
	}

	tokenChan := make(chan *tokenResponse, 1)
	errChan := make(chan error, 1)

	server := startLocalServer(port, state, tokenChan, errChan)
	if server == nil {
		return fmt.Errorf("failed to start local server")
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		server.Shutdown(ctx)
		cancel()
	}()

	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)
	loginURL := fmt.Sprintf("%s/api/auth/oauth/start?redirect_uri=%s&state=%s",
		registry,
		redirectURI,
		state)

	fmt.Println("Opening browser for authentication...")
	fmt.Println()
	fmt.Println("If the browser does not open, visit:")
	fmt.Printf("  %s\n", loginURL)
	fmt.Println()

	if err := openBrowser(loginURL); err != nil {
		fmt.Println("Could not open browser automatically. Please visit the URL manually.")
	}

	fmt.Println("Waiting for authentication...")

	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	select {
	case token := <-tokenChan:
		if token.Error != "" {
			return fmt.Errorf("authentication error: %s", token.Error)
		}

		if err := config.SetToken(token.AccessToken); err != nil {
			return fmt.Errorf("failed to store token: %w", err)
		}

		fmt.Println()
		fmt.Println("Authentication successful!")
		fmt.Println("Your token has been stored securely.")
		return nil

	case err := <-errChan:
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
			return errfmt.ConnectionError(registry, err)
		}
		if strings.Contains(err.Error(), "timeout") {
			return errfmt.NewDetailedError(err).
				WithReason("Authentication timed out").
				WithContext("Timeout", "5 minutes").
				WithSuggestion("Try again or check your network connection")
		}
		return errfmt.NetworkError(err)

	case <-timeout.C:
		return fmt.Errorf("authentication timed out after 5 minutes")
	}
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func findAvailablePort() (int, error) {
	for port := 8080; port <= 8090; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found between 8080-8090")
}

func startLocalServer(port int, expectedState string, tokenChan chan<- *tokenResponse, errChan chan<- error) *http.Server {
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if errMsg := query.Get("error"); errMsg != "" {
			tokenChan <- &tokenResponse{Error: errMsg}
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}

		state := query.Get("state")
		if state != expectedState {
			errChan <- fmt.Errorf("invalid state parameter")
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		accessToken := query.Get("access_token")
		refreshToken := query.Get("refresh_token")
		tokenType := query.Get("token_type")

		if accessToken == "" {
			errChan <- fmt.Errorf("no access token received")
			http.Error(w, "No access token", http.StatusBadRequest)
			return
		}

		tokenChan <- &tokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    tokenType,
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authentication Successful</title></head>
<body style="font-family: sans-serif; text-align: center; padding: 50px;">
<h1 style="color: #22c55e;">Authentication Successful!</h1>
<p>You can close this window and return to the CLI.</p>
<script>setTimeout(function() { window.close(); }, 2000);</script>
</body>
</html>`)
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	return server
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch {
	case isCommandAvailable("open"):
		cmd = exec.Command("open", url)
	case isCommandAvailable("xdg-open"):
		cmd = exec.Command("xdg-open", url)
	case isCommandAvailable("rundll32"):
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("no browser command available")
	}

	return cmd.Start()
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
