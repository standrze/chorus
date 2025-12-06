/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	app "github.com/standrze/chorus/internal"
	clog "github.com/standrze/chorus/pkg/log"
)

// rootCmd represents the base command when called without any subcommands
var cfgFile string
var cfg app.Config
var debug bool

var rootCmd = &cobra.Command{
	Use:   "chorus",
	Short: "Chorus AI Agent Orchestrator",
	Long: `Chorus is a powerful AI agent orchestration platform that allows you 
to create, manage, and coordinate multiple AI agents to solve complex tasks.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		clog.SetDebug(debug)
		cfg.Debug = debug
		err := app.Start(&cfg)
		if err != nil {
			clog.Error("Failed to start app", "error", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.json)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory with name "config" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("json")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// Only log error if config file was explicitly set, otherwise it might be optional?
		// But in this app logic, config seems required for Agents. Let's warn.
		// log.Println("No config file found.")
	} else {
		if debug {
			log.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}
}
