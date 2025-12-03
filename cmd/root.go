/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	chorus "github.com/standrze/chorus/pkg/agent"

	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chorus",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		client := openai.NewClient(
			option.WithBaseURL("http://localhost:12434/engines/llama.cpp/v1"),
		)

		ctx := context.Background()

		agent := chorus.NewAgent(&client, chorus.WithName("agent"), chorus.WithModel("ai/gpt-oss"), chorus.WithReasoningEffort(openai.ReasoningEffortMedium))
		agent2 := chorus.NewAgent(&client, chorus.WithName("agent2"), chorus.WithModel("ai/gpt-oss"), chorus.WithReasoningEffort(openai.ReasoningEffortMedium))

		agent.SystemMessage(`
			You are a helpful assistant.
		`)
		agent2.SystemMessage(`
			You are a helpful assistant. 
		`)

		conversation := chorus.NewConversation(ctx, agent, agent2)
		conversation.Start("Ask a question to the user and wait for the answer.")

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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chorus.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
