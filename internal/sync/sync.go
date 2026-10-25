// Package sync is the local-mode sync engine. It diffs the todo files against
// the base snapshot from the last successful sync, pushes the differences
// with POST /v1/sync, and writes the server's response back to the files.
package sync

import (
	"context"
	"errors"
	"time"

	"github.com/kvnloughead/godo/internal/syncproto"
)

var errNotImplemented = errors.New("sync: not implemented")

// ErrResyncRequired is returned by a Client when the server responds 410 Gone
// because the client's last_seq is older than the tombstone pruning horizon.
// The engine handles it by syncing again with Since set to 0.
var ErrResyncRequired = errors.New("full resync required")

// Client sends sync requests to the server. Tests can substitute a fake.
type Client interface {
	Sync(ctx context.Context, req syncproto.Request) (syncproto.Response, error)
}

// Status describes this device's sync state.
type Status struct {
	// Pending is the number of local changes not yet accepted by the server.
	Pending int

	// LastSyncedAt is the time of the last successful sync, or zero if the
	// device has never synced.
	LastSyncedAt time.Time

	LastSeq int64
}

// Engine syncs the todo files in one directory.
type Engine struct {
	todoDir  string
	stateDir string
	client   Client
}

// New returns an Engine for the todo files in todoDir. Sync state (sync.json
// and base.json) is kept in stateDir.
func New(todoDir, stateDir string, client Client) *Engine {
	return &Engine{todoDir: todoDir, stateDir: stateDir, client: client}
}

// Sync runs one sync cycle while holding a file lock on todoDir. A failed sync
// leaves the files and state unchanged, so the next call retries the same
// changes.
func (e *Engine) Sync(ctx context.Context) error {
	return errNotImplemented
}

// Status reports pending changes by diffing the files against the base
// snapshot. It does no network I/O.
func (e *Engine) Status() (Status, error) {
	return Status{}, errNotImplemented
}
