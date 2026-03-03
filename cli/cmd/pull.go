package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/emblem"
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
			return fmt.Errorf("failed to get emblem: %w", err)
		}
		
		if version == "latest" {
			version = emblemInfo.LatestVersion
		}
		
		ver, err := client.GetEmblemVersion(name, version)
		if err != nil {
			return fmt.Errorf("failed to get emblem version: %w", err)
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

		if verbose {
			fmt.Printf("Pulling emblem %s@%s...\n", name, version)
		}

		cfg := config.Get()
		client := api.NewClient()
		client.SetToken(cfg.Token)

		emblem, err := client.GetEmblem(name)
		if err != nil {
			return fmt.Errorf("failed to get emblem: %w", err)
		}

		if version == "latest" {
			version = emblem.LatestVersion
		}

		ver, err := client.GetEmblemVersion(name, version)
		if err != nil {
			return fmt.Errorf("failed to get emblem version: %w", err)
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
