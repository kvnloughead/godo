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

// Command represents a todo action
type Commands struct {
	Name    string   // e.g., "delete"
	Aliases []string // e.g., ["rm", "del"]
	Action  func(todoID int) error
}

// Mode represents an interactive session
type Mode struct {
	commands map[string]*Commands
	todos    []types.Todo
}

// New creates a new interactive mode with the predefined commands
func New(commands map[string]*Commands) *Mode {
	return &Mode{
		commands: commands,
	}
}

// Prompt is the main entry point for interactive mode
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
		m.ShowHelp()
		return nil
	}

	return m.ExecuteCommand(input)
}

// ExecuteCommand handles both shorthand and longform commands
func (m *Mode) ExecuteCommand(input string) error {
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

// parseInput handles both "2rm" style and plain number inputs
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

// promptForCommand asks user for a command and validates it
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

// execute runs the specified command on the todo
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
