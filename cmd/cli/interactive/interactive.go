// Package interactive provides an interactive command-line interface for
// managing items through commands and their aliases. It supports both shorthand
// (e.g., "1rm") and longform command entry with command prompting.
package interactive

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/kvnloughead/godo/cmd/cli/types"
)

// Command represents an action that can be performed on an item. Each command
// hasa primary name, optional aliases, and an action function to execute.
type Command struct {
	// The long form of the command. e.g., "delete"
	Name string
	// The shorthand aliases for the command. e.g., ["rm", "del"]
	Aliases []string
	// The function to execute when the command is run.
	Action func(todoID int) error
}

// Mode manages an interactive session, holding the available commands
// and current items being managed.
type Mode struct {
	commands map[string]*Command
	todos    []types.Todo
}

// New creates a new interactive mode with the provided commands.
// The commands map uses the command's Name as the key.
//
// Example:
//
//	interactive.New(map[string]*interactive.Command{
//		"delete": {
//			Name:    "delete",
//			Aliases: []string{"rm", "del"},
//			Action:  func(todoID int) error { return nil },
//		},
//	})
func New(commands map[string]*Command) *Mode {
	return &Mode{
		commands: commands,
	}
}

// Prompt starts an interactive session, displaying the current items and
// accepting user commands. It handles command parsing, validation, and
// execution. Returns an error if command execution fails.
func (m *Mode) Prompt(todos []types.Todo) error {
	m.todos = todos

	fmt.Print("Enter command (? for help): ")
	var input string
	fmt.Scanln(&input)

	if input == "q" || input == "quit" || input == "exit" {
		fmt.Print("Exiting interactive mode.\n\n")
		os.Exit(0)
	}

	if input == "?" || input == "help" {
		m.showHelp()
		return nil
	}

	return m.executeCommand(input)
}

// executeCommand handles both shorthand and longform commands. It parses the
// input, validates the item number, and either executes the command directly or
// prompts for one if none was provided.
func (m *Mode) executeCommand(input string) error {
	num, cmd, err := m.parseInput(input)
	if err != nil {
		return fmt.Errorf("unknown command: %s", input)
	}

	if num < 1 || num > len(m.todos) {
		return fmt.Errorf("invalid todo number: %d", num)
	}

	if cmd == "" {
		return m.promptForCommand(num)
	}

	return m.execute(num, cmd)
}

// parseInput handles both "2rm" style and plain number inputs. It returns the
// item number and command string, or an error if the input is invalid.
func (m *Mode) parseInput(input string) (num int, cmd string, err error) {
	re := regexp.MustCompile(`^(\d+)(.*)$`)
	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return 0, "", fmt.Errorf("input must start with a number")
	}

	num, _ = strconv.Atoi(matches[1])
	cmd = matches[2]
	return num, cmd, nil
}

// promptForCommand asks user for a command when only an item number was
// provided. It displays available commands with their aliases and validates
// the user's input.
func (m *Mode) promptForCommand(todoNum int) error {
	fmt.Println("\nAvailable commands:")
	for _, cmd := range m.commands {
		aliases := strings.Join(cmd.Aliases, "/")
		fmt.Printf("  %s (%s)\n", cmd.Name, aliases)
	}

	fmt.Print("\nEnter command: ")
	var cmd string
	fmt.Scanln(&cmd)

	return m.execute(todoNum, cmd)
}

// execute runs the specified command on the selected item. It checks both the
// command name and its aliases to find a match. Returns an error if the command
// is not recognized.
func (m *Mode) execute(todoNum int, cmdStr string) error {
	// First check aliases
	for _, cmd := range m.commands {
		if cmdStr == cmd.Name || slices.Contains(cmd.Aliases, cmdStr) {
			todoID := m.todos[todoNum-1].ID
			return cmd.Action(todoID)
		}
	}

	return fmt.Errorf("unknown command: %s", cmdStr)
}

// showHelp displays the available commands in interactive mode.
func (m *Mode) showHelp() {
	fmt.Println("\nUsage:")
	fmt.Println("  Enter a number to select a todo, followed by a command to perform an action on it.")
	fmt.Println("\nExamples:")
	fmt.Println("  1rm|1del|1delete will delete todo 1")
	fmt.Println("  1d|1done will mark todo 1 as done")
	fmt.Println("  1d-|1undone will mark todo 1 as not done")
	fmt.Println("  1u|1update will update the text of todo 1")
	fmt.Println("  1a|1archive will archive todo 1")
	fmt.Println("\nOther Commands:")
	fmt.Println("  ?        Show this help")
	fmt.Println("  q        Quit")
}
