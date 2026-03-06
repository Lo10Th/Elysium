package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Version = "0.2.4"

var rootCmd = &cobra.Command{
	Use:                "ely",
	Short:              "Elysium - The API App Store",
	Long:               `Elysium is an API app store that allows you to discover, download, and use APIs programmatically through defined emblems.`,
	Version:            Version,
	DisableFlagParsing: false,
	TraverseChildren:   true,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.elysium/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress output")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "output format (table, json, yaml, csv, plain)")
	rootCmd.PersistentFlags().Bool("no-check", false, "disable update notifications")

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

func initConfig() {
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	cfgFile := viper.GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home + "/.elysium")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("elysium")
	viper.AutomaticEnv()

	viper.ReadInConfig()
}

func Execute() {
	cobra.EnablePrefixMatching = true

	if err := initConfigEarly(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		firstArg := os.Args[1]
		if !strings.HasPrefix(firstArg, "-") && firstArg != "help" && firstArg != "completion" {
			if !isKnownCommand(firstArg) {
				if isInstalledEmblem(firstArg) {
					emblemName := firstArg
					actionArgs := os.Args[2:]

					format := "table"
					for i, arg := range actionArgs {
						switch arg {
						case "-o", "--output":
							if i+1 < len(actionArgs) {
								format = actionArgs[i+1]
							}
						case "--pretty":
							prettyOutput = true
						case "--fields":
							if i+1 < len(actionArgs) {
								outputFields = actionArgs[i+1]
							}
						case "--format":
							if i+1 < len(actionArgs) {
								outputTemplate = actionArgs[i+1]
							}
						case "--width":
							if i+1 < len(actionArgs) {
								if n, err := fmt.Sscanf(actionArgs[i+1], "%d", &tableWidth); n != 1 || err != nil {
									tableWidth = 0
								}
							}
						case "--no-color":
							noColorOutput = true
						}
					}
					outputFormat = format

					if err := executeEmblemAction(emblemName, actionArgs); err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						os.Exit(1)
					}
					return
				}
			}
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isKnownCommand(cmd string) bool {
	commands := []string{
		"check-updates", "execute", "help", "completion", "init", "info", "keys", "list", "login", "logout", "outdated", "pull", "search", "self-update", "test", "update", "validate", "whoami",
	}
	for _, c := range commands {
		if c == cmd {
			return true
		}
	}
	return false
}

func initConfigEarly() error {
	if err := config.Init(); err != nil {
		return err
	}

	cfgFile := viper.GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		viper.AddConfigPath(home + "/.elysium")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("elysium")
	viper.AutomaticEnv()

	viper.ReadInConfig()
	return nil
}
