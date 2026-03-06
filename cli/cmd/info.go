package cmd

import (
	"fmt"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/emblem"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <emblem-name>[@version]",
	Short: "Show detailed information about an emblem",
	Long: `Display information about an emblem including its description,
available actions, authentication requirements, and installed version.

If the emblem is installed locally, it will read from cache.
Otherwise, it will fetch from the remote registry.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")

		name, version, err := emblem.ParseVersionConstraint(args[0])
		if err != nil {
			return err
		}

		cfg := config.Get()

		if verbose {
			fmt.Printf("Fetching info for %s@%s...\n", name, version)
		}

		var def *emblem.Definition

		installed, ok := cfg.Installed[name]
		if ok && (version == "latest" || version == installed) {
			if verbose {
				fmt.Println("Reading from local cache...")
			}
			def, err = emblem.LoadFromCache(name, installed)
			if err != nil {
				fmt.Printf("Failed to load from cache: %v\n", err)
				fmt.Println("Fetching from registry...")
				def = nil
			}
		}

		if def == nil {
			client := api.NewClient()
			client.SetToken(cfg.Token)

			if version == "latest" {
				emblemInfo, err := client.GetEmblem(name)
				if err != nil {
					return fmt.Errorf("failed to get emblem info: %w", err)
				}
				version = emblemInfo.LatestVersion
			}

			ver, err := client.GetEmblemVersion(name, version)
			if err != nil {
				return fmt.Errorf("failed to get emblem version: %w", err)
			}

			def, err = emblem.Parse([]byte(ver.YAMLContent))
			if err != nil {
				return fmt.Errorf("failed to parse emblem: %w", err)
			}
		}

		fmt.Println()
		fmt.Printf("  %s (%s)\n", def.Name, def.Version)
		fmt.Println()
		fmt.Printf("  Description:  %s\n", def.Description)
		fmt.Printf("  Base URL:     %s\n", def.BaseURL)
		author := def.Author
		if author == "" {
			author = "Unknown"
		}
		fmt.Printf("  Author:       %s\n", author)
		fmt.Printf("  License:      %s\n", def.License)
		if verbose {
			fmt.Printf("  Category:     %s\n", def.Category)
			fmt.Printf("  Tags:         %v\n", def.Tags)
		}

		fmt.Println()
		fmt.Println("  Authentication:")
		fmt.Printf("    Type:       %s\n", def.Auth.Type)
		if def.Auth.Type != "none" {
			fmt.Printf("    Env Var:    %s\n", def.Auth.KeyEnv)
			if def.Auth.Header != "" {
				fmt.Printf("    Header:     %s\n", def.Auth.Header)
			}
		}

		fmt.Println()
		fmt.Println("  Available Actions:")
		for actionName, action := range def.Actions {
			fmt.Printf("    %-25s %s\n", actionName, action.Description)
		}

		fmt.Println()
		fmt.Printf("  Use: ely %s <action> [flags]\n", def.Name)
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
