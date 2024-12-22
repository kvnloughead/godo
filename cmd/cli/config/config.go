package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	defaultConfigFile = "settings.json"
	defaultAPIBaseURL = "http://godo.kevinloughead.com/v1"
)

type Config struct {
	APIBaseURL string `json:"api_base_url"`
}

// LoadConfig loads the configuration file for the CLI. The config file is
// loaded in this order:
//  1. If a specific config file is provided as a flag or env var, use it.
//  2. If no specific config file is provided, load ~/.config/godo/settings.json
//  3. If no config file is found, use the default configuration.
func LoadConfig(cfgFile string, logger *slog.Logger) (Config, error) {
	// Default configuration
	config := Config{
		APIBaseURL: defaultAPIBaseURL,
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "godo")

	// If no specific config file is provided, use default
	if cfgFile == "" {
		cfgFile = filepath.Join(configDir, defaultConfigFile)
	}

	// Ensure config file exists (creates it with defaults if it doesn't)
	if err := EnsureConfigFile(cfgFile); err != nil {
		logger.Error("error ensuring config file exists", "error", err)
		return config, err
	}

	// Read the config file
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		logger.Error("error reading config file", "error", err)
		return config, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		logger.Error("error parsing config file", "error", err)
		return config, err
	}

	// Override with environment variables
	if url := os.Getenv("GODO_API_URL"); url != "" {
		config.APIBaseURL = url
	}

	return config, nil
}

// EnsureConfigFile checks if the config file exists. If it doesn't, it creates
// it with the default configuration.
func EnsureConfigFile(cfgFile string) error {
	// Check if config file exists
	if _, err := os.Stat(cfgFile); err == nil {
		return nil // Config exists, don't recreate
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(cfgFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Create default settings.json
	defaultConfig := Config{
		APIBaseURL: defaultAPIBaseURL,
	}
	data, err := json.MarshalIndent(defaultConfig, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgFile, data, 0644)
}
