// Package filestore implements store.Store for local mode by reading and
// writing todo.txt and done.txt in a directory. It does no network I/O.
// Syncing is handled separately by package sync.
package filestore

import (
	"context"
	"errors"

	"github.com/kvnloughead/godo/internal/store"
	"github.com/kvnloughead/godo/internal/todo"
)

var errNotImplemented = errors.New("filestore: not implemented")

// Store is a store.Store backed by todo.txt (unarchived todos) and done.txt
// (archived todos) in a single directory.
type Store struct {
	dir string
}

var _ store.Store = (*Store)(nil)

// New returns a Store that reads and writes the todo files in dir.
func New(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) List(ctx context.Context, f store.Filter) ([]todo.Todo, error) {
	return nil, errNotImplemented
}

func (s *Store) Get(ctx context.Context, id string) (todo.Todo, error) {
	return todo.Todo{}, errNotImplemented
}

func (s *Store) Add(ctx context.Context, t todo.Todo) (todo.Todo, error) {
	return todo.Todo{}, errNotImplemented
}

func (s *Store) Update(ctx context.Context, t todo.Todo) (todo.Todo, error) {
	return todo.Todo{}, errNotImplemented
}

func (s *Store) Delete(ctx context.Context, id string) error {
	return errNotImplemented
}
