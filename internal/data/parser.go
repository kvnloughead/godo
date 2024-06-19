package data

import (
	"fmt"
	"regexp"
	"strings"
)

// priorityRX matches todo.txt style priority. Priorities are listed as (X) for
// X in A..Z. They must occur at the front of the string and be followed by one
// or more space.
var priorityRX = regexp.MustCompile(`^\(([A-Z])\) `)

// ParseTodo parses a string that is written in todo.txt format, as described in
// the GitHub repo: https://github.com/todotxt/todo.txt. It returns an instance
// of the Todo struct.
func ParseTodo(text string) (Todo, error) {
	var todo = Todo{}

	todo.Text = text
	todo.Completed = strings.HasPrefix(text, "x ")

	// If todo is completed, then any priority listed will not be effective.
	match := priorityRX.FindStringSubmatch(text)
	if len(match) > 0 {
		fmt.Println(match)
		todo.Priority = match[1]
	}

	return todo, nil
}
