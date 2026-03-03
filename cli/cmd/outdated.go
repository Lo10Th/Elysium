package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
)

const (
	outdatedNameColumnWidth    = 24
	outdatedCurrentColumnWidth = 12
)

var outdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "Show installed emblems that have updates available",
	Long:  `Check all installed emblems against the registry and list any that have a newer version available.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()

		if len(cfg.Installed) == 0 {
			fmt.Println("No emblems installed.")
			return nil
		}

		client := api.NewClient()
		client.SetToken(cfg.Token)

		type outdatedEntry struct {
			name    string
			current string
			latest  string
		}

		var outdated []outdatedEntry
		var errors []string

		for name, currentVersion := range cfg.Installed {
			emblemInfo, err := client.GetEmblem(name)
			if err != nil {
				if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "connect: connection refused") {
					return err
				}
				errors = append(errors, fmt.Sprintf("%s: %v", name, err))
				continue
			}

			if emblemInfo.LatestVersion != currentVersion {
				outdated = append(outdated, outdatedEntry{
					name:    name,
					current: currentVersion,
					latest:  emblemInfo.LatestVersion,
				})
			}
		}

		for _, e := range errors {
			fmt.Printf("Warning: %s\n", e)
		}

		if len(outdated) == 0 {
			fmt.Println("All installed emblems are up to date.")
			return nil
		}

		sort.Slice(outdated, func(i, j int) bool { return outdated[i].name < outdated[j].name })

		fmt.Printf("%-*s %-*s %s\n", outdatedNameColumnWidth, "NAME", outdatedCurrentColumnWidth, "CURRENT", "LATEST")
		for _, e := range outdated {
			fmt.Printf("%-*s %-*s %s\n", outdatedNameColumnWidth, e.name, outdatedCurrentColumnWidth, e.current, e.latest)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(outdatedCmd)
}
