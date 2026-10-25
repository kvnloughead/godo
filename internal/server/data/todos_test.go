package data

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/kvnloughead/godo/internal/assert"
)

// testFilters returns the filters that GET /v1/todos uses by default.
func testFilters() Filters {
	return Filters{Page: 1, PageSize: 100, Sort: "id", SortSafelist: []string{"id"}}
}

// TestTodoWritesSetSeq checks that insert, update, and delete each take the
// user's next seq value, bump the version, and set updated_at.
func TestTodoWritesSetSeq(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	user := newTestUser(t, db)
	m := TodoModel{DB: db}

	todo := insertTestTodo(t, m, user.ID, "buy milk")
	if todo.UID == "" {
		t.Error("Insert should generate a uid")
	}
	assert.Equal(t, todo.Seq, int64(1))
	assert.Equal(t, todo.Version, int32(1))
	assert.Equal(t, syncSeqFor(t, db, user.ID), int64(1))

	inserted := todo.UpdatedAt
	todo.Text = "buy oat milk"
	err := m.Update(todo)
	assert.IsNil(t, err)
	assert.Equal(t, todo.Seq, int64(2))
	assert.Equal(t, todo.Version, int32(2))
	if todo.UpdatedAt.Before(inserted) {
		t.Errorf("updated_at went backwards: %v -> %v", inserted, todo.UpdatedAt)
	}
	assert.Equal(t, syncSeqFor(t, db, user.ID), int64(2))

	err = m.Delete(todo)
	assert.IsNil(t, err)
	assert.Equal(t, todo.Seq, int64(3))
	assert.Equal(t, todo.Version, int32(3))
	assert.Equal(t, syncSeqFor(t, db, user.ID), int64(3))

	// The stored row should match what the model reported.
	var seq int64
	var version int32
	err = db.QueryRow(`SELECT seq, version FROM todos WHERE id = $1`, todo.ID).Scan(&seq, &version)
	assert.IsNil(t, err)
	assert.Equal(t, seq, int64(3))
	assert.Equal(t, version, int32(3))
}

// TestTodoInsertKeepsUID checks that a uid supplied by the caller is stored
// as-is, which POST /sync will rely on for client-generated IDs.
func TestTodoInsertKeepsUID(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	user := newTestUser(t, db)
	m := TodoModel{DB: db}

	uid := newUID()
	todo := &Todo{UserID: user.ID, UID: uid, Text: "from a client"}
	err := m.Insert(todo)
	assert.IsNil(t, err)

	got, err := m.GetTodoIfOwned(todo.ID, user.ID)
	assert.IsNil(t, err)
	assert.Equal(t, got.UID, uid)
}

// TestTodoSoftDelete checks that a deleted todo stays in the table as a
// tombstone but is hidden from every read, and can't be written again.
func TestTodoSoftDelete(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	user := newTestUser(t, db)
	m := TodoModel{DB: db}

	keep := insertTestTodo(t, m, user.ID, "keep me")
	gone := insertTestTodo(t, m, user.ID, "delete me")

	err := m.Delete(gone)
	assert.IsNil(t, err)

	// The row is still there, with deleted_at set.
	var deleted bool
	err = db.QueryRow(`SELECT deleted_at IS NOT NULL FROM todos WHERE id = $1`, gone.ID).Scan(&deleted)
	assert.IsNil(t, err)
	assert.Equal(t, deleted, true)

	// Get hides it.
	_, err = m.GetTodoIfOwned(gone.ID, user.ID)
	if !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("GetTodoIfOwned: got %v; want %v", err, ErrRecordNotFound)
	}

	// List and its count hide it, including with archived todos included.
	for _, includeArchived := range []bool{false, true} {
		f := testFilters()
		f.IncludeArchived = includeArchived
		todos, meta, err := m.GetAll("", user.ID, nil, nil, f)
		assert.IsNil(t, err)
		assert.Equal(t, len(todos), 1)
		assert.Equal(t, meta.TotalRecords, 1)
		if len(todos) == 1 {
			assert.Equal(t, todos[0].ID, keep.ID)
		}
	}

	// Writes to a tombstone are conflicts, and don't use up a seq value.
	before := syncSeqFor(t, db, user.ID)

	gone.Text = "resurrect me"
	err = m.Update(gone)
	if !errors.Is(err, ErrEditConflict) {
		t.Errorf("Update after delete: got %v; want %v", err, ErrEditConflict)
	}
	err = m.Delete(gone)
	if !errors.Is(err, ErrEditConflict) {
		t.Errorf("Delete after delete: got %v; want %v", err, ErrEditConflict)
	}

	assert.Equal(t, syncSeqFor(t, db, user.ID), before)
}

// TestTodoDeleteIsOwnerScoped checks that one user can't delete another
// user's todo.
func TestTodoDeleteIsOwnerScoped(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	owner := newTestUser(t, db)
	other := newTestUser(t, db)
	m := TodoModel{DB: db}

	todo := insertTestTodo(t, m, owner.ID, "mine")

	stolen := *todo
	stolen.UserID = other.ID
	err := m.Delete(&stolen)
	if !errors.Is(err, ErrEditConflict) {
		t.Errorf("got %v; want %v", err, ErrEditConflict)
	}

	_, err = m.GetTodoIfOwned(todo.ID, owner.ID)
	assert.IsNil(t, err)
}

