/*
Copyright © 2024 Kevin Loughead <kvnloughead@gmail.com>
*/
package cmd

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a todo item by its ID",
	Long: `
Delete a todo item by its ID. The ID can be found in the leftmost column when
listing todos.

Examples:

    # Delete todo with ID 123
    godo delete 123

This command requires authentication. Run 'godo auth -h' for more information
about authentication.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil || id < 1 {
			fmt.Println("Error: ID must be a positive integer")
			return
		}

		url := fmt.Sprintf("%s/todos/%d", app.Config.APIBaseURL, id)
		stdoutMsg := "\nError: failed to delete todo item. \nCheck `~/.config/godo/logs` for details.\n"

		// handleError captures parameters that are common to all errors
		handleError := func(logMsg string, err error) {
			app.handleError(logMsg, stdoutMsg, err,
				"method", http.MethodDelete,
				"url", url)
		}

		token, err := app.ReadTokenFromFile()
		if err != nil {
			app.handleAuthenticationError("Failed to read token", err)
			return
		}

		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			handleError("Failed to create request", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+string(token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			handleError("Failed to send request", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			switch resp.StatusCode {
			case http.StatusNotFound:
				fmt.Println("Error: todo not found")
			default:
				handleError("Failed to delete todo", fmt.Errorf("response status: %s", resp.Status))
			}
			return
		}

		fmt.Println("Todo deleted successfully")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
