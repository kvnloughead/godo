/*
Copyright © 2024 Kevin Loughead <kvnloughead@gmail.com>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"

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
	cfgFile string
	app     *CLIApplication
)

func init() {
	// First, set up the flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/godo/settings.json)")

	// Then initialize the application
	cobra.OnInitialize(func() {
		logger := logger.NewLogger()
		cliConfig, err := config.LoadConfig(cfgFile, logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		app = &CLIApplication{
			Logger: logger,
			Config: cliConfig,
		}
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type CLIConfig struct {
	APIBaseURL string `json:"api_base_url"`
	// Add other CLI-specific settings here
}

type CLIApplication struct {
	Logger *slog.Logger
	Config config.Config
}

func NewCLIApplication() (*CLIApplication, error) {
	// Get config file path from flag
	cfgFile, err := rootCmd.PersistentFlags().GetString("config")
	if err != nil {
		return nil, err
	}

	logger := logger.NewLogger()

	// Use the config package's LoadConfig function
	cliConfig, err := config.LoadConfig(cfgFile, logger)
	if err != nil {
		return nil, err
	}

	return &CLIApplication{
		Logger: logger,
		Config: cliConfig,
	}, nil
}
