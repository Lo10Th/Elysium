package cmd

import (
	"fmt"

	"github.com/elysium/elysium/cli/internal/api"
	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for emblems in the registry",
	Long: `Search for emblems by name or description.

Examples:
  # Search for payment APIs
  ely search payment
  
  # Search in a specific category
  ely search payment --category payments
  
  # Limit results
  ely search api --limit 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		category, _ := cmd.Flags().GetString("category")
		sort, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		verbose, _ := cmd.Flags().GetBool("verbose")

		cfg := config.Get()
		client := api.NewClient()
		client.SetToken(cfg.Token)

		emblems, err := client.SearchEmblems(query, category, sort, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to search emblems: %w", err)
		}

		if len(emblems) == 0 {
			fmt.Println("No emblems found matching your query.")
			return nil
		}

		fmt.Println()
		fmt.Printf("  Search Results for \"%s\" (%d found)\n", query, len(emblems))
		fmt.Println()
		fmt.Printf("  %-30s %-15s %s\n", "NAME", "VERSION", "DESCRIPTION")
		fmt.Println("  " + "─" + string(make([]byte, 80)))

		for _, e := range emblems {
			desc := e.Description
			if len(desc) > 45 {
				desc = desc[:42] + "..."
			}
			fmt.Printf("  %-30s %-15s %s\n", e.Name, e.LatestVersion, desc)
			if e.AuthorName != "" && verbose {
				verified := ""
				if e.AuthorVerified {
					verified = " ✓"
				}
				fmt.Printf("    by %s%s\n", e.AuthorName, verified)
			}
		}

		fmt.Println()
		fmt.Println("  Install: ely pull <emblem-name>")
		fmt.Println("  Info:    ely info <emblem-name>")
		fmt.Println()

		return nil
	},
}

func init() {
	searchCmd.Flags().String("category", "", "Filter by category")
	searchCmd.Flags().String("sort", "downloads", "Sort by: downloads, recent, name")
	searchCmd.Flags().Int("limit", 20, "Limit number of results")
	searchCmd.Flags().Int("offset", 0, "Offset for pagination")
	searchCmd.Flags().Bool("verbose", false, "Show detailed information including author")

	rootCmd.AddCommand(searchCmd)
}
