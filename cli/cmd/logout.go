package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"keyring"
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

		err := keyring.Delete("elysium", "token")
		if err != nil {
			fmt.Println("No credentials stored or already logged out.")
			return nil
		}

		fmt.Println("Logged out successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
