package cmd

import (
	"fmt"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/elysium/elysium/cli/internal/emblem"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List installed emblems",
	Long:    `List all emblems that are currently installed in your local cache.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")

		cfg := config.Get()

		if len(cfg.Installed) == 0 {
			fmt.Println("No emblems installed.")
			fmt.Println()
			fmt.Println("Install an emblem with: ely pull <emblem-name>")
			return nil
		}

		fmt.Println("Installed emblems:")
		fmt.Println()

		for name, version := range cfg.Installed {
			if verbose {
				cachePath, _ := emblem.GetCachePath(name, version)
				fmt.Printf("  %-30s %s\n", name+"@"+version, cachePath)
			} else {
				fmt.Printf("  %-30s %s\n", name, version)
			}
		}

		fmt.Println()
		fmt.Printf("Total: %d emblem(s)\n", len(cfg.Installed))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
