// Package store defines the Store interface that GoDo clients use to read and
// write todos. Local mode uses store/filestore and online mode uses
// store/apistore. The mode is chosen once, at startup.
package store

import (
	"context"
	"errors"

	"github.com/kvnloughead/godo/internal/todo"
)

var (
	// ErrNotFound is returned when no todo has the given ID.
	ErrNotFound = errors.New("todo not found")

	// ErrConflict is returned by online-mode writes when the todo's Version no
	// longer matches the server's (HTTP 409). The caller should refetch it.
	ErrConflict = errors.New("todo was changed elsewhere")

	// ErrUnavailable is returned by online-mode operations when the server
	// can't be reached.
	ErrUnavailable = errors.New("server unavailable")
)

// ArchiveFilter selects todos by archived status.
type ArchiveFilter int

const (
	ExcludeArchived ArchiveFilter = iota // default: only unarchived todos
	IncludeArchived
	OnlyArchived
)

// CompletionFilter selects todos by completion status.
type CompletionFilter int

const (
	AnyCompletion CompletionFilter = iota // default: done and undone todos
	OnlyDone
	OnlyUndone
)

// Filter narrows the todos returned by Store.List. The zero value returns all
// unarchived todos.
type Filter struct {
	// Text, if set, matches todos whose description contains it, ignoring
	// case. Searching for +project or @context works the same way.
	Text string

	Archived   ArchiveFilter
	Completion CompletionFilter
}

// Store reads and writes todos.
//
// Add assigns ID and Mod if they are unset. Update and Delete return
// ErrNotFound for an unknown ID. Archiving, completing, and changing priority
// are all done with Update.
type Store interface {
	// List returns the todos matching f in a stable order.
	List(ctx context.Context, f Filter) ([]todo.Todo, error)
	Get(ctx context.Context, id string) (todo.Todo, error)
	Add(ctx context.Context, t todo.Todo) (todo.Todo, error)
	Update(ctx context.Context, t todo.Todo) (todo.Todo, error)
	Delete(ctx context.Context, id string) error
}
