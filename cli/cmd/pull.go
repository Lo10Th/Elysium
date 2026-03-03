package cmd

import (
	"fmt"
	"strings"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/emblem"
	"github.com/elysium/elysium/cli/internal/errfmt"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull <emblem-name>[@version]",
	Short: "Download and cache an emblem",
	Long: `Download an emblem from the registry and cache it locally.

Examples:
  # Pull the latest version
  ely pull clothing-shop
  
  # Pull a specific version
  ely pull clothing-shop@1.0.0
  
  # Pull with version constraint
  ely pull clothing-shop@^1.0.0`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")

		name, version, err := emblem.ParseVersionConstraint(args[0])
		if err != nil {
			return err
		}

		if verbose {
			fmt.Printf("Pulling emblem %s@%s...\n", name, version)
		}

		cfg := config.Get()
		client := api.NewClient()
		client.SetToken(cfg.Token)

		emblemInfo, err := client.GetEmblem(name)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return errfmt.EmblemNotFoundError(name)
			}
			if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
				return errfmt.ConnectionError(cfg.Registry, err)
			}
			if strings.Contains(err.Error(), "timeout") {
				return errfmt.NewDetailedError(err).
					WithReason("Request timed out").
					WithSuggestion("Try again or check your network connection")
			}
			return errfmt.NetworkError(err)
		}

		if version == "latest" {
			version = emblemInfo.LatestVersion
		}

		ver, err := client.GetEmblemVersion(name, version)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return errfmt.NewDetailedError(fmt.Errorf("version '%s' not found for emblem '%s'", version, name)).
					WithSuggestion(fmt.Sprintf("Try: ely pull %s (to get latest version)", name))
			}
			if strings.Contains(err.Error(), "connection refused") {
				return errfmt.ConnectionError(cfg.Registry, err)
			}
			return errfmt.NetworkError(err)
		}

		if err := emblem.SaveToCache(name, version, []byte(ver.YAMLContent)); err != nil {
			return fmt.Errorf("failed to cache emblem: %w", err)
		}

		if err := config.InstallEmblem(name, version); err != nil {
			return fmt.Errorf("failed to update installed list: %w", err)
		}

		fmt.Printf("✓ Downloaded %s@%s\n", name, version)
		fmt.Printf("  Installed to: %s\n", config.GetCacheDir())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
