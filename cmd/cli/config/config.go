package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIBaseURL string `json:"api_base_url"`
	// Add other CLI-specific settings here as needed
}

func LoadConfig(cfgFile string) (Config, error) {
	// Default configuration
	config := Config{
		APIBaseURL: "http://localhost:4000/v1",
	}

	// 1. Load from config file
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "godo")
	if cfgFile == "" {
		cfgFile = filepath.Join(configDir, "settings.json")
	}

	if _, err := os.Stat(cfgFile); err == nil {
		data, err := os.ReadFile(cfgFile)
		if err != nil {
			return config, err
		}
		if err := json.Unmarshal(data, &config); err != nil {
			return config, err
		}
	}

	// 2. Override with environment variables
	if url := os.Getenv("GODO_API_URL"); url != "" {
		config.APIBaseURL = url
	}

	return config, nil
}

func EnsureConfigFile(cfgFile string) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "godo")
	if cfgFile == "" {
		cfgFile = filepath.Join(configDir, "settings.json")
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// If config file doesn't exist, create it with default values
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		defaultConfig := Config{
			APIBaseURL: "http://localhost:4000/v1",
		}

		data, err := json.MarshalIndent(defaultConfig, "", "    ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(cfgFile, data, 0644); err != nil {
			return err
		}
	}

	return nil
}
