// Package merge holds the conflict-resolution rules for sync. It is pure: no
// I/O, no clock, no database. The server uses Resolve, and the client sync
// engine uses Apply.
package merge

import (
	"github.com/kvnloughead/godo/internal/syncproto"
	"github.com/kvnloughead/godo/internal/todo"
)

// Resolve is called by the server when an incoming change conflicts with the
// stored row, meaning incoming.BaseVersion != stored.Version. It returns the
// change to store. The caller assigns the new version and seq.
//
// v1 rule: last writer wins by Mod. A delete that loses to an edit brings the
// todo back.
func Resolve(stored, incoming syncproto.Change) syncproto.Change {
	panic("merge.Resolve: not implemented")
}

// Apply is called by the client after a successful sync to write the server's
// changes into the local todos.
//
//   - local is the current contents of todo.txt and done.txt, re-read after
//     the sync response arrived.
//   - sent maps each ID pushed in the request to the todo as it was sent.
//   - remote is Response.Changes.
//
// The server's version wins for every todo in remote, except that a todo
// whose local line differs from what was sent (because it was edited while
// the request was in flight) keeps the local version, to be pushed on the
// next sync. Deleted changes remove the todo, and Archived moves it between
// files.
func Apply(local []todo.Todo, sent map[string]todo.Todo, remote []syncproto.Change) ([]todo.Todo, error) {
	panic("merge.Apply: not implemented")
}
