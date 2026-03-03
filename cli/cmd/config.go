package cmd

import (
	"fmt"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration",
	RunE:  runConfigList,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get config value",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configEmblemCmd = &cobra.Command{
	Use:   "emblem <name>",
	Short: "Get emblem configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigEmblem,
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configEmblemCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigList(cmd *cobra.Command, args []string) error {
	fmt.Println("Global Configuration:")
	fmt.Printf("  registry: %s\n", config.GetRegistry())
	fmt.Printf("  output: %s\n", config.GetOutput())
	fmt.Printf("  cache_dir: %s\n", config.GetCacheDir())

	fmt.Println("\nInstalled Emblems:")
	for name, version := range config.GetInstalledEmblems() {
		fmt.Printf("  %s: %s\n", name, version)
	}

	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	switch key {
	case "registry":
		fmt.Println(config.GetRegistry())
	case "output":
		fmt.Println(config.GetOutput())
	case "cache_dir":
		fmt.Println(config.GetCacheDir())
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return nil
}

func runConfigEmblem(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.GetEmblemConfig(name)
	if err != nil {
		return err
	}

	fmt.Printf("Configuration for %s:\n", name)
	for key, value := range cfg {
		fmt.Printf("  %s: %v\n", key, value)
	}

	return nil
}
