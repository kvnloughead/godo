/*
Copyright © 2024 Kevin Loughead <kvnloughead@gmail.com>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <text>",
	Short: "Add a new todo item with the given text",
	Long: `
Add a new todo item with the given text. Text with spaces must be enclosed in quotes. For example:

    # Add a todo item
    godo add "Buy groceries"

This command requires authentication. Run 'godo auth -h' for more information.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := app.Config.APIBaseURL + "/todos"
		stdoutMsg := "\nError: failed to add todo item. \nCheck `~/.config/godo/logs` for details.\n"

		// handleError captures parameters that are common to all errors
		handleError := func(logMsg string, err error) {
			app.handleError(logMsg, stdoutMsg, err,
				"method", http.MethodPost,
				"url", url)
		}

		token, err := app.ReadTokenFromFile()
		if err != nil {
			app.handleAuthenticationError("Failed to read token", err)
			return
		}

		text := args[0]
		payload := map[string]string{"text": text}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			handleError("Failed to marshal JSON", err)
			return
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			handleError("Failed to create request", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+string(token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			handleError("Failed to send request", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			handleError("Failed to add todo", fmt.Errorf("response status: %s", resp.Status))
			return
		}

		fmt.Println("Todo added successfully")

	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
