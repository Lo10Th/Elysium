package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os/exec"
	"runtime"
)

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
