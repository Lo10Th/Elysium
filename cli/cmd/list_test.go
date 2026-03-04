package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
)

func TestListCmd_NoEmblems(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// Fresh config with no installed emblems.
	err := listCmd.RunE(listCmd, []string{})
	if err != nil {
		t.Errorf("listCmd.RunE() unexpected error: %v", err)
	}
}

func TestListCmd_WithInstalledEmblems(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("shop-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	if err := config.InstallEmblem("payment-api", "2.1.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	err := listCmd.RunE(listCmd, []string{})
	if err != nil {
		t.Errorf("listCmd.RunE() unexpected error: %v", err)
	}
}

func TestListCmd_VerboseMode(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	cfg := config.Get()

	// Write a cache file so GetCachePath returns a real path.
	if err := config.InstallEmblem("verbose-api", "3.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	cacheDir := filepath.Join(cfg.CacheDir, "verbose-api@3.0.0")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	verboseCmd := &cobra.Command{}
	verboseCmd.Flags().BoolP("verbose", "v", true, "verbose")
	_ = verboseCmd.Flags().Set("verbose", "true")

	// Use the root verbose flag.
	oldArgs := os.Args
	os.Args = []string{"ely", "--verbose", "list"}
	defer func() { os.Args = oldArgs }()

	// Call directly with verbose flag retrieved from root.
	err := listCmd.RunE(listCmd, []string{})
	if err != nil {
		t.Errorf("listCmd.RunE() verbose unexpected error: %v", err)
	}
}
