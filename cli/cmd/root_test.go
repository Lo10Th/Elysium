package cmd

import (
	"testing"
)

func TestIsKnownCommand_KnownCommands(t *testing.T) {
	known := []string{
		"check-updates", "execute", "help", "completion", "init",
		"info", "keys", "list", "login", "logout", "outdated",
		"pull", "search", "self-update", "test", "update",
		"validate", "whoami",
	}
	for _, cmd := range known {
		if !isKnownCommand(cmd) {
			t.Errorf("isKnownCommand(%q) = false, want true", cmd)
		}
	}
}

func TestIsKnownCommand_UnknownCommands(t *testing.T) {
	unknown := []string{
		"foobar",
		"clothing-shop",
		"stripe",
		"my-emblem",
		"",
		"PULL",  // case sensitive
		"Login", // case sensitive
	}
	for _, cmd := range unknown {
		if isKnownCommand(cmd) {
			t.Errorf("isKnownCommand(%q) = true, want false", cmd)
		}
	}
}

func TestInitConfigEarly_NoError(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	// initConfigEarly calls config.Init (already done) and viper setup.
	err := initConfigEarly()
	if err != nil {
		t.Errorf("initConfigEarly() unexpected error: %v", err)
	}
}
