package data

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"time"

	validator "github.com/kvnloughead/godo/internal/validator"
	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
)

// Todo is a struct representing data for a single todo entry. It is intended to
// be compatible with todo.txt syntax (http://todotxt.org/). How this syntax
// maps to a Todo document will be covered in cmd/godo.
//
// UID is the todo's stable public identifier (a ULID for todos created after
// migration 000006). ID is the internal database key, and is still used in
// routes for now. Seq is the user's sync seq as of the todo's last write, and
// is internal to the server.
type Todo struct {
	ID        int64     `json:"id"`
	UID       string    `json:"uid"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Text      string    `json:"text"`
	Contexts  []string  `json:"contexts,omitempty"`
	Projects  []string  `json:"projects,omitempty"`
	Priority  string    `json:"priority"`
	Completed bool      `json:"completed"`
	Version   int32     `json:"version"`
	Archived  bool      `json:"archived"`
	Seq       int64     `json:"-"`
}

// newUID returns a new ULID string for use as a todo's UID.
func newUID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

// NilToSlices converts the calling structs Contexts and Projects fields to
// empty slices if they are nil. This allows them to be inserted into
// non-nullable Postrgresql fields.
func (t *Todo) NilToSlices() {
	if t.Contexts == nil {
		t.Contexts = []string{}
	}
	if t.Projects == nil {
		t.Projects = []string{}
	}
}

// TodoModel struct wraps an sql.DB connection pool and implements
// basic CRUD operations.
type TodoModel struct {
	DB *sql.DB
}

// GetAll retrieves a slice of todos from the database. The slice can be
// filtered, sorted, and paginated via several optional query parameters.
//
//   - text: if provided, fuzzy matches on the todo's text.
//   - archived status: by default, only unarchived todos are included. The
//     query params "include-archived" and "only-archived" can be used to
//     override this behavior.
//   - contexts: if provided, only todos that have each of the provided contexts
//     are included.
//   - projects: if provided, only todos that have each of the provided projects
//     are included.
//   - sort: the key to sort by. Prepend with '-' for descending order. Defaults
//     to ID, ascending.
//   - page_size: the number of records to show per "page".
//   - page: the page number to return.
//
// Soft-deleted todos are never included.
//
// Pagination metadata is returned in the response, unless no records are found.
func (m TodoModel) GetAll(text string, userID int64, contexts []string, projects []string, filters Filters) ([]*Todo, PaginationData, error) {
	whereClause := `WHERE text ILIKE '%%' || $1 || '%%' AND user_id = $2 AND deleted_at IS NULL`

	// Handle archived/unarchived filtering.
	if filters.OnlyArchived {
		whereClause += " AND archived = true" // show only archived
	} else if !filters.IncludeArchived {
		whereClause += " AND archived = false" // show only unarchived (default)
	}

	// Handle completion status filtering
	if filters.Done {
		whereClause += " AND completed = true"
	} else if filters.Undone {
		whereClause += " AND completed = false"
	}

	query := fmt.Sprintf(` 
		SELECT 
			count(*) OVER(),
			id, uid, created_at, updated_at, text, contexts, projects, priority, completed, archived, version
		FROM todos
		%s
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, whereClause, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	args := []any{text, userID, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, PaginationData{}, err
	}
	defer rows.Close() // Defer closing after handling errors.

	// totalRecords will receive the number of records returned by the query
	// (i.e., the value of count(*) OVER()).
	totalRecords := 0
	todos := []*Todo{}

	// Iterate through rows, reading each record in an entry in a Todo slice.
	for rows.Next() {
		var m Todo
		err = rows.Scan(
			&totalRecords,
			&m.ID,
			&m.UID,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.Text,
			pq.Array(&m.Contexts),
			pq.Array(&m.Projects),
			&m.Priority,
			&m.Completed,
			&m.Archived,
			&m.Version,
		)
		if err != nil {
			return nil, PaginationData{}, err
		}
		todos = append(todos, &m)
	}

	// rows.Err() will contain any errors that occurred during iteration.
	err = rows.Err()
	if err != nil {
		return nil, PaginationData{}, err
	}

	paginationData := calculatePaginationData(totalRecords, filters.Page, filters.PageSize)
	return todos, paginationData, nil
}

// Insert adds a new record to the todo table. It accepts a pointer to a
// Todo struct and runs an INSERT query. The id, created_at, updated_at, and
// version fields are generated automatically. If todo.UID is empty, a new
// ULID is generated for it.
//
// The insert runs in a transaction that holds the user's seq lock (see
// SeqLock), and the new row's seq is set to the user's next seq value.
func (m TodoModel) Insert(todo *Todo) error {
	// The query returns the system-generated id, created_at, updated_at, and
	// version fields so that we can assign them to the todo struct argument.
	query := `
		INSERT INTO todos (uid, text, user_id, contexts, projects, priority, completed, archived, seq, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id, created_at, updated_at, version`

	todo.NilToSlices()
	if todo.UID == "" {
		todo.UID = newUID()
	}

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	return WithSeqLock(ctx, m.DB, todo.UserID, func(tx *sql.Tx, seq *SeqLock) error {
		todo.Seq = seq.Next()

		// The args slice contains the fields provided in the todo struct
		// arguement. Note that we are converting the string slice todo.Contexts to
		// an array the is compatible with the contexts field's text[] type.
		args := []any{todo.UID, todo.Text, todo.UserID, pq.Array(todo.Contexts), pq.Array(todo.Projects), todo.Priority, todo.Completed, todo.Archived, todo.Seq}

		return tx.QueryRowContext(ctx, query, args...).Scan(
			&todo.ID, &todo.CreatedAt, &todo.UpdatedAt, &todo.Version)
	})
}

