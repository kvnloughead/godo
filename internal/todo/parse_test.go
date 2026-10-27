package todo

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// parseTimeMustCompile accepts a date string and a layout string and parses
// with time.Parse. It panics if a parsing error is returned. Not intended for
// production use.
func parseTimeMustCompile(layout, dateStr string) time.Time {
	t, err := time.Parse(layout, dateStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse date string: layout=%s time=%s err=%v", layout, dateStr, err))
	}
	return t
}

var (
	completedAtString = "2026-02-02"
	createdAtString   = "2026-01-02"
	RFC3339String     = "2026-10-26T14:53:13Z"

	expectedCompletedAt = parseTimeMustCompile(todoTxtDateLayout, completedAtString)
	expectedCreatedAt   = parseTimeMustCompile(todoTxtDateLayout, createdAtString)
	expectedMod         = parseTimeMustCompile(time.RFC3339, RFC3339String)
)

func TestExtractModAndID(t *testing.T) {
	tests := []struct {
		description string
		word        string
		assertions  func(todo *Todo, prefix string, err error)
	}{
		{
			description: "If not mod: or id: it's a no-op",
			word:        "z",
			assertions: func(todo *Todo, prefix string, err error) {
				assert.Nil(t, err)
				assert.Equal(t, &Todo{}, todo)
			},
		},
		{
			description: "Valid mod: is parsed",
			word:        "mod:" + RFC3339String,
			assertions: func(todo *Todo, prefix string, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "mod", prefix)
				assert.Equal(t, &Todo{Mod: expectedMod}, todo)
			},
		},
		{
			description: "Valid id: is parsed",
			word:        "id:123",
			assertions: func(todo *Todo, prefix string, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "id", prefix)
				assert.Equal(t, &Todo{ID: "123"}, todo)
			},
		},
		{
			description: "Invalid mod: returns an time.ParseError",
			word:        "mod:bad-date",
			assertions: func(todo *Todo, prefix string, err error) {
				var parseErr *time.ParseError
				assert.ErrorAs(t, err, &parseErr)
				assert.Equal(t, &Todo{}, todo)
			},
		},
		{
			description: "Invalid id: returns an time.ParseError",
			word:        "id://",
			assertions: func(todo *Todo, prefix string, err error) {
				assert.EqualError(t, err, "ID must consist of 1 to 64 alphanumerics or hyphens; got //")
				assert.Equal(t, &Todo{}, todo)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			todo := Todo{}
			prefix, err := ExtractModAndID(tt.word, &todo)
			tt.assertions(&todo, prefix, err)
		})
	}
}

func TestParseBasic(t *testing.T) {
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
	}
}

func TestParseWithEmptyDescription(t *testing.T) {
	cases := []string{
		"", " ", " x", "x (B)", "(B) 2025-02-01", "x 2025-02-01  2025-01-01", fmt.Sprintf("x 2025-02-01  2025-01-01 mod:%s id:123", RFC3339String),
	}

	for _, s := range cases {
		_, err := Parse(s)
		assert.EqualError(t, err, "Description must not be empty.")
	}
}

