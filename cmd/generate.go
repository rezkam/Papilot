package cmd

import (
	"fmt"

	"github.com/rezkam/papilot/openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a curl command from a user requested API calls",
	RunE:  runGenerateCmd,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func runGenerateCmd(cmd *cobra.Command, args []string) error {
	// make sure user provider the command to run
	if len(args) < 1 {
		return fmt.Errorf("usage: papilot generate <command>")
	}
	configPath := getConfigPath()

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	provider, err := openai.NewProvider(configPath)
	if err != nil {
		return fmt.Errorf("error creating provider: %w", err)
	}

	// Generate the curl command
	curlCommand, err := provider.GenerateCurlCommand(args[0])
	if err != nil {
		return fmt.Errorf("error generating curl command: %w", err)
	}

	fmt.Println(curlCommand)

	return nil
}
