/*
Copyright © 2024 Kevin Loughead <kvnloughead@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

type PaginationData struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	FirstPage    int `json:"first_page"`
	LastPage     int `json:"last_page"`
	TotalRecords int `json:"total_records"`
}

type Todo struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	CreatedAt string `json:"created_at"`
	Text      string `json:"text"`
	Priority  string `json:"priority"`
	Completed bool   `json:"completed"`
	Version   int    `json:"version"`
}

type TodoResponse struct {
	PaginationData PaginationData `json:"paginationData"`
	Todos          []Todo         `json:"todos"`
}

// listCmd lists a users todos, filtered by an optional pattern argument. By
// default, the todos are listed by text only. The --verbose flag will display
// the entire documents.
var listCmd = &cobra.Command{
	Use:   "list [pattern]",
	Short: "List all todo items",
	Long: `
List all todo items for the authenticated user. Items are displayed in
chronological order and can be filtered by a plain text search pattern. If the
pattern contains multiple words it must be enclosed in quotes.	

Examples:

    # List all todos
    godo list


    # List all todos with @phone in the text
    godo list @phone

This command requires authentication. Run 'godo auth -h' for more information
about authentication.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		baseURL := app.Config.APIBaseURL + "/todos"
		if len(args) > 0 {
			searchText := strings.ReplaceAll(args[0], "+", "%2B")
			searchPattern := url.QueryEscape(searchText)
			baseURL = baseURL + "?text=" + searchPattern
		}

		stdoutMsg := "\nError: failed to list todo items. \nCheck `~/.config/godo/logs` for details.\n"

		// handleError captures parameters that are common to all errors
		handleError := func(logMsg string, err error) {
			app.handleError(logMsg, stdoutMsg, err,
				"method", http.MethodGet,
				"url", baseURL)
		}

		token, err := app.ReadTokenFromFile()
		if err != nil {
			app.handleAuthenticationError("Failed to read token", err)
			return
		}

		req, err := http.NewRequest(http.MethodGet, baseURL, nil)
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
			handleError("Failed to retrieve todos", fmt.Errorf("response status: %s", resp.Status))
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			handleError("Failed to read body", err)
			return
		}

		var todoResponse TodoResponse
		if err := json.Unmarshal(body, &todoResponse); err != nil {
			handleError("Failed to unmarshal JSON", err)
			return
		}

		if len(todoResponse.Todos) == 0 {
			fmt.Println("No matches found.")
			return
		}

		for _, todo := range todoResponse.Todos {
			fmt.Println(todo.Text)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
