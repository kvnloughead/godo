// Package syncproto defines the wire types for POST /v1/sync. It is shared by
// the server (internal/server/syncsvc) and the client sync engine
// (internal/sync), and depends only on package todo.
package syncproto

import (
	"time"

	"github.com/kvnloughead/godo/internal/todo"
)

// Request is the body of POST /v1/sync.
type Request struct {
	// Since is the highest seq the client has applied. Zero requests a full
	// resync.
	Since int64 `json:"since"`

	// Changes are the client's local changes since its last successful sync.
	Changes []Change `json:"changes"`
}

// Response is the body of a successful POST /v1/sync.
type Response struct {
	// Seq is the new high-water mark. The client stores it as its last_seq.
	Seq int64 `json:"seq"`

	// Changes holds every row for the user with seq > Request.Since, in its
	// final, server-resolved form. This includes the rows the client just
	// pushed.
	Changes []Change `json:"changes"`
}

// Change is one todo in a sync request or response.
//
// In a request, BaseVersion is the server version the client last saw for
// this todo (0 for a new todo), and Version is unset. In a response, Version
// is the todo's current server version, and BaseVersion is unset.
type Change struct {
	ID string `json:"id"`

	// Line is the todo.txt line without the id: and mod: tags (todo.Todo.Line).
	Line string `json:"line"`

	Archived bool      `json:"archived"`
	Deleted  bool      `json:"deleted"`
	Mod      time.Time `json:"mod"`

	BaseVersion int `json:"base_version,omitempty"`
	Version     int `json:"version,omitempty"`
}

// FromTodo returns the request Change for t. Set deleted for a todo that was
// removed locally.
func FromTodo(t todo.Todo, deleted bool) Change {
	return Change{
		ID:          t.ID,
		Line:        t.Line(),
		Archived:    t.Archived,
		Deleted:     deleted,
		Mod:         t.Mod,
		BaseVersion: t.Version,
	}
}

// Todo parses c into a todo.Todo, restoring the fields that aren't encoded in
// Line. It is meaningless for a deleted change.
func (c Change) Todo() (todo.Todo, error) {
	t, err := todo.Parse(c.Line)
	if err != nil {
		return todo.Todo{}, err
	}
	t.ID = c.ID
	t.Mod = c.Mod
	t.Archived = c.Archived
	t.Version = c.Version
	return t, nil
}