// TestTodoUpdateStaleVersion checks that an update with an old version is a
// conflict and doesn't use up a seq value.
func TestTodoUpdateStaleVersion(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	user := newTestUser(t, db)
	m := TodoModel{DB: db}

	todo := insertTestTodo(t, m, user.ID, "v1")
	stale := *todo

	todo.Text = "v2"
	assert.IsNil(t, m.Update(todo))
	before := syncSeqFor(t, db, user.ID)

	stale.Text = "also v2"
	err := m.Update(&stale)
	if !errors.Is(err, ErrEditConflict) {
		t.Errorf("got %v; want %v", err, ErrEditConflict)
	}
	assert.Equal(t, syncSeqFor(t, db, user.ID), before)
}

// TestConcurrentInsertsSeqContiguous runs many inserts for one user at once,
// while a reader repeatedly checks what a syncing client would see. Because
// each insert takes exactly one seq value, every snapshot must contain exactly
// the seq values 1..max. A gap would mean a later seq committed before an
// earlier one, which is the bug the seq lock exists to prevent: a client could
// record last_seq = max and never receive the missing row.
func TestConcurrentInsertsSeqContiguous(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	user := newTestUser(t, db)
	m := TodoModel{DB: db}

	const writers = 50

	var wg sync.WaitGroup
	done := make(chan struct{})
	readerErr := make(chan error, 1)
	snapshots := 0

	go func() {
		defer close(readerErr)
		for {
			select {
			case <-done:
				return
			default:
			}

			var count, distinct, max int64
			err := db.QueryRow(`
				SELECT count(*), count(DISTINCT seq), COALESCE(max(seq), 0)
				FROM todos WHERE user_id = $1`, user.ID).Scan(&count, &distinct, &max)
			if err != nil {
				readerErr <- err
				return
			}
			if count != distinct || count != max {
				readerErr <- fmt.Errorf("snapshot has %d rows, %d distinct seqs, max seq %d", count, distinct, max)
				return
			}
			snapshots++
		}
	}()

	errs := make(chan error, writers)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs <- m.Insert(&Todo{UserID: user.ID, Text: fmt.Sprintf("todo %d", i)})
		}(i)
	}
	wg.Wait()
	close(done)
	close(errs)

	for err := range errs {
		assert.IsNil(t, err)
	}
	assert.IsNil(t, <-readerErr)
	t.Logf("reader checked %d snapshots", snapshots)

	assertSeqsUnique(t, db, user.ID, writers)
}

// TestConcurrentMixedWritesSeqUnique runs inserts, updates, and deletes for
// one user at once, and checks that every successful write got its own seq
// value and that users.sync_seq ends at the highest one.
func TestConcurrentMixedWritesSeqUnique(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	user := newTestUser(t, db)
	other := newTestUser(t, db)
	m := TodoModel{DB: db}

	const seeded = 20
	todos := make([]*Todo, seeded)
	for i := range todos {
		todos[i] = insertTestTodo(t, m, user.ID, fmt.Sprintf("seed %d", i))
	}

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		succeeded int
	)
	record := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, ErrEditConflict):
			// Expected when two goroutines race on the same todo.
		default:
			t.Error(err)
		}
	}

	for i := 0; i < seeded; i++ {
		// Two updates and an insert per seeded todo, plus a delete for every
		// fourth one. Updates and deletes on the same todo race, so some will
		// conflict.
		for j := 0; j < 2; j++ {
			wg.Add(1)
			go func(base Todo, j int) {
				defer wg.Done()
				base.Text = fmt.Sprintf("%s edit %d", base.Text, j)
				record(m.Update(&base))
			}(*todos[i], j)
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			record(m.Insert(&Todo{UserID: user.ID, Text: fmt.Sprintf("new %d", i)}))
		}(i)
		if i%4 == 0 {
			wg.Add(1)
			go func(base Todo) {
				defer wg.Done()
				record(m.Delete(&base))
			}(*todos[i])
		}

		// A write for a different user runs alongside, and must not disturb
		// this user's seq values.
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := m.Insert(&Todo{UserID: other.ID, Text: fmt.Sprintf("other %d", i)})
			if err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()

	// Each successful write took one seq value after the seeding.
	assert.Equal(t, syncSeqFor(t, db, user.ID), int64(seeded+succeeded))
	assert.Equal(t, syncSeqFor(t, db, other.ID), int64(seeded))

	// Rows only keep the seq of their latest write, so there are gaps, but no
	// two rows may share a seq, and none may exceed sync_seq.
	assertSeqsUnique(t, db, user.ID, seeded+seeded)
}

// assertSeqsUnique checks that the user has wantRows todos (including
// tombstones), that no two share a seq value, and that the highest seq equals
// users.sync_seq.
func assertSeqsUnique(t *testing.T, db *sql.DB, userID int64, wantRows int) {
	t.Helper()

	var count, distinct, max int64
	err := db.QueryRow(`
		SELECT count(*), count(DISTINCT seq), COALESCE(max(seq), 0)
		FROM todos WHERE user_id = $1`, userID).Scan(&count, &distinct, &max)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, count, int64(wantRows))
	assert.Equal(t, distinct, count)
	assert.Equal(t, max, syncSeqFor(t, db, userID))
}
