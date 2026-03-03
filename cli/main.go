package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "ely",
		Short: "Elysium - The API App Store",
		Long: `Elysium is an API app store that allows you to discover, 
download, and use APIs programmatically through defined emblems.

An emblem is a YAML file that describes an API's endpoints, parameters,
authentication, and types - enabling developers and AI agents to interact
with APIs without reading extensive documentation.`,
		Version: "1.0.0",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.elysium/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress output")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "output format (table, json, plain)")
}

func initConfig() {
	// Config initialization will be handled in internal/config
}
