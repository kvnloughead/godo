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

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <text>",
	Short: "Add a new todo item with the given text.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]

		token, err := app.ReadTokenFromFile()
		if err != nil {
			app.handleAuthenticationError("Failed to send request", err)
			return
		}

		msg := "\nError: failed to add todo item. \nCheck `~/.config/godo/logs` for details.\n"

		payload := map[string]string{"title": title}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			app.handleError("Failed to marshal JSON", msg, err)
			return
		}

		url := app.Config.APIBaseURL + "/v1/todos"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			app.handleError("Failed to create request", msg, err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+string(token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			app.handleError("Failed to send request", msg, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			app.handleError("Failed to add todo", msg, fmt.Errorf("response status: %s", resp.Status))
			return
		}

		fmt.Println("Todo added successfully")

	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
