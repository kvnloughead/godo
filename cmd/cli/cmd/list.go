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

// boolFlags is a list of boolean flags that map to URL query parameters.
var boolFlags = []types.QueryFlag{
	{Flag: "include-archived", Param: "include-archived", Msg: "include archived todos"},
	{Flag: "only-archived", Param: "only-archived", Msg: "only show archived todos"},
	{Flag: "done", Param: "done", Short: "d", Msg: "show only completed todos"},
	{Flag: "undone", Param: "undone", Short: "u", Msg: "show only incomplete todos"},
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

Without the --plain flag, the command enters an interactive mode. When in
interactive mode, the command will prompt for a command and one or more todo
IDs. The command will then be applied to the corresponding todos.

To see the available interactive-mode commands, run 'godo list' and press '?'.

Examples:
    # List unarchived todos in plain text format
    godo list --plain

    # List unarchived todos with @phone in the text in interactive mode
    godo list @phone

    # List all todos (including archived) in interactive mode
    godo list --all

    # List only archived and uncompleted todos
    godo list --only-archived --undone

This command requires authentication. Run 'godo auth -h' for more information.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get boolean flags that map to URL query parameters.
		params := url.Values{}

		// Add boolean query parameters.
		for _, f := range boolFlags {
			if val, _ := cmd.Flags().GetBool(f.Flag); val {
				params.Add(f.Param, "true")
			}
		}

		// Get other flags.
		plain, _ := cmd.Flags().GetBool("plain")

		// Set up interactive commands.
		commands := map[string]*interactive.Command{
			"delete": {
				Name:    "delete",
				Aliases: []string{"rm", "del"},
				Action: func(todoIDs []int) error {
					dummyCmd := &cobra.Command{}
					for _, todoID := range todoIDs {
						deleteCmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
					}
					return nil
				},
			},
			"done": {
				Name:    "done",
				Aliases: []string{"d", "complete"},
				Action: func(todoIDs []int) error {
					dummyCmd := &cobra.Command{}
					for _, todoID := range todoIDs {
						doneCmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
					}
					return nil
				},
			},
			"undone": {
				Name:    "undone",
				Aliases: []string{"u", "undone", "undo", "incomplete"},
				Action: func(todoIDs []int) error {
					dummyCmd := &cobra.Command{}
					for _, todoID := range todoIDs {
						undoneCmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
					}
					return nil
				},
			},
			"archive": {
				Name:    "archive",
				Aliases: []string{"a"},
				Action: func(todoIDs []int) error {
					dummyCmd := &cobra.Command{}
					for _, todoID := range todoIDs {
						archiveCmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
					}
					return nil
				},
			},
			"unarchive": {
				Name:    "unarchive",
				Aliases: []string{"ua"},
				Action: func(todoIDs []int) error {
					dummyCmd := &cobra.Command{}
					for _, todoID := range todoIDs {
						unarchiveCmd.Run(dummyCmd, []string{strconv.Itoa(todoID)})
					}
					return nil
				},
			},
		}
		interactive := interactive.New(commands)

		// Fetch todos and display them. If plain mode is enabled, the loop
		// will exit after the todos are displayed. Otherwise, the loop will
		// continue until the user exits interactive mode.
		for {
			todos, err := fetchTodos(args, params)
			if err != nil {
				return
			}

			// Store the ordered todos for interactive mode
			orderedTodos := displayTodos(todos, plain)

			if plain {
				break
			}

			if err := interactive.Prompt(orderedTodos); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	},
}

// fetchTodos retrieves todos from the API, handling authentication and
// filtering.
func fetchTodos(args []string, params url.Values) ([]types.Todo, error) {
	// Add query parameters to the base URL.
	baseURL := app.Config.APIBaseURL + "/todos"
	if len(args) > 0 {
		searchText := strings.ReplaceAll(args[0], "+", "%2B")
		searchPattern := url.QueryEscape(searchText)
		params.Add("text", searchPattern)
	}
	baseURL = baseURL + "?" + params.Encode()

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
func displayTodos(todos []types.Todo, plain bool) []types.Todo {
	if plain {
		fmt.Println("id\tcompleted\ttext")
		for _, todo := range todos {
			fmt.Printf("%d\t%t\t\t%s\n", todo.ID, todo.Completed, todo.Text)
		}
		return todos
	} else {
		// Split todos into active and archived
		var active, archived []types.Todo
		for _, todo := range todos {
			if todo.Archived {
				archived = append(archived, todo)
			} else {
				active = append(active, todo)
			}
		}

		// Sort each slice by completion status (uncompleted first)
		sortTodos := func(todos []types.Todo) {
			sort.Slice(todos, func(i, j int) bool {
				return !todos[i].Completed
			})
		}
		sortTodos(active)
		sortTodos(archived)

		// Combine the slices in display order
		orderedTodos := append(active, archived...)

		// Display active todos
		displayIndex := 1
		output := func(todos []types.Todo, heading string) {
			if len(todos) > 0 {
				fmt.Println("\n" + heading + ":\n")
				for _, todo := range todos {
					if todo.Completed {
						fmt.Printf("%2d. [\033[90m✓\033[0m] \033[90m%s\033[0m\n", displayIndex, todo.Text)
					} else {
						fmt.Printf("%2d. [ ] %s\n", displayIndex, todo.Text)
					}
					displayIndex++
				}
			}
		}
		output(active, "Todos")
		output(archived, "Archived")

		return orderedTodos
	}
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("plain", "p", false, "output in plain text to stdout")

	// Add boolean flags that map to URL query parameters.
	for _, f := range boolFlags {
		if f.Short != "" {
			listCmd.Flags().BoolP(f.Flag, f.Short, false, f.Msg)
		} else {
			listCmd.Flags().Bool(f.Flag, false, f.Msg)
		}
	}

	// Mark flags as mutually exclusive.
	listCmd.MarkFlagsMutuallyExclusive("only-archived", "include-archived")
	listCmd.MarkFlagsMutuallyExclusive("done", "undone")
}
