package cmd

import (
	"fmt"
	"os"

	"github.com/elysium/elysium/cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "ely",
	Short: "Elysium - The API App Store",
	Long: `Elysium is an API app store that allows you to discover, 
download, and use APIs programmatically through defined emblems.

An emblem is a YAML file that describes an API's endpoints, parameters,
authentication, and types - enabling developers and AI agents to interact
with APIs without reading extensive documentation.`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.elysium/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress output")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "output format (table, json, plain)")

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
