package cmd

import (
	"fmt"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display the currently logged in user",
	Long:  `Show the email and username of the currently authenticated user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()

		if cfg.Token == "" {
			fmt.Println("Not logged in. Run 'ely login' first.")
			return nil
		}

		registry := viper.GetString("registry")
		if registry == "" {
			registry = cfg.Registry
		}

		// Show user info
		if cfg.Username != "" && cfg.UserEmail != "" {
			fmt.Printf("Logged in as: %s (%s)\n", cfg.Username, cfg.UserEmail)
		} else if cfg.UserEmail != "" {
			fmt.Printf("Logged in as: %s\n", cfg.UserEmail)
		} else if cfg.Username != "" {
			fmt.Printf("Logged in as: %s\n", cfg.Username)
		} else {
			fmt.Println("Logged in (user details not available)")
		}

		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			fmt.Printf("\nRegistry: %s\n", registry)
			if cfg.Token != "" {
				fmt.Printf("Token: %s...\n", cfg.Token[:min(20, len(cfg.Token))])
			}
		}

		return nil
	},
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
