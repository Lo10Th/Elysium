package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display the currently logged in user",
	Long:  `Show the email and username of the currently authenticated user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := keyring.Get("elysium", "token")
		if err != nil {
			fmt.Println("Not logged in. Run 'ely login' first.")
			return nil
		}

		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			fmt.Printf("Token: %s...\n", token[:20])
		}

		registry := viper.GetString("registry")
		if registry == "" {
			registry = "https://registry.elysium.dev"
		}

		fmt.Printf("Logged in via: %s\n", registry)
		fmt.Println("Token is stored in system keyring.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
