package todo

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	priorityRegex = regexp.MustCompile(`\([A-Z]\)`)
	projectRegex  = regexp.MustCompile(`\+\S+`)
	contextRegex  = regexp.MustCompile(`@\S+`)
	tagRegex      = regexp.MustCompile(`[a-zA-Z1-9]:\S+`)

	todoTxtDateLayout = "2006-01-02"
)

// Parse parses a single todo.txt line, including GoDo's id: and mod: tags. The
// format is described here :https://github.com/todotxt/todo.txt#todotxt-format-rules)
//
// Parse never sets Archived or Version, since neither is encoded in the line.
// It returns an error if the description is empty. Projects, contexts, and
// user supplied tags count as part of the description.
//
// It also returns an error if a GoDo tag is malformed (for example, a mod:
// value that isn't an RFC 3339 timestamp).
//
// Formatting the result with String must reproduce the input line, apart from
// moving id: and mod: to the end and normalizing whitespace.
func Parse(line string) (Todo, error) {
	words := strings.Fields(line)
	todo := Todo{}

	// first returns the first word in words or "" if words is empty
	first := func() string {
		if len(words) == 0 {
			return ""
		}
		return words[0]
	}

	if first() == "x" {
		todo.Completed = true
		words = words[1:]
	}

	if priorityRegex.MatchString(first()) {
		todo.Priority = string(words[0][1])
		words = words[1:]
	}

	if t1, err := time.Parse(todoTxtDateLayout, first()); err == nil {
		words = words[1:]
		if t2, err := time.Parse(todoTxtDateLayout, first()); err == nil {
			if t1.Before(t2) {
				return Todo{}, errors.New("Creation date must come before completion date.")
			}
			todo.CompletedAt = t1
			todo.CreatedAt = t2
			words = words[1:]
		} else {
			todo.CreatedAt = t1
		}
	}
	if len(words) == 0 {
		return Todo{}, errors.New("Description must not be empty.")
	}

	plainWords := []string{}
	tags := map[string]string{}
	for _, word := range words {
		if projectRegex.MatchString(word) {
			todo.Projects = append(todo.Projects, word[1:])
		} else if contextRegex.MatchString(word) {
			todo.Contexts = append(todo.Contexts, word[1:])
		} else if tagRegex.MatchString(word) {
			split := strings.Split(word, ":")
			tags[split[0]] = split[1]
		} else {
			plainWords = append(plainWords, word)
		}
	}

	if len(tags) != 0 {
		todo.Tags = tags
	}

	if len(plainWords) == 0 {
		return Todo{}, errors.New("Todo text must not be empty.")
	}

	todo.Description = strings.Join(plainWords, " ")

	return todo, nil
}

// Returns the string representation of a todo's tags, with one
// leading whitespace.
func (t Todo) TagString() string {
	nTags := len(t.Tags)
	if len(t.Tags) == 0 {
		return ""
	}

	keys := make([]string, 0, nTags)
	for k := range t.Tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := ""
	for _, k := range keys {
		result = result + fmt.Sprintf(" %s:%s", k, t.Tags[k])
	}

	return result
}

// Line formats t as a todo.txt line without the id: and mod: tags. This is the
// form sent over the wire in sync requests.
func (t Todo) Line() string {
	prefix := ""

	if t.Completed {
		prefix += "x "
	}
	if t.Priority != "" {
		prefix += fmt.Sprintf("(%s) ", t.Priority)
	}
	if !t.CompletedAt.IsZero() {
		prefix += t.CompletedAt.Format(todoTxtDateLayout) + " "
	}
	if !t.CreatedAt.IsZero() {
		prefix += t.CreatedAt.Format(todoTxtDateLayout) + " "
	}

	result := prefix + t.Description

	for _, p := range t.Projects {
		result = result + " +" + p
	}
	for _, c := range t.Contexts {
		result = result + " @" + c
	}
	result += t.TagString()

	return result
}

// String formats t as a full todo.txt line plus including id: and mod:. This
// is the form written to todo.txt and done.txt.
func (t Todo) String() string {
	result := t.Line()
	result += " mod:" + t.Mod.Format(time.RFC3339)
	result += " id:" + t.ID
	return result
}
