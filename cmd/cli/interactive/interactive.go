package interactive

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kvnloughead/godo/cmd/cli/config"
	"github.com/kvnloughead/godo/cmd/cli/types"
	"github.com/spf13/cobra"
)

// Mode represents an interactive session for managing todos.
type Mode struct {
	config   *config.Config
	commands *Commands
}

// Commands holds references to the cobra commands.
type Commands struct {
	Delete *cobra.Command
}

// New creates a new interactive mode with the given config and commands.
func New(config *config.Config, commands *Commands) *Mode {
	return &Mode{
		config:   config,
		commands: commands,
	}
}

func (m *Mode) Prompt(todos []types.Todo) error {
	fmt.Print("Enter command (? for help): ")
	var input string
	fmt.Scanln(&input)

	if input == "q" || input == "quit" {
		return nil
	}

	if input == "?" || input == "help" {
		m.ShowHelp()
		return nil
	}

	return m.ExecuteCommand(input, todos)
}

// ExecuteCommand processes the user's input and executes the corresponding
// action. It handles both shorthand commands and long-form commands.
//
// Available commands are described in ShowHelp().
func (m *Mode) ExecuteCommand(input string, todos []types.Todo) error {
	// If the input is a number that is less than 1 or greater than the
	// number of todos, return an error. Otherwise, if the input is a
	// shorthand command, execute it.
	if num, cmd, ok := m.parseShorthand(input); ok {
		if num < 1 || num > len(todos) {
			return fmt.Errorf("invalid todo number: %d", num)
		}
		if cmd != "" {
			return m.executeActionOnTodo(num, cmd, todos)
		}
	}

	// If the input is a long-form command, attempt to execute it.

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "d", "done":
		return m.promptAndExecute("done", todos)
	case "d-", "undone":
		return m.promptAndExecute("undone", todos)
	case "rm", "delete":
		return m.promptAndExecute("delete", todos)
	case "a", "archive":
		return m.promptAndExecute("archive", todos)
	case "u", "update":
		return m.promptAndExecute("update", todos)
	default:
		return fmt.Errorf("unknown command: %s", parts[0])
	}
}

// parseShorthand parses input strings like "3d" or "2rm" into a todo number
// and command. Returns the todo number (1-based), command string, and whether
// parsing was successful. Valid commands are: d (done), d- (undone), rm
// (delete), a (archive), u (update)
func (m *Mode) parseShorthand(input string) (int, string, bool) {
	// If input is a number without a command, return it
	re := regexp.MustCompile(`^\d+$`)
	matches := re.FindStringSubmatch(input)
	if matches != nil {
		num, _ := strconv.Atoi(matches[0])
		return num, "", true
	}

	// Check if input is a number followed by a command
	re = regexp.MustCompile(`^(\d+)(d|d-|rm|a|u|done|undone|delete|archive|update)$`)
	matches = re.FindStringSubmatch(input)
	if matches == nil {
		return 0, "", false
	}

	num, _ := strconv.Atoi(matches[1])
	cmd := matches[2]
	return num, cmd, true
}

// executeActionOnTodo executes the specified command on a todo item.
// num is 1-based index in the displayed list.
// cmd must be one of: "d", "d-", "rm", "a", "u"
// Returns an error if the action fails or the todo number is invalid.
func (m *Mode) executeActionOnTodo(num int, cmd string, todos []types.Todo) error {
	if num < 1 || num > len(todos) {
		return fmt.Errorf("invalid todo number: %d", num)
	}
	todo := todos[num-1]

	// Create dummy cobra command for reuse of existing commands
	dummyCmd := &cobra.Command{}
	idStr := strconv.Itoa(todo.ID)

	switch cmd {
	// case "d":
	// 	doneCmd.Run(dummyCmd, []string{idStr})
	// case "d-":
	// 	undoneCmd.Run(dummyCmd, []string{idStr})
	case "rm":
		m.commands.Delete.Run(dummyCmd, []string{idStr})
	// case "a":
	// 	archiveCmd.Run(dummyCmd, []string{idStr})
	// case "u":
	// 	updateCmd.Run(dummyCmd, []string{idStr})
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}

	return nil
}

// promptAndExecute prompts the user for a todo number and executes the
// specified action. action must be one of the following:
//
//	"(d)one", "(d-)undone", "(rm)delete", "(a)rchive", "(u)pdate"
//
// Returns an error if the action fails or the todo number is invalid.
func (m *Mode) promptAndExecute(action string, todos []types.Todo) error {
	fmt.Print("Enter todo number: ")
	var input string
	fmt.Scanln(&input)

	num, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("invalid input: must be a number")
	}

	// Convert long-form commands to their shorthand equivalents
	cmdMap := map[string]string{
		"done":    "d",
		"undone":  "d-",
		"delete":  "rm",
		"archive": "a",
		"update":  "u",
	}

	cmd, ok := cmdMap[action]
	if !ok {
		return fmt.Errorf("unknown action: %s", action)
	}

	return m.executeActionOnTodo(num, cmd, todos)
}

// ShowHelp displays the available commands in interactive mode.
func (m *Mode) ShowHelp() {
	fmt.Println("\nShorthandCommands:")
	fmt.Println("  <num>d   Mark todo as done (shorthand)")
	fmt.Println("  <num>d-  Mark todo as not done (shorthand)")
	fmt.Println("  <num>rm  Delete todo (shorthand)")
	fmt.Println("  <num>a   Archive todo (shorthand)")
	fmt.Println("  <num>u   Update todo (shorthand)")
	fmt.Println("\nLong-form Commands):")
	fmt.Println("  done     Mark todo as done (interactive)")
	fmt.Println("  undone   Mark todo as not done (interactive)")
	fmt.Println("  delete   Delete todo (interactive)")
	fmt.Println("  archive  Archive todo (interactive)")
	fmt.Println("  update   Update todo text (interactive)")
	fmt.Println("\nOther Commands:")
	fmt.Println("  ?        Show this help")
	fmt.Println("  q        Quit")
}
