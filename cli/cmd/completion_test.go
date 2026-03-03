package cmd

import (
	"bytes"
	"strings"
	"testing"

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
