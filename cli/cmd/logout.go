package cmd

import (
	"fmt"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from the Elysium registry",
	Long:  `Remove your stored credentials from the system keyring.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")

		if verbose {
			fmt.Println("Removing stored credentials...")
		}

		// Clear config
		if err := config.ClearAuth(); err != nil {
			if verbose {
				fmt.Printf("Warning: Could not clear config: %v\n", err)
			}
		}

		// Also clear from keyring if present
		err := keyring.Delete("elysium", "token")
		if err != nil && verbose {
			fmt.Printf("Note: No keyring entry to remove\n")
		}

		fmt.Println("✓ Logged out successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
