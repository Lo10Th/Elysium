package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/emblem"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate a shell completion script for the specified shell.

To load completions:

Bash:
  $ source <(ely completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ ely completion bash > /etc/bash_completion.d/ely
  # macOS:
  $ ely completion bash > $(brew --prefix)/etc/bash_completion.d/ely

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ ely completion zsh > "${fpath[1]}/_ely"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ ely completion fish | source

  # To load completions for each session, execute once:
  $ ely completion fish > ~/.config/fish/completions/ely.fish

PowerShell:
  PS> ely completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> ely completion powershell > ely.ps1
  # and source this file from your PowerShell profile.
`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(out)
		case "zsh":
			return cmd.Root().GenZshCompletion(out)
		case "fish":
			return cmd.Root().GenFishCompletion(out, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(out)
		default:
			return fmt.Errorf("unsupported shell %q, must be one of: bash, zsh, fish, powershell", args[0])
		}
	},
}

// installedEmblemNames returns the names of all locally installed emblems.
func installedEmblemNames() []string {
	cfg := config.Get()
	if cfg == nil {
		return nil
	}
	names := make([]string, 0, len(cfg.Installed))
	for name := range cfg.Installed {
		names = append(names, name)
	}
	return names
}

// emblemActionNames returns the action names for the given installed emblem.
func emblemActionNames(emblemName string) []string {
	cfg := config.Get()
	if cfg == nil {
		return nil
	}
	version, ok := cfg.Installed[emblemName]
	if !ok {
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	cachePath := filepath.Join(home, ".elysium", "cache",
		fmt.Sprintf("%s@%s", emblemName, version), "emblem.yaml")
	def, err := emblem.Load(cachePath)
	if err != nil {
		return nil
	}
	return def.ListActions()
}

// emblemActionParams returns the parameter names (as --flag strings) for the
// given emblem action so they can be offered as flag completions.
func emblemActionParams(emblemName, actionName string) []string {
	cfg := config.Get()
	if cfg == nil {
		return nil
	}
	version, ok := cfg.Installed[emblemName]
	if !ok {
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	cachePath := filepath.Join(home, ".elysium", "cache",
		fmt.Sprintf("%s@%s", emblemName, version), "emblem.yaml")
	def, err := emblem.Load(cachePath)
	if err != nil {
		return nil
	}
	action, err := def.GetAction(actionName)
	if err != nil {
		return nil
	}
	params := make([]string, 0, len(action.Parameters))
	for _, p := range action.Parameters {
		params = append(params, "--"+p.Name)
	}
	return params
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Dynamic completion for the `execute` command:
	//   Arg 0 → installed emblem names
	//   Arg 1 → actions of the chosen emblem
	//   Further args → parameter flags of the chosen action
	executeCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return installedEmblemNames(), cobra.ShellCompDirectiveNoFileComp
		case 1:
			return emblemActionNames(args[0]), cobra.ShellCompDirectiveNoFileComp
		default:
			return emblemActionParams(args[0], args[1]), cobra.ShellCompDirectiveNoFileComp
		}
	}
}
