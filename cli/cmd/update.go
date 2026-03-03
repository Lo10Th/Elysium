package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/errfmt"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [emblem-name...]",
	Short: "Update installed emblems to the latest version",
	Long: `Update one or more installed emblems to their latest available version.

Examples:
  # Update all installed emblems
  ely update --all

  # Update specific emblems
  ely update clothing-shop stripe`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		cfg := config.Get()

		if !all && len(args) == 0 {
			return fmt.Errorf("specify emblem name(s) or use --all to update all installed emblems")
		}

		var names []string
		if all {
			for name := range cfg.Installed {
				names = append(names, name)
			}
			sort.Strings(names)
			if len(names) == 0 {
				fmt.Println("No emblems installed.")
				return nil
			}
			fmt.Println("Checking for updates...")
			fmt.Println()
		} else {
			names = args
		}

		client := api.NewClient()
		client.SetToken(cfg.Token)

		var failures []string
		for _, name := range names {
			currentVersion, installed := config.GetInstalledVersion(name)
			if !installed {
				fmt.Printf("%s: not installed\n", name)
				failures = append(failures, name)
				continue
			}

			emblemInfo, err := client.GetEmblem(name)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					fmt.Printf("%s: ✗ %v\n", name, errfmt.EmblemNotFoundError(name))
				} else {
					fmt.Printf("%s: ✗ %v\n", name, err)
				}
				failures = append(failures, name)
				continue
			}

			latestVersion := emblemInfo.LatestVersion
			if currentVersion == latestVersion {
				fmt.Printf("%s: %s (up to date)\n", name, currentVersion)
				continue
			}

			fmt.Printf("%s: %s → %s (update available)\n", name, currentVersion, latestVersion)

			if err := pullSingleEmblem(name, false); err != nil {
				fmt.Printf("  ✗ Failed to update: %v\n", err)
				failures = append(failures, name)
				continue
			}

			fmt.Printf("  ✓ Updated to %s\n", latestVersion)
		}

		if len(failures) > 0 {
			return fmt.Errorf("%d emblem(s) failed to update: %s", len(failures), strings.Join(failures, ", "))
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().Bool("all", false, "update all installed emblems")
	rootCmd.AddCommand(updateCmd)
}
