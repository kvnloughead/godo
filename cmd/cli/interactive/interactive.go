// Package interactive provides an interactive command-line interface for
// managing items through commands and their aliases. It supports both shorthand
// (e.g., "1rm") and longform command entry with command prompting.
package interactive

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/kvnloughead/godo/cmd/cli/types"
)

// Command represents an action that can be performed on an item. Each command
// has a primary name, optional aliases, and an action function to execute.
type Command struct {
	// The long form of the command. e.g., "delete"
	Name string
	// The shorthand aliases for the command. e.g., ["rm", "del"]
	Aliases []string
	// The function to execute when the command is run.
	Action func([]int) error
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

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}
	input = strings.TrimSpace(input)

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

// executeCommand handles command execution with multiple todo IDs
func (m *Mode) executeCommand(input string) error {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return fmt.Errorf("input cannot be empty")
	}

	cmdStr := fields[0]
	var cmd *Command

	// Find matching command
	for _, c := range m.commands {
		if cmdStr == c.Name || slices.Contains(c.Aliases, cmdStr) {
			cmd = c
			break
		}
	}

	if cmd == nil {
		return fmt.Errorf("unknown command: %s", cmdStr)
	}

	// Parse todo numbers
	if len(fields) < 2 {
		return fmt.Errorf("no todo numbers provided")
	}

	var ids []int
	for _, numStr := range fields[1:] {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return fmt.Errorf("invalid todo number: %s", numStr)
		}
		if num < 1 || num > len(m.todos) {
			return fmt.Errorf("todo number out of range: %d", num)
		}
		// Convert from 1-based display number to actual todo ID
		ids = append(ids, m.todos[num-1].ID)
	}

	return cmd.Action(ids)
}

// showHelp displays the available commands in interactive mode.
func (m *Mode) showHelp() {
	fmt.Println("\nUsage:")
	fmt.Println("  Enter a command (or command alias) followed by one or more todo numbers")
	fmt.Println("\nExamples:")
	fmt.Println("  rm 1 2 3      Delete todos 1, 2, and 3")
	fmt.Println("  done 4 5      Mark todos 4 and 5 as done")
	fmt.Println("  archive 6     Archive todo 6")
	fmt.Println("\nCommands:")
	for _, cmd := range m.commands {
		aliases := strings.Join(cmd.Aliases, "/")
		fmt.Printf("  %s (%s)\n", cmd.Name, aliases)
	}
	fmt.Println("\nOther Commands:")
	fmt.Println("  ?        Show this help")
	fmt.Println("  q        Quit")
}
