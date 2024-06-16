/*
Copyright © 2024 Kevin Loughead <kvnloughead@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/kvnloughead/godo/internal/injector"
	"github.com/kvnloughead/godo/internal/logger"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gd [command]",
	Short: "gd is a CLI todo tracker",
	Long: "\n" + `gd is a CLI todo tracker application written in Go. It supports todo.txt syntax and is backed by an HTTP server and Postrgresql database.
	
	TODO - add better help text.
	`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

type CLIApplication struct {
	*injector.Application
}

var app *CLIApplication

func NewCLIApplication(app *injector.Application) *CLIApplication {
	return &CLIApplication{Application: app}
}

func init() {
	cfg := injector.LoadConfig()
	logger := logger.NewLogger()

	baseApp := injector.NewApplication(cfg, logger, nil)
	app = NewCLIApplication(baseApp)

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.config/godo/settings.json)")
}
