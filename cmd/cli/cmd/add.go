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
	Use:   "add [text]",
	Short: "Add a new todo item with the given text.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]

		token, err := app.ReadTokenFromFile()
		if err != nil {
			app.Logger.Error("Failed to read token from file", "error", err)
			return
		}

		payload := map[string]string{"title": title}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			app.Logger.Error("Failed to marshal JSON", err)
			return
		}

		url := app.Config.APIBaseURL + "/v1/todos"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			app.Logger.Error("Failed to create request", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+string(token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			app.Logger.Error("Failed to send request", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			app.Logger.Error(fmt.Sprintf("Failed to add todo: %s", resp.Status))
			return
		}

		fmt.Println("Todo added successfully")

	},
}

func init() {

	rootCmd.AddCommand(addCmd)
}
