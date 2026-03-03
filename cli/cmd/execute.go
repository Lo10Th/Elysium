package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/elysium/elysium/cli/internal/emblem"
	"github.com/elysium/elysium/cli/internal/executor"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	paramsJSON   string
	paramsFile   string
	outputFormat string
)

func isInstalledEmblem(name string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	configPath := filepath.Join(home, ".elysium", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	type Config struct {
		Installed map[string]string `yaml:"installed"`
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return false
	}

	version, ok := cfg.Installed[name]
	if !ok {
		return false
	}

	cachePath := filepath.Join(home, ".elysium", "cache", fmt.Sprintf("%s@%s", name, version), "emblem.yaml")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func executeEmblemAction(emblemName string, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".elysium", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	type Config struct {
		Installed map[string]string `yaml:"installed"`
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	installedVersion, ok := cfg.Installed[emblemName]
	if !ok {
		return fmt.Errorf("emblem '%s' is not installed. Run: ely pull %s", emblemName, emblemName)
	}

	var actionName string
	if len(args) > 0 {
		actionName = args[0]
	}

	cachePath := filepath.Join(home, ".elysium", "cache", fmt.Sprintf("%s@%s", emblemName, installedVersion), "emblem.yaml")
	def, err := emblem.Load(cachePath)
	if err != nil {
		return fmt.Errorf("failed to load emblem: %w", err)
	}

	if actionName == "" {
		fmt.Printf("Emblem: %s (%s)\n", def.Name, def.Version)
		fmt.Printf("Description: %s\n\n", def.Description)
		fmt.Println("Available actions:")
		for name := range def.Actions {
			fmt.Printf("  - %s\n", name)
		}
		fmt.Printf("\nUsage: ely %s <action> [flags]\n", emblemName)
		return nil
	}

	action, exists := def.Actions[actionName]
	if !exists {
		fmt.Printf("Unknown action: %s\n\n", actionName)
		fmt.Println("Available actions:")
		for name := range def.Actions {
			fmt.Printf("  - %s\n", name)
		}
		return fmt.Errorf("action not found")
	}

	_ = action

	params, err := parseParams(args[1:])
	if err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}

	exec := executor.New(def)
	result, err := exec.Execute(actionName, params, outputFormat)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	fmt.Println(string(result))
	return nil
}

func parseParams(flagArgs []string) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	if paramsJSON != "" {
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return nil, fmt.Errorf("invalid JSON in --params: %w", err)
		}
	}

	if paramsFile != "" {
		data, err := os.ReadFile(paramsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read params file: %w", err)
		}
		if err := json.Unmarshal(data, &params); err != nil {
			return nil, fmt.Errorf("invalid JSON in params file: %w", err)
		}
	}

	for i := 0; i < len(flagArgs); i++ {
		arg := flagArgs[i]

		if strings.HasPrefix(arg, "--") {
			key := strings.TrimPrefix(arg, "--")
			if i+1 < len(flagArgs) && !strings.HasPrefix(flagArgs[i+1], "-") {
				value := flagArgs[i+1]
				if isJSON(value) {
					var parsed interface{}
					if err := json.Unmarshal([]byte(value), &parsed); err == nil {
						params[key] = parsed
					} else {
						params[key] = value
					}
				} else {
					params[key] = value
				}
				i++
			} else {
				params[key] = "true"
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) == 2 {
			key := strings.TrimPrefix(arg, "-")
			if i+1 < len(flagArgs) && !strings.HasPrefix(flagArgs[i+1], "-") {
				params[key] = flagArgs[i+1]
				i++
			} else {
				params[key] = "true"
			}
		} else if !strings.HasPrefix(arg, "-") {
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) == 2 {
					params[parts[0]] = parts[1]
				}
			}
		}
	}

	return params, nil
}

func isJSON(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")
}

var executeCmd = &cobra.Command{
	Use:   "execute <emblem> <action> [flags]",
	Short: "Execute an emblem action",
	Long: `Execute an action from an installed emblem.

Examples:
  # Execute with flags
  ely execute clothing-shop list-products --category shoes

  # Execute with JSON params
  ely execute clothing-shop list-products --params '{"category": "shoes"}'

  # Execute with params from file
  ely execute clothing-shop create-product --params-file product.json`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		emblemName := args[0]
		actionName := args[1]

		if !isInstalledEmblem(emblemName) {
			return fmt.Errorf("emblem '%s' is not installed. Run: ely pull %s", emblemName, emblemName)
		}

		return executeEmblemAction(emblemName, []string{actionName})
	},
}

func init() {
	executeCmd.Flags().StringVar(&paramsJSON, "params", "", "Parameters as JSON string")
	executeCmd.Flags().StringVar(&paramsFile, "params-file", "", "Load parameters from JSON file")
	rootCmd.AddCommand(executeCmd)
}
