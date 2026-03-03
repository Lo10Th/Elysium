package cmd

import (
	"fmt"
	"os"

	"github.com/elysium/elysium/cli/internal/emblem"
	"github.com/elysium/elysium/cli/internal/validator"
	"github.com/spf13/cobra"
)

var strictMode bool

var validateCmd = &cobra.Command{
	Use:   "validate <emblem.yaml>",
	Short: "Validate an emblem YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: runValidate,
}

func init() {
	validateCmd.Flags().BoolVar(&strictMode, "strict", false, "Enable strict validation")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	def, err := emblem.Load(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse emblem: %w", err)
	}

	v := validator.New()
	errors := v.Validate(def)

	if len(errors) > 0 {
		fmt.Println("❌ Validation failed:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("validation failed")
	}

	if strictMode {
		errors = v.ValidateStrict(def)
		if len(errors) > 0 {
			fmt.Println("❌ Strict validation failed:")
			for _, err := range errors {
				fmt.Printf("  - %s\n", err)
			}
			return fmt.Errorf("strict validation failed")
		}
	}

	warnings := v.CheckBestPractices(def)
	if len(warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, warn := range warnings {
			fmt.Printf("  - %s\n", warn)
		}
	}

	fmt.Println("✓ Emblem is valid")
	return nil
}
