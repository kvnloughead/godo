package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	validator "github.com/kvnloughead/godo/internal"
	"github.com/lib/pq"
)

// Todo is a struct representing data for a single todo entry. It is intended to
// be compatible with todo.txt syntax (http://todotxt.org/). How this syntax
// maps to a Todo document will be covered in cmd/cli.
type Todo struct {
	ID        int64             `json:"id"`
	UserID    int64             `json:"user_id"`
	CreatedAt time.Time         `json:"created_at"`
	Title     string            `json:"title"`
	Contexts  []string          `json:"contexts,omitempty"`
	Projects  []string          `json:"projects,omitempty"`
	Priority  rune              `json:"priority"`
	Completed bool              `json:"completed"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Version   int32             `json:"version"`
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
//   - title: if provided, fuzzy matches on the todo's title.
//   - contexts: if provided, only todos that have each of the provided contexts
//     are included.
//   - projects: if provided, only todos that have each of the provided projects
//     are included.
//   - sort: the key to sort by. Prepend with '-' for descending order. Defaults
//     to ID, ascending.
//   - page_size: the number of records to show per "page".
//   - page: the page number to return.
//
// Pagination metadata is returned in the response, unless no records are found.
func (m TodoModel) GetAll(title string, userID int64, contexts []string, projects []string, filters Filters) ([]*Todo, PaginationData, error) {
	// We are using fmt.Sprintf to interpolate column names, since it is not
	// possible to do that with postgresql placeholders.
	query := fmt.Sprintf(`
		SELECT 
			count(*) OVER(),
			id, created_at, title, contexts, projects, priority, completed, metadata, version
		FROM todos
		WHERE (to_tsvector('english', title)
					 @@ plainto_tsquery('english', $1) OR $1 = '')
		AND (contexts @> $2 OR $2 = '{}')
		AND (projects @> $3 OR $3 = '{}')
		AND user_id = $4
		ORDER BY %s %s, id ASC
		LIMIT $5 OFFSET $6`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	// Retrieve matching rows from database.
	args := []any{title, pq.Array(contexts), pq.Array(projects), userID, filters.limit(), filters.offset()}
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
			&m.CreatedAt,
			&m.Title,
			pq.Array(&m.Contexts),
			pq.Array(&m.Projects),
			&m.Priority,
			&m.Completed,
			&m.Metadata,
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
// Todo struct and runs an INSERT query. The id, created_at, and version fields
// are generated automatically.
func (m TodoModel) Insert(todo *Todo) error {
	// The query returns the system-generated id, created_at, and version fields
	// so that we can assign them to the todo struct argument.
	query := `
		INSERT INTO todos (title, user_id, contexts, projects, priority, completed)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, version`

	todo.NilToSlices()

	// The args slice contains the fields provided in the todo struct arguement.
	// Note that we are converting the string slice todo.Contexts to an array the
	// is compatible with the contexts field's text[] type.
	args := []any{todo.Title, todo.UserID, pq.Array(todo.Contexts), pq.Array(todo.Projects), todo.Priority, todo.Completed}

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(
		&todo.ID, &todo.CreatedAt, &todo.Version)
}

// Get retrieves a a specific record in the todos table by its ID. If the ID
// argument is less then 1, or if there is no todo with a matching ID in the
// database, and ErrRecordNotFound is returned. If a todo is found, a pointer
// to the corresponding Todo struct is returned.
func (m TodoModel) Get(id int64) (*Todo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM todos WHERE ID = $1`

	var todo Todo

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&todo.ID,
		&todo.CreatedAt,
		&todo.Title,
		&todo.Year,
		&todo.Runtime,
		pq.Array(&todo.Genres),
		&todo.Version,
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
// The record's version field is incremented by 1 after update.
//
// Prevents edit conflicts by verifying that the version of the record in the
// UPDATE query is the same as the version of the todo argument. In case of
// an edit conflict, an ErrEditConflict error is returned.
func (m TodoModel) Update(todo *Todo) error {
	query := `
		UPDATE todos
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args := []any{
		todo.Title,
		todo.Year,
		todo.Runtime,
		pq.Array(todo.Genres),
		todo.ID,
		todo.Version,
	}

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&todo.Version)
	if err != nil {
		switch {
		// An sql.ErrNoRows is returned if there are no matching records. Since we
		// know that the record exists already, this can be assumed to be due to a
		// version mismatch (hence an edit conflict).
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete deletes a specific record from the todos table. Returns an
// ErrNoRecordFound error if no record is found.
func (m TodoModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM todos WHERE id = $1`

	ctx, cancel := CreateTimeoutContext(QueryTimeout)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows are effected, then there was no record found.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// ValidateTodo validates the fields of a Todo struct. The fields must meet
// the following requirements:
//
//   - Title is the only required field. It contains the full text of the todo.
//
//   - Title must be less than 500 bytes.
//
//   - There can be between 0 and 5 unique, string-valued contexts.
//
//   - There can be between 0 and 5 unique, string-valued projects.
//
//   - There can be a priority, a single character between A and Z, or a 0. A
//     priority of 0 indicates that no priority was specified.
//
//     TODO - validate metadata
func ValidateTodo(v *validator.Validator, m *Todo) {

	v.Check(m.Title != "", "title", "must be provided")
	v.Check(len(m.Title) < 500, "title", "must be less than 500 bytes")

	v.Check(len(m.Contexts) <= 5, "contexts", "must be no more than 5 contexts")
	v.Check(validator.Unique(m.Contexts), "contexts", "must not contain duplicate values")

	v.Check(len(m.Projects) <= 5, "contexts", "must be no more than 5 projects")
	v.Check(validator.Unique(m.Projects), "projects", "must not contain duplicate values")

	v.Check(m.Priority == 0 || (m.Priority >= 'A' && m.Priority <= 'Z'), "priority", "must be a capital letter (A to Z)")
}
