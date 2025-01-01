// Package cmd implements the CLI commands for the godo application.
// It provides commands for user management (register, activate, auth)
// and todo management (add, list, and delete).
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/kvnloughead/godo/cmd/cli/config"
	"github.com/kvnloughead/godo/cmd/cli/token"
	"github.com/kvnloughead/godo/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// rootCmd is the base command for the godo CLI. It provides a way to manage
// todos and user accounts.
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
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile,
		"config",
		"c",
		"",
		"config file (default is $HOME/.config/godo/settings.json)",
	)

	// Log the command, its arguments, and all flags and their values
	// (excluding password).
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		flags := make(map[string]string)
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name != "password" && f.Value.String() != "" {
				flags[f.Name] = f.Value.String()
			}
		})
		app.Logger.Info("executing command",
			"command", cmd.Name(),
			"args", args,
			"flags", flags)
		return nil
	}

	// Then initialize the application
	cobra.OnInitialize(func() {
		logger := logger.NewLogger()
		cliConfig, err := config.LoadConfig(cfgFile, logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		app = &CLIApplication{
			Logger:       logger,
			Config:       cliConfig,
			TokenManager: token.NewManager(filepath.Join(os.Getenv("HOME"), ".config/godo"), cliConfig.APIBaseURL),
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
	Logger       *slog.Logger
	Config       config.Config
	TokenManager *token.Manager
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
		Logger:       logger,
		Config:       cliConfig,
		TokenManager: token.NewManager(cliConfig.APIBaseURL, cliConfig.APIBaseURL),
	}, nil
}
