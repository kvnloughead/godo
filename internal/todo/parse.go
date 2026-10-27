package todo

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	priorityRegex = regexp.MustCompile(`^\([A-Z]\)$`)
	projectRegex  = regexp.MustCompile(`^\+\S+$`)
	contextRegex  = regexp.MustCompile(`^@\S+$`)
	tagRegex      = regexp.MustCompile(`^[a-zA-Z0-9_]+:[a-zA-Z0-9_]+$`)

	idRegex = regexp.MustCompile(`^[0-9A-Za-z-]{1,64}$`)

	todoTxtDateLayout = "2006-01-02"
)

// Parse parses a single todo.txt line, including GoDo's id: and mod: tags. A
// summary of the format can be found in docs/TODOTXT.md. For more details, see
// here: https://github.com/todotxt/todo.txt#todotxt-format-rules.
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

	// last returns the last word in words or "" if words is empty
	last := func() string {
		if len(words) == 0 {
			return ""
		}
		return words[len(words)-1]
	}

	// Extract and validate mod: and id: tags from the end of the string. Errors
	// are returned if a tag is invalid or if more than one of each tag is found.
	prefix, err := ExtractModAndID(last(), &todo)
	if err != nil {
		return Todo{}, err
	}
	if prefix != "" {
		words = words[:len(words)-1]
		nextPrefix, err := ExtractModAndID(last(), &todo)
		if err != nil {
			return Todo{}, err
		}
		if prefix == nextPrefix {
			return Todo{}, fmt.Errorf("Multiple %s tags are not allowed", prefix)
		}
		words = words[:len(words)-1]
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
		if !todo.Completed {
			todo.CreatedAt = t1
		} else {
			todo.CompletedAt = t1
			if t2, err := time.Parse(todoTxtDateLayout, first()); err == nil {
				if t1.Before(t2) {
					return Todo{}, errors.New("Creation date must come before completion date.")
				}
				todo.CompletedAt = t1
				todo.CreatedAt = t2
				words = words[1:]
			}
		}
	}

	if len(words) == 0 {
		return Todo{}, errors.New("Description must not be empty.")
	}

	tags := map[string]string{}
	for _, word := range words {
		prefix, _, _ := strings.Cut(word, ":")
		if prefix == "mod" || prefix == "id" {
			return Todo{}, errors.New("mod: and id: tags are reserved for internal use and must appear at the end of the line.")
		}
		if projectRegex.MatchString(word) {
			todo.Projects = append(todo.Projects, word[1:])
		} else if contextRegex.MatchString(word) {
			todo.Contexts = append(todo.Contexts, word[1:])
		} else if tagRegex.MatchString(word) {
			err := ParseTag(word, &todo, tags)
			if err != nil {
				return Todo{}, err
			}
		}
	}

	if len(tags) != 0 {
		todo.Tags = tags
	}

	todo.Description = strings.Join(words, " ")

	return todo, nil
}

// ExtractModAndID checks if the given word starts with mod: or id:. If so,
// it parses it with ParseTag. Otherwise it does nothing. It returns the prefix
// ("mod" or "id") and and any errors from ParseTag.
func ExtractModAndID(word string, todo *Todo) (string, error) {
	prefix, _, _ := strings.Cut(word, ":")
	if prefix == "mod" || prefix == "id" {
		err := ParseTag(word, todo, map[string]string{})
		if err != nil {
			return "", err
		}
		return prefix, nil
	}
	return "", nil
}

// ParseTag validates and parses a todo.txt key:value flag. If the key is mod
// or id, it gets routed to Todo.Mod or Todo.ID. Otherwise the pair is added to
// the provided tags map. Errors are returned if any of the following
// validations fail:
//
// - mod: values must be valid RFC3339 times.
//
// - id: values are loosely validated to contain 1 to 64 alphanumerics and
// hyphens (i.e, the character set of a UUID).
//
// - User supplied keys and values may only contain alphanumerics and
// underscores.
func ParseTag(tag string, todo *Todo, tags map[string]string) error {
	split := strings.SplitN(tag, ":", 2)
	key := split[0]
	val := split[1]

	switch key {
	case "mod":
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return err
		}
		todo.Mod = t
	case "id":
		if !idRegex.MatchString(val) {
			return fmt.Errorf("ID must consist of 1 to 64 alphanumerics or hyphens; got %s", val)
		}
		todo.ID = val
	default:
		if !tagRegex.MatchString(tag) {
			return errors.New("Keys and values can only contain alphanumerics or underscores.")
		}
		tags[key] = val
	}
	return nil
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
