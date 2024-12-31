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

// listCmd displays todos and can be filtered by a plain text search pattern.
// By default, the command enters an interactive mode. With the --plain flag
// set, the todos are output in plain text.
var listCmd = &cobra.Command{
	Use:   "list [--all|--archived|--unarchived|--done|--undone|--plain] [pattern]",
	Short: "List and manage todo items",
	Long: `List and manage todo items for the authenticated user.

Items can be filtered by a plain text search pattern. If the pattern contains multiple words it must be enclosed in quotes.

The --plain flag outputs the todos in plain text format suitable for scripts and piping to other commands. The output has the following columns:

  - id: the todo ID
  - completed: the todo completion status
  - text: the todo text

Without the --plain flag, the command enters an interactive mode. To see the
available commands, run 'godo list' and press '?'.

Examples:
    # List unarchived todos in plain text format
    godo list --plain

    # List unarchived todos with @phone in the text in interactive mode
    godo list @phone

    # List all todos (including archived) in interactive mode
    godo list --all

		# List only archived and uncompleted todos
		godo list --archived --undone

This command requires authentication. Run 'godo auth -h' for more information.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags.
		plain, _ := cmd.Flags().GetBool("plain")

		// Set up interactive commands.
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

		// Fetch todos and display them. If plain mode is enabled, the loop
		// will exit after the todos are displayed. Otherwise, the loop will
		// continue until the user exits interactive mode.
		for {
			todos, err := fetchTodos(args)
			if err != nil {
				return
			}

			displayTodos(todos, plain)

			if plain {
				break
			}

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

// displayTodos outputs todos in either plain text or interactive mode. In plain
// text mode, the output is suitable for scripts and piping to other commands.
// It has the following columns:
//
//   - id: the todo ID
//   - completed: the todo completion status
//   - text: the todo text
//
// In interactive mode, the output is formatted for use with the interactive //
// package.
func displayTodos(todos []types.Todo, plain bool) {
	// Sort todos by archived status (unarchived first) and completion status
	// (incomplete first). This prevents gaps in the displayed todo numbers.
	sort.Slice(todos, func(i, j int) bool {
		if todos[i].Archived != todos[j].Archived {
			return !todos[i].Archived
		}
		return !todos[i].Completed
	})

	// Output in plain text mode
	if plain {
		fmt.Println("id\tcompleted\ttext")
		for _, todo := range todos {
			if todo.Archived {
				continue
			}
			fmt.Printf("%d\t%t\t\t%s\n", todo.ID, todo.Completed, todo.Text)
		}
		// Output in interactive mode
	} else {
		fmt.Println("\nTodos:")
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
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("plain", "p", false, "output in plain text to stdout")
}
