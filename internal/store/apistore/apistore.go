// Package apistore implements store.Store for online mode by calling the
// /todos API for every operation.
package apistore

import (
	"context"
	"errors"
	"net/http"

	"github.com/kvnloughead/godo/internal/store"
	"github.com/kvnloughead/godo/internal/todo"
)

var errNotImplemented = errors.New("apistore: not implemented")

// Store is a store.Store backed by the GoDo HTTP API.
type Store struct {
	baseURL string
	token   string
	client  *http.Client
}

var _ store.Store = (*Store)(nil)

// New returns a Store that sends requests to baseURL (for example
// "https://godo.example.com/v1"), authenticated with token.
func New(baseURL, token string, client *http.Client) *Store {
	return &Store{baseURL: baseURL, token: token, client: client}
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
