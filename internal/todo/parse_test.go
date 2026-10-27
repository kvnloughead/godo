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
	}

	testParseWithProjectsTagsAndContexts(t)
	testParseWithEmptyDescription(t)
}

func testParseWithEmptyDescription(t *testing.T) {
	cases := []string{
		"", " ", " x", "x (B)", "(B) 2025-02-01", "x 2025-02-01  2025-01-01",
	}

	for _, s := range cases {
		_, err := Parse(s)
		assert.EqualError(t, err, "Description must not be empty.")
	}
}

func TestParseWithDates(t *testing.T) {
	expectedCompletedAt, expectedCreatedAt, _, err := computeExpectedDates()
	if err != nil {
		t.Fatalf(err.Error())
	}

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
			description: "Completed with CreatedAt",
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
	_, _, expectedMod, err := computeExpectedDates()
	if err != nil {
		t.Fatalf(err.Error())
	}

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
}

var (
	completedAtString = "2026-02-02"
	createdAtString   = "2026-01-02"
	RFC3339String     = "2026-10-26T14:53:13Z"
)

// computeExpectedDates computes time.Time values to use as expected values for
// CompletedAt, CreatedAt, and Mod. The first two use the "2006-01-02" format,
// and Mod uses RFC3339.
func computeExpectedDates() (expectedCompletedAt, expectedCreatedAt, expectedMod time.Time, err error) {
	layout := "2006-01-02"

	expectedCreatedAt, err = time.Parse(layout, createdAtString)
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, fmt.Errorf("Failed to parse time: layout=%q dateStr=%q: %w", layout, createdAtString, err)
	}

	expectedCompletedAt, err = time.Parse(layout, completedAtString)
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, fmt.Errorf("Failed to parse time: layout=%q dateStr=%q: %w", layout, completedAtString, err)
	}

	expectedMod, err = time.Parse(time.RFC3339, RFC3339String)
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, fmt.Errorf("Failed to parse time: layout=%q dateStr=%q: %w", time.RFC3339, RFC3339String, err)
	}

	return expectedCompletedAt, expectedCreatedAt, expectedMod, nil
}

type stringAndLineTests []struct {
	description string
	todo        Todo
	expected    string
}

// prepareStringAndLineTests receives the expected date values and uses them to
// create a set of tests usable for both String and Line tests. Use
// computeExpectedDates to get the needed time.Time values.
func prepareStringAndLineTests(expectedCompletedAt, expectedCreatedAt, expectedMod time.Time) stringAndLineTests {
	tests := stringAndLineTests{
		{
			description: "",
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

	return tests
}

func TestLine(t *testing.T) {
	expectedCompletedAt, expectedCreatedAt, expectedMod, err := computeExpectedDates()
	if err != nil {
		t.Fatalf(err.Error())
	}

	tests := prepareStringAndLineTests(expectedCompletedAt, expectedCreatedAt, expectedMod)
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
	expectedCompletedAt, expectedCreatedAt, expectedMod, err := computeExpectedDates()
	if err != nil {
		t.Fatalf(err.Error())
	}

	tests := prepareStringAndLineTests(expectedCompletedAt, expectedCreatedAt, expectedMod)
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			suffix := fmt.Sprintf(" mod:%s id:%s", RFC3339String, tt.todo.ID)

			actual := tt.todo.String()
			expected := tt.expected + suffix
			assert.Equal(t, expected, actual)

			tt.todo.Completed = true
			actual = tt.todo.String()
			expected = "x " + tt.expected + suffix
			assert.Equal(t, expected, actual)
		})
	}
}
