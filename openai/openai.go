package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/term"
)

const (
	model = "gpt-4o-mini"
	//url                = "https://api.openai.com/v1/chat/completions"
	url                = "http://127.0.0.1:1234/v1/chat/completions"
	maxTokens          = 4000
	swaggerDocPath     = "./papiswaggerdoc.json"
	promptTemplatePath = "./prompt_template.txt"
	httpCallTimeout    = 360 * time.Second
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

func generatePromptFromTemplate(userCommand string) (string, error) {
	swaggerJSON, err := os.ReadFile(swaggerDocPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Swagger JSON file: %w", err)
	}

	templateContent, err := os.ReadFile(promptTemplatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New("prompt").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		SwaggerJSON string
		UserCommand string
	}{
		SwaggerJSON: string(swaggerJSON),
		UserCommand: userCommand,
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

func (p *Provider) GenerateCurlCommand(userCommand string) (string, error) {
	prompt, err := generatePromptFromTemplate(userCommand)
	if err != nil {
		return "", fmt.Errorf("failed to generate prompt: %w", err)
	}
	reqPayload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are an intelligent assistant tasked with generating curl commands and not code for me to run based on both a list of endpoints and user instructions. Only generate the curl command nothing else."},
			{"role": "user", "content": prompt},
		},
		"max_tokens": maxTokens,
		"n":          1,
	}

	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to encode request payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	defer req.Body.Close()

	//req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.cfg.APIKey))
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: httpCallTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorMessage := fmt.Sprintf("api request failed with status code: %d and error %s", resp.StatusCode, string(body))
		return "", errors.New(errorMessage)
	}

	var respPayload response
	err = json.NewDecoder(resp.Body).Decode(&respPayload)
	if err != nil {
		return "", fmt.Errorf("failed to decode response payload: %w", err)
	}

	// Extract the commit message from the response
	if len(respPayload.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return respPayload.Choices[0].Message.Content, nil
}
