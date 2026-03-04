package cmd

import (
	"testing"

	"github.com/elysium/elysium/cli/internal/config"
)

// --- min helper ---

func TestMin_AlessThanB(t *testing.T) {
	if got := min(3, 7); got != 3 {
		t.Errorf("min(3, 7) = %d, want 3", got)
	}
}

func TestMin_AGreaterThanB(t *testing.T) {
	if got := min(10, 4); got != 4 {
		t.Errorf("min(10, 4) = %d, want 4", got)
	}
}

func TestMin_Equal(t *testing.T) {
	if got := min(5, 5); got != 5 {
		t.Errorf("min(5, 5) = %d, want 5", got)
	}
}

func TestMin_Zero(t *testing.T) {
	if got := min(0, 0); got != 0 {
		t.Errorf("min(0, 0) = %d, want 0", got)
	}
}

// --- whoamiCmd RunE tests ---

func TestWhoamiCmd_NotLoggedIn(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// No token set – should print "Not logged in" and return nil.
	err := whoamiCmd.RunE(whoamiCmd, []string{})
	if err != nil {
		t.Errorf("whoamiCmd.RunE() unexpected error: %v", err)
	}
}

func TestWhoamiCmd_LoggedIn_UsernameAndEmail(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	cfg := config.Get()
	cfg.Token = "test-token-abc"
	cfg.Username = "alice"
	cfg.UserEmail = "alice@example.com"

	err := whoamiCmd.RunE(whoamiCmd, []string{})
	if err != nil {
		t.Errorf("whoamiCmd.RunE() unexpected error: %v", err)
	}
}

func TestWhoamiCmd_LoggedIn_EmailOnly(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	cfg := config.Get()
	cfg.Token = "test-token-def"
	cfg.Username = ""
	cfg.UserEmail = "bob@example.com"

	err := whoamiCmd.RunE(whoamiCmd, []string{})
	if err != nil {
		t.Errorf("whoamiCmd.RunE() (email only) unexpected error: %v", err)
	}
}

func TestWhoamiCmd_LoggedIn_UsernameOnly(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	cfg := config.Get()
	cfg.Token = "test-token-ghi"
	cfg.Username = "charlie"
	cfg.UserEmail = ""

	err := whoamiCmd.RunE(whoamiCmd, []string{})
	if err != nil {
		t.Errorf("whoamiCmd.RunE() (username only) unexpected error: %v", err)
	}
}

func TestWhoamiCmd_LoggedIn_NoUserDetails(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	cfg := config.Get()
	cfg.Token = "test-token-jkl"
	cfg.Username = ""
	cfg.UserEmail = ""

	err := whoamiCmd.RunE(whoamiCmd, []string{})
	if err != nil {
		t.Errorf("whoamiCmd.RunE() (no user details) unexpected error: %v", err)
	}
}

func TestWhoamiCmd_Verbose(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	cfg := config.Get()
	cfg.Token = "test-token-verbose-abc123"
	cfg.Username = "dave"
	cfg.UserEmail = "dave@example.com"

	// Set verbose flag on the command.
	if err := whoamiCmd.Flags().Set("verbose", "true"); err != nil {
		// verbose is a persistent flag on rootCmd; skip if not reachable.
		t.Skip("verbose flag not available on whoamiCmd directly")
	}
	defer whoamiCmd.Flags().Set("verbose", "false") //nolint:errcheck

	err := whoamiCmd.RunE(whoamiCmd, []string{})
	if err != nil {
		t.Errorf("whoamiCmd.RunE() (verbose) unexpected error: %v", err)
	}
}
