package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/kvnloughead/godo/cmd/cli/interactive"
	"github.com/kvnloughead/godo/cmd/cli/types"
	"github.com/spf13/cobra"
)

// interactiveCmd creates an interactive command struct. It accepts a command
// name, aliases, and the actual command to run. The command to be executed is
// wrapped in a dummy command so that it can be executed in interactive mode.
//
// Interactive commands are logged to the log file, but not until the interactive
// interactive mode is exited.
func interactiveCmd(cmdName string, aliases []string, cmd *cobra.Command) *interactive.Command {
	return &interactive.Command{
		Name:    cmdName,
		Aliases: aliases,
		Action: func(todoID int) error {
			app.Logger.Info("executing interactive command",
				"command", cmdName,
				"todo_id", todoID)
			dummyCmd := &cobra.Command{}
			cmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
			return nil
		},
	}
}

// commands is a map of interactive commands for managing todos.
var commands = map[string]*interactive.Command{
	"delete":    interactiveCmd("delete", []string{"rm", "del"}, deleteCmd),
	"done":      interactiveCmd("done", []string{"d", "complete"}, doneCmd),
	"undone":    interactiveCmd("undone", []string{"u", "undo", "incomplete"}, undoneCmd),
	"archive":   interactiveCmd("archive", []string{"a"}, archiveCmd),
	"unarchive": interactiveCmd("unarchive", []string{"ua"}, unarchiveCmd),
}

// listCmd displays todos and enters an interactive mode for managing them.
// Results can be filtered by a plain text search pattern.
var listCmd = &cobra.Command{
	Use:   "list [pattern]",
	Short: "List and manage todo items",
	Long: `List and manage todo items for the authenticated user. Items can be filtered by a plain text search pattern. If the pattern contains multiple words it must be 
enclosed in quotes.

The command enters an interactive mode where you can manage todos using these commands:
  <number>rm|del|delete  Delete the selected todo
  <number>d|done        Mark the selected todo as done
  <number>u|undone      Mark the selected todo as not done
  <number>e|edit        Edit the selected todo's text
  <number>a|archive     Archive the selected todo
	<number>ua|unarchive  Unarchive the selected todo

Other commands:
  ?        Show help
  q        Exit interactive mode

Examples:
    # List all todos
    godo list

    # List all todos with @phone in the text
    godo list @phone

    # In interactive mode:
    1rm     # Delete todo #1
    2d      # Mark todo #2 as done
    3u      # Mark todo #3 as not done

This command requires authentication. Run 'godo auth -h' for more information.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		commands := map[string]*interactive.Command{
			"delete": {
				Name:    "delete",
				Aliases: []string{"rm", "del"},
				Action: func(todoID int) error {
					dummyCmd := &cobra.Command{}
					deleteCmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
					return nil
				},
			},
			"done": {
				Name:    "done",
				Aliases: []string{"d", "complete"},
				Action: func(todoID int) error {
					dummyCmd := &cobra.Command{}
					doneCmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
					return nil
				},
			},
			"undone": {
				Name:    "undone",
				Aliases: []string{"u", "undone", "undo", "incomplete"},
				Action: func(todoID int) error {
					dummyCmd := &cobra.Command{}
					undoneCmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
					return nil
				},
			},
			"archive": {
				Name:    "archive",
				Aliases: []string{"a"},
				Action: func(todoID int) error {
					dummyCmd := &cobra.Command{}
					archiveCmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
					return nil
				},
			},
			"unarchive": {
				Name:    "unarchive",
				Aliases: []string{"ua"},
				Action: func(todoID int) error {
					dummyCmd := &cobra.Command{}
					unarchiveCmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
					return nil
				},
			},
		}
		interactive := interactive.New(commands)

		for {
			todos, err := fetchTodos(args)
			if err != nil {
				return
			}

			displayTodos(todos)

			if err := interactive.Prompt(todos); err != nil {
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

	req, err := app.createJSONRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, handleError("Failed to create request", err)
	}
	req.Header.Set("Authorization", "Bearer "+string(token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, handleError("Failed to send request", err)
	}
	defer resp.Body.Close()

	// Read response body and log it
	body, err := app.readTodoListResponse(resp, handleError)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, handleError("Failed to retrieve todos", fmt.Errorf("response status: %s", resp.Status))
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

	// Sort todos by archived status (unarchived first) and completion status
	// (incomplete first). This prevents gaps in the displayed todo numbers.
	sort.Slice(todos, func(i, j int) bool {
		if todos[i].Archived != todos[j].Archived {
			return !todos[i].Archived
		}
		return !todos[i].Completed
	})

	for i, todo := range todos {
		if todo.Archived {
			continue
		}
		if todo.Completed {
			fmt.Printf("%2d. [\033[90m✓\033[0m] \033[90m%s\033[0m\n",
				i+1, todo.Text)
		} else {
			fmt.Printf("%2d. [ ] %s\n", i+1, todo.Text)
		}
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
