/*
Copyright © 2024 Kevin Loughead <kvnloughead@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/kvnloughead/godo/cmd/cli/config"
	"github.com/kvnloughead/godo/internal/logger"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "godo [command]",
		Short: "godo is a CLI todo tracker",
		Long: "\n" + `godo is a CLI todo tracker application written in Go. It supports todo.txt syntax and is backed by an HTTP server and Postrgresql database.
	`,
	}
	app *CLIApplication
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

type CLIConfig struct {
	APIBaseURL string `json:"api_base_url"`
	// Add other CLI-specific settings here
}

type CLIApplication struct {
	Logger *slog.Logger
	Config CLIConfig
}

func loadCLIConfig(cfgFile string) (CLIConfig, error) {
	// Default configuration
	config := CLIConfig{
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

func NewCLIApplication() (*CLIApplication, error) {
	// Get config file path from flag
	cfgFile, err := rootCmd.PersistentFlags().GetString("config")
	if err != nil {
		return nil, err
	}

	// Load CLI config
	cliConfig, err := loadCLIConfig(cfgFile)
	if err != nil {
		return nil, err
	}

	return &CLIApplication{
		Logger: logger.NewLogger(),
		Config: cliConfig,
	}, nil
}

func init() {
	// Add flags first
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.config/godo/settings.json)")

	// Create new CLI application
	var err error
	app, err = NewCLIApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		os.Exit(1)
	}

	// Ensure config file exists
	cfgFile, err := rootCmd.PersistentFlags().GetString("config")
	if err != nil {
		app.Logger.Error("failed to get config file flag", "error", err)
		os.Exit(1)
	}

	if err := config.EnsureConfigFile(cfgFile); err != nil {
		app.Logger.Error("failed to ensure config file exists", "error", err)
		os.Exit(1)
	}
}
