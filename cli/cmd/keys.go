package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage API keys",
	Long:  `Create, list, and delete API keys for the Elysium registry.`,
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API keys",
	Long:  `List all API keys associated with your account.`,
	RunE:  runKeysList,
}

var keysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	Long:  `Create a new API key for programmatic access to the Elysium registry.`,
	RunE:  runKeysCreate,
}

var keysDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an API key",
	Long:  `Delete an API key by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runKeysDelete,
}

var keysShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show key details",
	Long:  `Show details for a specific API key.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runKeysShow,
}

var (
	keyName    string
	keyExpires string
	keyOutput  string
)

func init() {
	rootCmd.AddCommand(keysCmd)
	keysCmd.AddCommand(keysListCmd)
	keysCmd.AddCommand(keysCreateCmd)
	keysCmd.AddCommand(keysDeleteCmd)
	keysCmd.AddCommand(keysShowCmd)

	keysCreateCmd.Flags().StringVarP(&keyName, "name", "n", "", "Name for the API key (required)")
	keysCreateCmd.Flags().StringVarP(&keyExpires, "expires", "e", "", "Expiration date (RFC3339, e.g., 2024-12-31T23:59:59Z)")
	keysCreateCmd.MarkFlagRequired("name")

	keysListCmd.Flags().StringVarP(&keyOutput, "output", "o", "table", "Output format (table, json)")

	keysShowCmd.Flags().StringVarP(&keyOutput, "output", "o", "table", "Output format (table, json)")
}

func runKeysList(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	client := api.NewClient()
	client.SetToken(cfg.Token)

	keys, err := client.ListKeys()
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	if len(keys) == 0 {
		fmt.Println("No API keys found.")
		fmt.Println("Create one with: ely keys create --name <name>")
		return nil
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		printKeysJSON(keys)
	default:
		printKeysTable(keys)
	}

	return nil
}

func runKeysCreate(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	client := api.NewClient()
	client.SetToken(cfg.Token)

	var expiresAt *time.Time
	if keyExpires != "" {
		t, err := time.Parse(time.RFC3339, keyExpires)
		if err != nil {
			return fmt.Errorf("invalid expiration date format. Use RFC3339 (e.g., 2024-12-31T23:59:59Z): %w", err)
		}
		expiresAt = &t
	}

	key, err := client.CreateKey(keyName, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	fmt.Println("\u2713 API key created successfully!")
	fmt.Println()
	fmt.Printf("ID:        %s\n", key.ID)
	fmt.Printf("Name:      %s\n", key.Name)
	if key.Key != "" {
		fmt.Println()
		fmt.Printf("Key:       %s\n", key.Key)
		fmt.Println()
		fmt.Println("\u26a0  Save this key now! You won't be able to see it again.")
	}
	if key.ExpiresAt != nil {
		fmt.Printf("Expires:   %s\n", key.ExpiresAt.Format("2006-01-02 15:04"))
	}

	return nil
}

func runKeysDelete(cmd *cobra.Command, args []string) error {
	keyID := args[0]

	cfg := config.Get()
	client := api.NewClient()
	client.SetToken(cfg.Token)

	err := client.DeleteKey(keyID)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	fmt.Printf("\u2713 API key '%s' deleted successfully\n", keyID)
	return nil
}

func runKeysShow(cmd *cobra.Command, args []string) error {
	keyID := args[0]

	cfg := config.Get()
	client := api.NewClient()
	client.SetToken(cfg.Token)

	key, err := client.GetKey(keyID)
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		printKeyJSON(key)
	default:
		printKeyTable(key)
	}

	return nil
}

func printKeysTable(keys []api.Key) {
	fmt.Printf("%-20s %-20s %-20s %-10s\n", "ID", "NAME", "CREATED", "EXPIRES")
	fmt.Println(strings.Repeat("-", 70))

	for _, key := range keys {
		expires := "Never"
		if key.ExpiresAt != nil {
			expires = key.ExpiresAt.Format("2006-01-02")
		}

		fmt.Printf("%-20s %-20s %-20s %-10s\n",
			key.ID,
			key.Name,
			key.CreatedAt.Format("2006-01-02 15:04"),
			expires,
		)
	}
}

func printKeysJSON(keys []api.Key) {
	fmt.Println("[")
	for i, key := range keys {
		fmt.Printf("  {\"id\": \"%s\", \"name\": \"%s\", \"created_at\": \"%s\"",
			key.ID, key.Name, key.CreatedAt.Format(time.RFC3339))
		if key.ExpiresAt != nil {
			fmt.Printf(", \"expires_at\": \"%s\"", key.ExpiresAt.Format(time.RFC3339))
		}
		if i < len(keys)-1 {
			fmt.Println("},")
		} else {
			fmt.Println("}")
		}
	}
	fmt.Println("]")
}

func printKeyTable(key *api.Key) {
	fmt.Printf("ID:        %s\n", key.ID)
	fmt.Printf("Name:      %s\n", key.Name)
	fmt.Printf("Created:   %s\n", key.CreatedAt.Format("2006-01-02 15:04"))
	if key.ExpiresAt != nil {
		fmt.Printf("Expires:   %s\n", key.ExpiresAt.Format("2006-01-02 15:04"))
	} else {
		fmt.Println("Expires:   Never")
	}
}

func printKeyJSON(key *api.Key) {
	fmt.Printf("{\"id\": \"%s\", \"name\": \"%s\", \"created_at\": \"%s\"",
		key.ID, key.Name, key.CreatedAt.Format(time.RFC3339))
	if key.ExpiresAt != nil {
		fmt.Printf(", \"expires_at\": \"%s\"", key.ExpiresAt.Format(time.RFC3339))
	}
	fmt.Println("}")
}
