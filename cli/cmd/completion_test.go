package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
)

func TestCompletionCmdRegistered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "completion" {
			found = true
			break
		}
	}
	if !found {
		t.Error("completion command not registered on rootCmd")
	}
}

func TestCompletionCmdValidArgs(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "completion" {
			for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
				found := false
				for _, v := range c.ValidArgs {
					if v == shell {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("completion command missing ValidArg %q", shell)
				}
			}
			return
		}
	}
	t.Fatal("completion command not found")
}

func TestCompletionBash(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	cmd := &cobra.Command{}
	cmd.SetArgs([]string{"completion", "bash"})

	out, err := executeCompletionForShell("bash")
	if err != nil {
		t.Fatalf("completion bash returned error: %v", err)
	}
	if !strings.Contains(out, "bash") {
		preview := out
		if len(preview) > 200 {
			preview = preview[:200]
		}
		t.Errorf("completion bash output does not look like a bash script, got: %s", preview)
	}
}

func TestCompletionZsh(t *testing.T) {
	out, err := executeCompletionForShell("zsh")
	if err != nil {
		t.Fatalf("completion zsh returned error: %v", err)
	}
	if len(out) == 0 {
		t.Error("completion zsh returned empty output")
	}
}

func TestCompletionFish(t *testing.T) {
	out, err := executeCompletionForShell("fish")
	if err != nil {
		t.Fatalf("completion fish returned error: %v", err)
	}
	if len(out) == 0 {
		t.Error("completion fish returned empty output")
	}
}

func TestCompletionPowerShell(t *testing.T) {
	out, err := executeCompletionForShell("powershell")
	if err != nil {
		t.Fatalf("completion powershell returned error: %v", err)
	}
	if len(out) == 0 {
		t.Error("completion powershell returned empty output")
	}
}

func TestCompletionInvalidShell(t *testing.T) {
	_, err := executeCompletionForShell("invalidshell")
	if err == nil {
		t.Error("expected error for invalid shell, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("expected 'unsupported shell' in error, got: %v", err)
	}
}

// executeCompletionForShell runs the completion command for the given shell
// and returns the output as a string.
func executeCompletionForShell(shell string) (string, error) {
	var buf bytes.Buffer
	completionCmd.SetOut(&buf)
	err := completionCmd.RunE(completionCmd, []string{shell})
	return buf.String(), err
}

func TestInstalledEmblemNames(t *testing.T) {
	// Without initialising the config the function should return gracefully.
	names := installedEmblemNames()
	// Either nil or an empty slice is acceptable when config is uninitialised.
	_ = names
}

func TestEmblemActionNames(t *testing.T) {
	// Should not panic when the emblem is not installed.
	actions := emblemActionNames("nonexistent-emblem")
	if actions != nil {
		t.Errorf("expected nil for unknown emblem, got %v", actions)
	}
}

func TestEmblemActionParams(t *testing.T) {
	// Should not panic when the emblem or action is not installed.
	params := emblemActionParams("nonexistent-emblem", "nonexistent-action")
	if params != nil {
		t.Errorf("expected nil for unknown emblem/action, got %v", params)
	}
}

func TestExecuteCmdHasValidArgsFunction(t *testing.T) {
	if executeCmd.ValidArgsFunction == nil {
		t.Error("executeCmd should have a ValidArgsFunction for shell completion")
	}
}

func TestExecuteValidArgsFunctionFirstArg(t *testing.T) {
	// With no config initialised, first-arg completion should return empty/nil
	// without panicking.
	completions, directive := executeCmd.ValidArgsFunction(executeCmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
	_ = completions
}

func TestExecuteValidArgsFunctionSecondArg(t *testing.T) {
	// With no config initialised, second-arg completion should return empty/nil
	// without panicking.
	completions, directive := executeCmd.ValidArgsFunction(executeCmd, []string{"any-emblem"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
	_ = completions
}

func TestExecuteValidArgsFunctionFurtherArgs(t *testing.T) {
	// Parameter-flag completion should return empty/nil without panicking when
	// the emblem is not installed.
	completions, directive := executeCmd.ValidArgsFunction(executeCmd, []string{"any-emblem", "any-action"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
	_ = completions
}

// --- completion helper tests with real config ---

func TestInstalledEmblemNames_WithConfig(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	names := installedEmblemNames()
	if len(names) != 0 {
		t.Errorf("installedEmblemNames() = %v, want empty for fresh config", names)
	}
}

func TestInstalledEmblemNames_WithInstalledEmblems(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("alpha-shop", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	if err := config.InstallEmblem("beta-api", "2.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}

	names := installedEmblemNames()
	if len(names) != 2 {
		t.Errorf("installedEmblemNames() len = %d, want 2", len(names))
	}
}

// writeTestEmblemToCache writes a minimal emblem YAML to the config cache dir.
func writeTestEmblemToCache(t *testing.T, name, version string) {
	t.Helper()
	cfg := config.Get()
	cacheEntry := filepath.Join(cfg.CacheDir, fmt.Sprintf("%s@%s", name, version))
	if err := os.MkdirAll(cacheEntry, 0755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	emblemYAML := fmt.Sprintf(`apiVersion: v1
name: %s
version: %s
description: Completion test emblem
baseUrl: http://localhost:5000/api
auth:
  type: none
actions:
  list:
    description: List items
    method: GET
    path: /items
    parameters:
      - name: limit
        type: integer
        in: query
        required: false
  create:
    description: Create item
    method: POST
    path: /items
`, name, version)
	if err := os.WriteFile(filepath.Join(cacheEntry, "emblem.yaml"), []byte(emblemYAML), 0644); err != nil {
		t.Fatalf("writeTestEmblemToCache: %v", err)
	}
}

func TestEmblemActionNames_WithInstalledAndCachedEmblem(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("test-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	writeTestEmblemToCache(t, "test-api", "1.0.0")

	actions := emblemActionNames("test-api")
	if len(actions) == 0 {
		t.Error("emblemActionNames() = empty, want at least one action")
	}
}

func TestEmblemActionNames_NotInstalled_WithConfig(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	actions := emblemActionNames("never-installed")
	if actions != nil {
		t.Errorf("emblemActionNames(never-installed) = %v, want nil", actions)
	}
}

func TestEmblemActionNames_InstalledNoCacheFile(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("missing-cache", "3.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	// Intentionally do NOT write emblem to cache.

	actions := emblemActionNames("missing-cache")
	if actions != nil {
		t.Errorf("emblemActionNames(missing-cache) = %v, want nil when no cache", actions)
	}
}

func TestEmblemActionParams_WithCachedEmblem(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("param-api", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	writeTestEmblemToCache(t, "param-api", "1.0.0")

	params := emblemActionParams("param-api", "list")
	if len(params) == 0 {
		t.Error("emblemActionParams() = empty, want at least --limit")
	}
}

func TestEmblemActionParams_InvalidAction_WithConfig(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	if err := config.InstallEmblem("param-api2", "1.0.0"); err != nil {
		t.Fatalf("InstallEmblem: %v", err)
	}
	writeTestEmblemToCache(t, "param-api2", "1.0.0")

	params := emblemActionParams("param-api2", "does-not-exist")
	if params != nil {
		t.Errorf("emblemActionParams(invalid-action) = %v, want nil", params)
	}
}

func TestEmblemActionParams_NotInstalled_WithConfig(t *testing.T) {
	cleanup := initTestConfig(t)
	defer cleanup()

	params := emblemActionParams("ghost-api", "list")
	if params != nil {
		t.Errorf("emblemActionParams(not-installed) = %v, want nil", params)
	}
}
