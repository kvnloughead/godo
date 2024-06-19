package data

import (
	"testing"

	"github.com/kvnloughead/godo/internal/assert"
)

func TestParseTodo(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		expectedTodo  Todo
		expectedError bool
	}{

		{
			name: "Lowercase x in front is completed",
			text: "x (A) do something",
			expectedTodo: Todo{
				Text: "x (A) do something", Completed: true,
			},
		},
		{
			name: "Capital X isn't completed",
			text: "X do something",
			expectedTodo: Todo{
				Text: "X do something", Completed: false,
			},
		},
		{
			name: "x in the middle is not completed",
			text: "(A) x do something",
			expectedTodo: Todo{
				Text: "(A) x do something", Completed: false, Priority: "A",
			},
		},
		{
			name: "(B) in front => priority B",
			text: "(B) do something",
			expectedTodo: Todo{
				Text: "(B) do something", Completed: false, Priority: "B",
			},
		},
		{
			name: "(b) in front => no priority",
			text: "(b) do something",
			expectedTodo: Todo{
				Text: "(b) do something", Completed: false,
			},
		},
		{
			name: "(B) in the middle front => no priority",
			text: "do (B) something",
			expectedTodo: Todo{
				Text: "do (B) something", Completed: false,
			},
		},
		{
			name: "No space after (B) => no priority",
			text: "(B)do something",
			expectedTodo: Todo{
				Text: "(B)do something", Completed: false,
			},
		},
		{
			name: "B in front => no priority",
			text: "B do something",
			expectedTodo: Todo{
				Text: "B do something", Completed: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo, err := ParseTodo(tt.text)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, todo.Text, tt.expectedTodo.Text)
			assert.Equal(t, todo.Completed, tt.expectedTodo.Completed)
			assert.Equal(t, todo.Priority, tt.expectedTodo.Priority)
			assert.Equal(t, len(todo.Contexts), 0)
			assert.Equal(t, len(todo.Projects), 0)
		})
	}
}
