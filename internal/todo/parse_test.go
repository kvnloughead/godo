package todo

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	// Basic tests
	basicTests := []struct {
		description string
		line        string
		expected    Todo
	}{
		{
			description: "Completed",
			line:        "x todo text",
			expected: Todo{
				Completed:   true,
				Description: "todo text",
			},
		},
		{
			description: "Extra whitespace",
			line:        " x  todo text  ",
			expected: Todo{
				Completed:   true,
				Description: "todo text",
			},
		},
		{
			description: "Not completed",
			line:        "todo text",
			expected: Todo{
				Description: "todo text",
			},
		},
		{
			description: "With priority",
			line:        "(A) todo text",
			expected: Todo{
				Description: "todo text",
				Priority:    "A",
			},
		},
	}

	for _, tt := range basicTests {
		t.Run(tt.description, func(t *testing.T) {
			todo, err := Parse(tt.line)
			if err != nil {
				t.Fatalf("Unexpected error, %v", err)
			}
			assert.Equal(t, tt.expected, todo)
		})

		testParseWithDates(t)
	}

	// Test errors
	errorTests := []struct {
		description string
		line        string
	}{
		{
			description: "Returns an error on empty string",
			line:        "",
		},
		{
			description: "Returns an error on with no plain todo text",
			line:        "x (A) +foo @bar",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.description, func(t *testing.T) {
			_, err := Parse(tt.line)
			if err == nil {
				t.Fatalf("Expected an error, got %v", err)
			}

		})
	}

	testParseWithDates(t)
	testParseWithProjectsTagsAndContexts(t)
	testString(t)
}

func testParseWithProjectsTagsAndContexts(t *testing.T) {
	happyTests := []struct {
		description string
		line        string
		expected    Todo
	}{
		{
			description: "With trailing projects and contexts",
			line:        "(A) todo text +foo +bar @con1 @con2",
			expected: Todo{
				Description: "todo text",
				Priority:    "A",
				Projects:    []string{"foo", "bar"},
				Contexts:    []string{"con1", "con2"},
			},
		},
		{
			description: "With projects and contexts mixed up",
			line:        "(A) todo text +foo @con1 +bar @con2",
			expected: Todo{
				Description: "todo text",
				Priority:    "A",
				Projects:    []string{"foo", "bar"},
				Contexts:    []string{"con1", "con2"},
			},
		},
		{
			description: "With projects contexts and tags",
			line:        "(A) todo text +foo due:tuesday @con1 +bar @con2",
			expected: Todo{
				Description: "todo text",
				Priority:    "A",
				Projects:    []string{"foo", "bar"},
				Contexts:    []string{"con1", "con2"},
				Tags:        map[string]string{"due": "tuesday"},
			},
		},
	}

	for _, tt := range happyTests {
		t.Run(tt.description, func(t *testing.T) {
			todo, err := Parse(tt.line)
			if err != nil {
				t.Fatalf("Unexpected error, %v", err)
			}
			assert.Equal(t, tt.expected, todo)
		})
	}

	errorTests := []struct {
		description string
		line        string
	}{
		{
			description: "CreatedAt is before CompletedAt",
			line:        "x (A) 2026-01-02 2026-02-02 todo text",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.description, func(t *testing.T) {
			_, err := Parse(tt.line)
			if err == nil {
				t.Fatalf("Expected an error, got nil")
			}
		})
	}
}

var (
	completedAtString = "2026-02-02"
	createdAtString   = "2026-01-02"
)

func computeExpectedDates() (time.Time, time.Time, error) {
	layout := "2006-01-02"

	expectedCreatedAt, err := time.Parse(layout, createdAtString)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf(fmt.Sprintf("Failed to parse time: layout='%s' dateStr='%s", layout, createdAtString))
	}

	expectedCompletedAt, err := time.Parse(layout, completedAtString)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf(fmt.Sprintf("Failed to parse time: layout='%s' dateStr='%s", layout, completedAtString))
	}

	return expectedCompletedAt, expectedCreatedAt, nil
}

func testParseWithDates(t *testing.T) {
	expectedCompletedAt, expectedCreatedAt, err := computeExpectedDates()
	if err != nil {
		t.Fatalf(err.Error())
	}

	happyTests := []struct {
		description string
		line        string
		expected    Todo
	}{
		{
			description: "With CreatedAt",
			line:        "x (A) 2026-01-02 todo text",
			expected: Todo{
				Completed:   true,
				Description: "todo text",
				Priority:    "A",
				CreatedAt:   expectedCreatedAt,
			},
		},
		{
			description: "With CreatedAt and CompletedAt",
			line:        "x (A) 2026-02-02 2026-01-02 todo text",
			expected: Todo{
				Completed:   true,
				Description: "todo text",
				Priority:    "A",
				CreatedAt:   expectedCreatedAt,
				CompletedAt: expectedCompletedAt,
			},
		},
		{
			description: "Invalid format remains in the text",
			line:        "x (A) 02-02-2026  todo text",
			expected: Todo{
				Completed:   true,
				Description: "02-02-2026 todo text",
				Priority:    "A",
			},
		},
	}

	for _, tt := range happyTests {
		t.Run(tt.description, func(t *testing.T) {
			todo, err := Parse(tt.line)
			if err != nil {
				t.Fatalf("Unexpected error, %v", err)
			}
			assert.Equal(t, todo, tt.expected)

		})
	}

	errorTests := []struct {
		description string
		line        string
	}{
		{
			description: "CreatedAt is before CompletedAt",
			line:        "x (A) 2026-01-02 2026-02-02 todo text",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.description, func(t *testing.T) {
			_, err := Parse(tt.line)
			if err == nil {
				t.Fatalf("Expected an error, got nil")
			}
		})
	}
}

func testTagString(t *testing.T) {
	tests := []struct {
		description string
		todo        Todo
		expected    string
		repeat      int
	}{
		{
			description: "No tags returns an empty string",
			todo:        Todo{},
			expected:    "",
		},
		{
			description: "Single tag",
			todo: Todo{
				Tags: map[string]string{"foo": "bar"},
			},
			expected: " foo:bar",
		},
		{
			description: "Lexicographic order is preserved",
			todo: Todo{
				Tags: map[string]string{"a": "a", "z": "z", "b": "b", "y": "y", "c": "c", "x": "x"},
			},
			expected: " a:a b:b c:c x:x y:y z:z",
			repeat:   9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			reps := tt.repeat + 1
			for range reps {
				actual := tt.todo.TagString()
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func testString(t *testing.T) {
	expectedCompletedAt, expectedCreatedAt, err := computeExpectedDates()
	if err != nil {
		t.Fatalf(err.Error())
	}

	tests := []struct {
		description string
		todo        Todo
		expected    string
	}{
		{
			description: "",
			todo: Todo{
				Priority:    "M",
				Description: "do something",
				CompletedAt: expectedCompletedAt,
				CreatedAt:   expectedCreatedAt,
				Projects:    []string{"foo", "bar"},
				Contexts:    []string{"con1", "con2"},
				Tags:        map[string]string{"due": "tuesday", "foo": "bar"},
			},
			expected: fmt.Sprintf("(M) %s %s do something +foo +bar @con1 @con2 due:tuesday foo:bar", completedAtString, createdAtString),
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			actual := tt.todo.String()
			assert.Equal(t, tt.expected, actual)

			tt.todo.Completed = true
			actual = tt.todo.String()
			assert.Equal(t, fmt.Sprintf("x %s", tt.expected), actual)
		})
	}
}
