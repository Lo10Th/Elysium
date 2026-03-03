package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Elysium registry",
	Long: `Authenticate with the Elysium registry by opening your browser
to the login page. Your credentials will be stored securely in 
your system keyring.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")

		if verbose {
			fmt.Println("Opening browser for authentication...")
		}

		registry := viper.GetString("registry")
		if registry == "" {
			registry = "https://registry.elysium.dev"
		}

		loginURL := registry + "/auth/login"

		if verbose {
			fmt.Printf("Login URL: %s\n", loginURL)
		}

		fmt.Println("Opening browser...")
		fmt.Println("Please complete authentication in your browser.")
		fmt.Println()
		fmt.Println("If the browser does not open, visit:")
		fmt.Printf("  %s\n", loginURL)
		fmt.Println()

		if err := exec.Command("open", loginURL).Start(); err != nil {
			if err := exec.Command("xdg-open", loginURL).Start(); err != nil {
				if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", loginURL).Start(); err != nil {
					fmt.Println("Could not open browser automatically. Please visit the URL manually.")
				}
			}
		}

		fmt.Println("Waiting for authentication...")
		fmt.Println("After completing login, your token will be stored securely.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
