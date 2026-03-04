package cmd

import (
	"testing"
)

func TestLogoutCmd_RunE(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// logoutCmd.RunE should complete without error even if keyring is empty.
	err := logoutCmd.RunE(logoutCmd, []string{})
	if err != nil {
		t.Errorf("logoutCmd.RunE() unexpected error: %v", err)
	}
}

func TestLogoutCmd_Verbose(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Access the verbose flag via the persistent flags on rootCmd.
	_ = rootCmd.PersistentFlags().Set("verbose", "true")
	defer func() { _ = rootCmd.PersistentFlags().Set("verbose", "false") }()

	err := logoutCmd.RunE(logoutCmd, []string{})
	if err != nil {
		t.Errorf("logoutCmd.RunE(verbose) unexpected error: %v", err)
	}
}
