package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elysium/elysium/cli/internal/emblem"
	"github.com/elysium/elysium/cli/internal/executor"
	"github.com/spf13/cobra"
)

var (
	actionName string
	dryRunFlag bool
)

var testCmd = &cobra.Command{
	Use:   "test <emblem-dir>",
	Short: "Test an emblem locally",
	Args:  cobra.ExactArgs(1),
	RunE: runTest,
}

func init() {
	testCmd.Flags().StringVarP(&actionName, "action", "a", "", "Action to test")
	testCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show request without executing")
	rootCmd.AddCommand(testCmd)
}

func runTest(cmd *cobra.Command, args []string) error {
	dir := args[0]
	emblemPath := filepath.Join(dir, "emblem.yaml")

	if _, err := os.Stat(emblemPath); os.IsNotExist(err) {
		return fmt.Errorf("emblem.yaml not found in %s", dir)
	}

	def, err := emblem.Load(emblemPath)
	if err != nil {
		return fmt.Errorf("failed to load emblem: %w", err)
	}

	if actionName == "" {
		fmt.Printf("Emblem: %s (%s)\n", def.Name, def.Version)
		fmt.Printf("Base URL: %s\n\n", def.BaseURL)
		fmt.Println("Available actions:")
		for name := range def.Actions {
			fmt.Printf("  - %s\n", name)
		}
		return nil
	}

	action, exists := def.Actions[actionName]
	if !exists {
		return fmt.Errorf("action '%s' not found", actionName)
	}

	if dryRunFlag {
		fmt.Printf("Action: %s\n", actionName)
		fmt.Printf("Method: %s\n", action.Method)
		fmt.Printf("Path: %s\n", action.Path)
		fmt.Printf("URL: %s%s\n", def.BaseURL, action.Path)
		return nil
	}

	exec := executor.New(def)
	result, err := exec.Execute(actionName, nil, "table")
	if err != nil {
		return err
	}

	fmt.Println(string(result))
	return nil
}
