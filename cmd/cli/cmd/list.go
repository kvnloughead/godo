/*
Copyright © 2024 Kevin Loughead <kvnloughead@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
	Title     string `json:"title"`
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
	Short: "A brief description of your command",
	Long:  `TODO - add long help text`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var pattern string
		if len(args) != 0 {
			pattern = args[0]
		}

		token, err := app.ReadTokenFromFile()
		if err != nil {
			app.handleAuthenticationError("Failed to send request", err)
			return
		}

		msg := "\nError: failed to list todo items. \nCheck `~/.config/godo/logs` for details.\n"

		url := app.Config.APIBaseURL + "/v1/todos"
		if pattern != "" {
			url = url + "?title=" + pattern
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			app.handleError("Failed to create request", msg, err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+string(token))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			app.handleError("Failed to send request", msg, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			app.handleError("Failed to retrieve todos", msg, fmt.Errorf("response status: %s", resp.Status))
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			app.handleError("Failed to read body", msg, err)
			return
		}

		var todoResponse TodoResponse
		if err := json.Unmarshal(body, &todoResponse); err != nil {
			app.handleError("Failed to unmarshal JSON", msg, err)
			return
		}

		if len(todoResponse.Todos) == 0 {
			fmt.Println("No matches found.")
			return
		}

		for _, todo := range todoResponse.Todos {
			fmt.Println(todo.Title)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
