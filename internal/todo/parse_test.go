package todo

import (
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

func testParseWithDates(t *testing.T) {
	layout := "2006-01-02"
	createdAtString := "2026-01-02"

	expectedCreatedAt, err := time.Parse(layout, createdAtString)
	if err != nil {
		t.Fatalf("Failed to parse time: layout='%s' dateStr='%s", layout, createdAtString)
	}

	completedAtString := "2026-02-02"
	expectedCompletedAt, err := time.Parse(layout, completedAtString)
	if err != nil {
		t.Fatalf("Failed to parse time: layout='%s' dateStr='%s", layout, completedAtString)
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
