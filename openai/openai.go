package openai

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/spf13/viper"
	"golang.org/x/term"
)

const (
	model     = "gpt-4o-mini"
	url       = "https://api.openai.com/v1/chat/completions"
	maxTokens = 500
)

type Config struct {
	APIKey string `mapstructure:"apiKey"`
}

func Configure() (any, error) {
	fmt.Print("Enter your OpenAI API Key: ")

	// Read password input from the terminal without echoing
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("error reading API Key: %w", err)
	}

	// Convert the byte slice to a string and trim whitespace
	apiKey := strings.TrimSpace(string(bytePassword))

	// Print a newline to move the cursor to the next line
	fmt.Println()

	if len(apiKey) == 0 {
		return nil, fmt.Errorf("API Key cannot be empty")
	}

	c := Config{}

	c.APIKey = apiKey
	return c, nil
}

type Provider struct {
	cfg Config
}

// NewProvider creates a new OpenAI provider using the configuration file
func NewProvider(filepath string) (*Provider, error) {
	if filepath == "" {
		return nil, fmt.Errorf("filepath cannot be empty")
	}

	// Set the configuration file path and read the config
	viper.SetConfigFile(filepath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	// Unmarshal the nested 'config' key into the cfg struct
	if err := viper.UnmarshalKey("config", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Check if the API key is provided
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is missing in the config file")
	}

	return &Provider{cfg: cfg}, nil
}

type response struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message message `json:"message"`
}

type message struct {
	Content string `json:"content"`
}