func TestParseWithDates(t *testing.T) {
	happyTests := []struct {
		description string
		line        string
		expected    Todo
	}{
		{
			description: "Uncompleted with CreatedAt",
			line:        "(A) 2026-01-02 todo text",
			expected: Todo{
				Completed:   false,
				Description: "todo text",
				Priority:    "A",
				CreatedAt:   expectedCreatedAt,
			},
		},
		{
			description: "Uncompleted can only have a creation date, a second date is left in the description",
			line:        "(A) 2026-01-02 2026-02-02 todo text",
			expected: Todo{
				Completed:   false,
				Description: "2026-02-02 todo text",
				Priority:    "A",
				CreatedAt:   expectedCreatedAt,
			},
		},
		{
			description: "Completed with one date",
			line:        "x (A) 2026-02-02 todo text",
			expected: Todo{
				Completed:   true,
				Description: "todo text",
				Priority:    "A",
				CompletedAt: expectedCompletedAt,
			},
		},
		{
			description: "Completed with two dates",
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
			description: "Completed may be the same as CreatedAt",
			line:        "x (A) 2026-02-02 2026-02-02 todo text",
			expected: Todo{
				Completed:   true,
				Description: "todo text",
				Priority:    "A",
				CreatedAt:   expectedCompletedAt,
				CompletedAt: expectedCompletedAt,
			},
		},
		{
			description: "Invalid format is left in the description",
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

func TestParseTag(t *testing.T) {
	tests := []struct {
		description string
		tag         string
		assertions  func(tag string, todo *Todo, tags map[string]string, err error)
	}{
		{
			description: "Valid mod:",
			tag:         "mod:" + RFC3339String,
			assertions: func(tag string, todo *Todo, tags map[string]string, err error) {
				assert.Equal(t, expectedMod, todo.Mod)
			},
		},
		{
			description: "Invalid mod:",
			tag:         "mod:" + "2026-01-02",
			assertions: func(tag string, todo *Todo, tags map[string]string, err error) {
				if err == nil {
					t.Fatal("Expected error, got nil.")
				}
			},
		},
		{
			description: "Valid id:",
			tag:         "id:" + "B8CF9dCB-25CA-4E1A-98FB-9F32Fb6F17",
			assertions: func(tag string, todo *Todo, tags map[string]string, err error) {
				assert.Equal(t, "B8CF9dCB-25CA-4E1A-98FB-9F32Fb6F17", todo.ID)
			},
		},
		{
			description: "id: with invalid ID returns an error",
			tag:         "id:" + "//foo.bar",
			assertions: func(tag string, todo *Todo, tags map[string]string, err error) {
				if err == nil {
					t.Fatal("Expected error, got nil.")
				}
			},
		},
		{
			description: "key:val",
			tag:         "key:val",
			assertions: func(tag string, todo *Todo, tags map[string]string, err error) {
				assert.Equal(t, map[string]string{"key": "val"}, tags)
			},
		},
		{
			description: "Alphanumerics and underscores are accepted for keys and values",
			tag:         "azAz09_:_azAz09",
			assertions: func(tag string, todo *Todo, tags map[string]string, err error) {
				assert.Equal(t, map[string]string{"azAz09_": "_azAz09"}, tags)
			},
		},
	}

	for _, tt := range tests {
		todo := Todo{}
		tags := map[string]string{}
		t.Run(tt.description, func(t *testing.T) {
			err := ParseTag(tt.tag, &todo, tags)
			tt.assertions(tt.tag, &todo, tags, err)
		})
	}

}

func TestParseWithProjectsTagsAndContexts(t *testing.T) {
	happyTests := []struct {
		description string
		line        string
		expected    Todo
	}{
		{
			description: "With trailing projects and contexts",
			line:        "(A) todo text +foo +bar @con1 @con2",
			expected: Todo{
				Description: "todo text +foo +bar @con1 @con2",
				Priority:    "A",
				Projects:    []string{"foo", "bar"},
				Contexts:    []string{"con1", "con2"},
			},
		},
		{
			description: "With projects and contexts mixed up",
			line:        "(A) todo text +foo @con1 +bar @con2",
			expected: Todo{
				Description: "todo text +foo @con1 +bar @con2",
				Priority:    "A",
				Projects:    []string{"foo", "bar"},
				Contexts:    []string{"con1", "con2"},
			},
		},
		{
			description: "With projects contexts and tags",
			line:        "(A) todo text +foo due:tuesday @con1 +bar @con2",
			expected: Todo{
				Description: "todo text +foo due:tuesday @con1 +bar @con2",
				Priority:    "A",
				Projects:    []string{"foo", "bar"},
				Contexts:    []string{"con1", "con2"},
				Tags:        map[string]string{"due": "tuesday"},
			},
		},
		{
			description: "Projects and contexts accept any non-whitespace characters",
			line:        "(A) todo text +https://github.com/todotxt/todo.txt#rule-3-contexts-and-projects-may-appear-anywhere-in-the-line-after-priorityprepended-date @https://github.com/todotxt/todo.txt#rule-3-contexts-and-projects-may-appear-anywhere-in-the-line-after-priorityprepended-date",
			expected: Todo{
				Description: "todo text +https://github.com/todotxt/todo.txt#rule-3-contexts-and-projects-may-appear-anywhere-in-the-line-after-priorityprepended-date @https://github.com/todotxt/todo.txt#rule-3-contexts-and-projects-may-appear-anywhere-in-the-line-after-priorityprepended-date",
				Priority:    "A",
				Projects:    []string{"https://github.com/todotxt/todo.txt#rule-3-contexts-and-projects-may-appear-anywhere-in-the-line-after-priorityprepended-date"},
				Contexts:    []string{"https://github.com/todotxt/todo.txt#rule-3-contexts-and-projects-may-appear-anywhere-in-the-line-after-priorityprepended-date"},
			},
		},
		{
			description: "With valid mod: and id:",
			line:        fmt.Sprintf("(A) todo text +foo due:tuesday @con1 +bar @con2 mod:%s id:123", RFC3339String),
			expected: Todo{
				Description: "todo text +foo due:tuesday @con1 +bar @con2",
				Priority:    "A",
				Projects:    []string{"foo", "bar"},
				Contexts:    []string{"con1", "con2"},
				Tags:        map[string]string{"due": "tuesday"},
				Mod:         expectedMod,
				ID:          "123",
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
			description: "mod: and id: must not be before the body",
			line:        fmt.Sprintf("(A) mod:%s todo text +foo due:tuesday @con1 +bar @con2 id:123", RFC3339String),
		},
		{
			description: "mod: and id: must not be inside the body",
			line:        fmt.Sprintf("(A) todo text +foo due:tuesday id:123 @con1 +bar @con2 mod:%s ", RFC3339String),
		},
		{
			description: "multiple mod: tags are not allowed",
			line:        fmt.Sprintf("(A) todo text +foo due:tuesday id:123 @con1 +bar @con2 mod:%s mod:%s", RFC3339String, RFC3339String),
		},
		{
			description: "multiple id: tags are not allowed",
			line:        "(A) todo text +foo due:tuesday id:123 @con1 +bar @con2 id:123 id:123",
		},
		{
			description: "mod: must be a valid RFC3339String",
			line:        fmt.Sprintf("(A) todo text +foo due:tuesday @con1 +bar @con2 mod:%s id:123", completedAtString),
		},
		{
			description: "id: may only contain between 1 and 64 alphanumerics or hyphens",
			line:        fmt.Sprintf("(A) todo text +foo due:tuesday @con1 +bar @con2 mod:%s id://", RFC3339String),
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

/* -------------------------- String and Line tests ------------------------- */

type stringAndLineTests []struct {
	description string
	todo        Todo
	expected    string
}

func TestLine(t *testing.T) {
	tests := stringAndLineTests{
		{
			description: "Basic Line test",
			todo: Todo{
				Priority:    "M",
				Description: "do something +foo +bar @con1 @con2 due:tuesday foo:bar",
				CompletedAt: expectedCompletedAt,
				CreatedAt:   expectedCreatedAt,
				Projects:    []string{"foo", "bar"},
				Contexts:    []string{"con1", "con2"},
				Tags:        map[string]string{"due": "tuesday", "foo": "bar"},
			},
			expected: fmt.Sprintf("(M) %s %s do something +foo +bar @con1 @con2 due:tuesday foo:bar", completedAtString, createdAtString),
		},
		{
			description: "mod: and id: should be omitted",
			todo: Todo{
				ID:          "01M5WB5V23APQXY9CZQ49HNABG",
				Mod:         expectedMod,
				Priority:    "M",
				Description: "do something +foo +bar @con1 @con2 due:tuesday foo:bar",
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
			actual := tt.todo.Line()
			assert.Equal(t, tt.expected, actual)

			tt.todo.Completed = true
			actual = tt.todo.Line()
			assert.Equal(t, fmt.Sprintf("x %s", tt.expected), actual)
		})
	}
}

func TestString(t *testing.T) {
	tests := stringAndLineTests{
		{
			description: "Basic String test",
			todo: Todo{
				ID:          "01M5WB5V23APQXY9CZQ49HNABG",
				Mod:         expectedMod,
				Priority:    "M",
				Description: "do something +foo +bar @con1 @con2 due:tuesday foo:bar",
				CompletedAt: expectedCompletedAt,
				CreatedAt:   expectedCreatedAt,
				Projects:    []string{"foo", "bar"},
				Contexts:    []string{"con1", "con2"},
				Tags:        map[string]string{"due": "tuesday", "foo": "bar"},
			},
			expected: fmt.Sprintf("(M) %s %s do something +foo +bar @con1 @con2 due:tuesday foo:bar id:01M5WB5V23APQXY9CZQ49HNABG mod:%s", completedAtString, createdAtString, RFC3339String),
		},
		{
			description: "mod: and id: should be omitted if they are empty or zero-valued",
			todo: Todo{
				Priority:    "M",
				Description: "do something +foo +bar @con1 @con2 due:tuesday foo:bar",
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
			assert.Equal(t, "x "+tt.expected, actual)
		})
	}
}
