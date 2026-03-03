package cmd

import (
	"fmt"
	"os"

	"github.com/elysium/elysium/cli/internal/scaffold"
	"github.com/spf13/cobra"
)

var (
	categoryFlag    string
	descriptionFlag string
	outputFlag      string
)

var initCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize a new emblem",
	Long: `Initialize a new emblem with template files.

Creates a directory structure with:
- emblem.yaml    # API definition
- README.md      # Documentation
- examples/      # Example requests
- .emblemignore  # Ignore patterns

Categories:
- payments  # Payment processing APIs
- ecommerce # E-commerce APIs
- auth      # Authentication APIs
- general   # Generic CRUD APIs (default)`,
	Example: `  # Initialize with default template
  ely init my-api

  # Initialize with category
  ely init payment-api --category payments

  # Initialize with description
  ely init shop-api --category ecommerce --description "Shop API"

  # Initialize in specific directory
  ely init my-api --output /path/to/dir`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&categoryFlag, "category", "c", "general", "Emblem category (payments, ecommerce, auth, general)")
	initCmd.Flags().StringVarP(&descriptionFlag, "description", "d", "", "Emblem description")
	initCmd.Flags().StringVarP(&outputFlag, "output", "o", ".", "Output directory")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := scaffold.ValidateName(name); err != nil {
		return err
	}

	outputDir := outputFlag
	if outputDir == "." {
		outputDir = name
	}

	targetDir := outputDir
	if outputDir == name {
		if _, err := os.Stat(targetDir); err == nil {
			return fmt.Errorf("directory '%s' already exists", targetDir)
		}
	}

	if err := scaffold.CreateDirectories(targetDir); err != nil {
		return err
	}

	tmpl := scaffold.EmblemTemplate{
		Name:        name,
		Category:    categoryFlag,
		Description: descriptionFlag,
		Version:     "1.0.0",
		BaseURL:     "https://api.example.com",
		Actions:     scaffold.GetCategoryTemplate(categoryFlag),
	}

	emblemPath := fmt.Sprintf("%s/emblem.yaml", targetDir)
	if err := scaffold.GenerateEmblem(tmpl, emblemPath); err != nil {
		return fmt.Errorf("failed to generate emblem.yaml: %w", err)
	}

	readmePath := fmt.Sprintf("%s/README.md", targetDir)
	if err := scaffold.GenerateREADME(tmpl, readmePath); err != nil {
		return fmt.Errorf("failed to generate README.md: %w", err)
	}

	examplesDir := fmt.Sprintf("%s/examples", targetDir)
	if err := scaffold.GenerateExamples(name, categoryFlag, examplesDir); err != nil {
		return fmt.Errorf("failed to generate examples: %w", err)
	}

	if err := scaffold.GenerateIgnoreFile(targetDir); err != nil {
		return fmt.Errorf("failed to generate .emblemignore: %w", err)
	}

	fmt.Printf("✓ Created emblem '%s' in ./%s/\n", name, targetDir)
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. Edit %s/emblem.yaml\n", targetDir)
	fmt.Printf("  2. Validate: ely validate %s/emblem.yaml\n", targetDir)
	fmt.Printf("  3. Test locally: ely test %s/ --local\n", targetDir)

	return nil
}