// GetTodoIfOwned retrieves a a specific record in the todos table by its ID, but only if the current user owns the todo item.
//
// An ErrRecordNotFound is returned in the following cases:
//
//   - If either the id or userID arguments are less then 1
//   - If there is no todo with matching id and userID
//   - If the todo has been soft-deleted
//
// If a todo is found, a pointer to the corresponding Todo struct is returned.
func (m TodoModel) GetTodoIfOwned(id, userID int64) (*Todo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, uid, user_id, created_at, updated_at, text, contexts, projects, priority, completed, archived, version, seq
		FROM todos WHERE ID = $1 AND user_id = $2 AND deleted_at IS NULL`

	var todo Todo

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id, userID).Scan(
		&todo.ID,
		&todo.UID,
		&todo.UserID,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.Text,
		pq.Array(&todo.Contexts),
		pq.Array(&todo.Projects),
		&todo.Priority,
		&todo.Completed,
		&todo.Archived,
		&todo.Version,
		&todo.Seq,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &todo, nil
}

// Update updates a specific record in the todos table. The caller should
// check for the existence of the record to be updated before calling Update.
// The record's version field is incremented by 1 after update, updated_at is
// set to the current time, and seq is set to the user's next seq value.
//
// Prevents edit conflicts by verifying that the version of the record in the
// UPDATE query is the same as the version of the todo argument. In case of
// an edit conflict, an ErrEditConflict error is returned.
func (m TodoModel) Update(todo *Todo) error {
	query := `
		UPDATE todos
		SET text = $1, contexts = $2, projects = $3, priority = $4, completed = $5, archived = $6,
			version = version + 1, seq = $7, updated_at = NOW()
		WHERE id = $8 AND user_id = $9 AND version = $10 AND deleted_at IS NULL
		RETURNING version, updated_at`

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	return WithSeqLock(ctx, m.DB, todo.UserID, func(tx *sql.Tx, seq *SeqLock) error {
		newSeq := seq.Next()

		args := []any{
			todo.Text,
			pq.Array(todo.Contexts),
			pq.Array(todo.Projects),
			todo.Priority,
			todo.Completed,
			todo.Archived,
			newSeq,
			todo.ID,
			todo.UserID,
			todo.Version,
		}

		return writeVersioned(ctx, tx, query, args, todo, newSeq)
	})
}

// Delete soft-deletes a specific record in the todos table by setting its
// deleted_at field. The row is kept as a tombstone so that the delete can
// reach other devices when they sync. Like Update, it increments the record's
// version, sets updated_at, and sets seq to the user's next seq value.
//
// The caller should check for the existence of the record before calling
// Delete. As with Update, if the record's version no longer matches the
// version of the todo argument, or the record has already been deleted, an
// ErrEditConflict error is returned.
func (m TodoModel) Delete(todo *Todo) error {
	query := `
		UPDATE todos
		SET deleted_at = NOW(), version = version + 1, seq = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3 AND version = $4 AND deleted_at IS NULL
		RETURNING version, updated_at`

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	return WithSeqLock(ctx, m.DB, todo.UserID, func(tx *sql.Tx, seq *SeqLock) error {
		newSeq := seq.Next()
		args := []any{newSeq, todo.ID, todo.UserID, todo.Version}
		return writeVersioned(ctx, tx, query, args, todo, newSeq)
	})
}

// writeVersioned runs an UPDATE query that is guarded by a version check and
// returns the new version and updated_at fields. On success, those fields and
// todo.Seq are updated.
//
// An sql.ErrNoRows is returned by the query if there are no matching records.
// Since the caller has already checked that the record exists, this can be
// assumed to be due to a version mismatch or a concurrent delete (hence an edit
// conflict).
func writeVersioned(ctx context.Context, tx *sql.Tx, query string, args []any, todo *Todo, newSeq int64) error {
	err := tx.QueryRowContext(ctx, query, args...).Scan(&todo.Version, &todo.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	todo.Seq = newSeq
	return nil
}

// ValidateTodo validates the fields of a Todo struct. The fields must meet
// the following requirements:
//
//   - Text is the only required field. It contains the full text of the todo.
//
//   - Text must be less than 500 bytes.
//
//   - There can be between 0 and 5 unique, string-valued contexts.
//
//   - There can be between 0 and 5 unique, string-valued projects.
//
//   - There can be a priority, a single character between A and Z, or an empty
//     string.
//
//   - Archived and Completed must be booleans.
func ValidateTodo(v *validator.Validator, t *Todo) {

	v.Check(t.Text != "", "text", "must be provided")
	v.Check(len(t.Text) < 500, "text", "must be less than 500 bytes")

	v.Check(len(t.Contexts) <= 5, "contexts", "must be no more than 5 contexts")
	v.Check(validator.Unique(t.Contexts), "contexts", "must not contain duplicate values")

	v.Check(len(t.Projects) <= 5, "contexts", "must be no more than 5 projects")
	v.Check(validator.Unique(t.Projects), "projects", "must not contain duplicate values")

	v.Check(priorityIsValid(t), "priority", "must be a capital letter (A to Z) or empty string")

	v.Check(reflect.TypeOf(t.Archived).Kind() == reflect.Bool, "archived", "must be boolean")
	v.Check(reflect.TypeOf(t.Completed).Kind() == reflect.Bool, "completed", "must be boolean")
}

// priorityIsValid returns true if the todo item's priority field is valid.
// Valid options are letters from A to Z and the empty string.
func priorityIsValid(t *Todo) bool {
	if len(t.Priority) == 0 {
		return true
	}

	if len(t.Priority) != 1 {
		return false
	}

	matched, _ := regexp.MatchString("^[A-Z]$", t.Priority)

	return matched
}
