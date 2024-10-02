package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rezkam/papilot/openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	providerName = "openai"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the configuration for the provider",
	RunE:  runInitCmd,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInitCmd(cmd *cobra.Command, args []string) error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	viper.SetConfigFile(configPath)

	// create the config directory if it doesn't exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}
	}

	cfg, err := openai.Configure()
	if err != nil {
		return fmt.Errorf("error prompting for provider config: %w", err)
	}

	// save the configs
	viper.Set("provider", providerName)

	// Save the entire configuration using reflection
	viper.Set("config", cfg)

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	// Change the file permissions to 0600 after writing the config
	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("error setting file permissions: %w", err)
	}

	fmt.Println("Provider configuration initialized successfully")
	return nil
}
