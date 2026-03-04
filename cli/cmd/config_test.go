package cmd

import (
	"testing"

	"github.com/elysium/elysium/cli/internal/config"
)

func TestRunConfigList_NoInstalledEmblems(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	err := runConfigList(nil, []string{})
	if err != nil {
		t.Errorf("runConfigList() unexpected error: %v", err)
	}
}

func TestRunConfigList_WithInstalledEmblems(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("clothing-shop", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem() error: %v", err)
	}
	if err := config.InstallEmblem("stripe", "2.3.1"); err != nil {
		t.Fatalf("InstallEmblem() error: %v", err)
	}

	err := runConfigList(nil, []string{})
	if err != nil {
		t.Errorf("runConfigList() unexpected error: %v", err)
	}
}

func TestRunConfigGet_Registry(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	err := runConfigGet(nil, []string{"registry"})
	if err != nil {
		t.Errorf("runConfigGet(registry) unexpected error: %v", err)
	}
}

func TestRunConfigGet_Output(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	err := runConfigGet(nil, []string{"output"})
	if err != nil {
		t.Errorf("runConfigGet(output) unexpected error: %v", err)
	}
}

func TestRunConfigGet_CacheDir(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	err := runConfigGet(nil, []string{"cache_dir"})
	if err != nil {
		t.Errorf("runConfigGet(cache_dir) unexpected error: %v", err)
	}
}

func TestRunConfigGet_UnknownKey(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	err := runConfigGet(nil, []string{"unknown_key_xyz"})
	if err == nil {
		t.Error("runConfigGet(unknown_key) expected error, got nil")
	}
	if err.Error() != "unknown config key: unknown_key_xyz" {
		t.Errorf("runConfigGet() error = %q, want %q", err.Error(), "unknown config key: unknown_key_xyz")
	}
}

func TestRunConfigEmblem_Installed(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("clothing-shop", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem() error: %v", err)
	}

	err := runConfigEmblem(nil, []string{"clothing-shop"})
	if err != nil {
		t.Errorf("runConfigEmblem() unexpected error: %v", err)
	}
}

func TestRunConfigEmblem_NotInstalled(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// config.GetEmblemConfig always returns data (cache_dir, registry) even for
	// unknown emblems, so this should succeed without error.
	err := runConfigEmblem(nil, []string{"not-installed-emblem"})
	if err != nil {
		t.Errorf("runConfigEmblem() unexpected error: %v", err)
	}
}
