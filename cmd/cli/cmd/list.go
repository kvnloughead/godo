package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kvnloughead/godo/cmd/cli/interactive"
	"github.com/kvnloughead/godo/cmd/cli/types"
	"github.com/spf13/cobra"
)

type PaginationData struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	FirstPage    int `json:"first_page"`
	LastPage     int `json:"last_page"`
	TotalRecords int `json:"total_records"`
}

// type Todo struct {
// 	ID        int    `json:"id"`
// 	UserID    int    `json:"user_id"`
// 	CreatedAt string `json:"created_at"`
// 	Text      string `json:"text"`
// 	Priority  string `json:"priority"`
// 	Completed bool   `json:"completed"`
// 	Version   int    `json:"version"`
// }

// type TodoResponse struct {
// 	PaginationData PaginationData `json:"paginationData"`
// 	Todos          []Todo         `json:"todos"`
// }

// listCmd displays todos and enters an interactive mode for managing them.
// Results can be filtered by a plain text search pattern.
var listCmd = &cobra.Command{
	Use:   "list [pattern]",
	Short: "List and manage todo items",
	Long: `
List and manage todo items for the authenticated user. Items can be filtered by 
a plain text search pattern. If the pattern contains multiple words it must be 
enclosed in quotes.

Examples:
    # List all todos
    godo list

    # List all todos with @phone in the text
    godo list @phone

This command requires authentication. Run 'godo auth -h' for more information.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		commands := &interactive.Commands{
			Delete: DeleteCmd,
		}
		interactive := interactive.New(&app.Config, commands)

		for {
			todos, err := fetchTodos(args)
			if err != nil {
				return
			}

			displayTodos(todos)

			fmt.Print("\nEnter command (? for help): ")
			var input string
			fmt.Scanln(&input)

			if input == "q" || input == "quit" {
				return
			}

			if input == "?" || input == "help" {
				interactive.ShowHelp()
				continue
			}

			if err := interactive.ExecuteCommand(input, todos); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	},
}

// fetchTodos retrieves todos from the API, handling authentication and filtering
func fetchTodos(args []string) ([]types.Todo, error) {
	baseURL := app.Config.APIBaseURL + "/todos"
	if len(args) > 0 {
		searchText := strings.ReplaceAll(args[0], "+", "%2B")
		searchPattern := url.QueryEscape(searchText)
		baseURL = baseURL + "?text=" + searchPattern
	}

	stdoutMsg := "\nError: failed to list todo items. \nCheck `~/.config/godo/logs` for details.\n"

	// handleError captures parameters that are common to all errors
	handleError := func(logMsg string, err error) error {
		app.handleError(logMsg, stdoutMsg, err,
			"method", http.MethodGet,
			"url", baseURL)
		return err
	}

	token, err := app.TokenManager.LoadToken()
	if err != nil {
		app.handleAuthenticationError("Failed to read token", err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, handleError("Failed to create request", err)
	}
	req.Header.Set("Authorization", "Bearer "+string(token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, handleError("Failed to send request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleError("Failed to retrieve todos", fmt.Errorf("response status: %s", resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, handleError("Failed to read body", err)
	}

	var todoResponse types.TodoResponse
	if err := json.Unmarshal(body, &todoResponse); err != nil {
		return nil, handleError("Failed to unmarshal JSON", err)
	}

	if len(todoResponse.Todos) == 0 {
		fmt.Println("No matches found.")
		return nil, nil
	}

	return todoResponse.Todos, nil
}

func displayTodos(todos []types.Todo) {
	fmt.Println("\nTodos:")
	for i, todo := range todos {
		status := " "
		if todo.Completed {
			status = "✓"
		}
		fmt.Printf("%2d. [%s] %s\n", i+1, status, todo.Text)
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
