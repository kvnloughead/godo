package data

import (
	"database/sql"
	"os"
	"strings"
	"testing"
)

// testDBEnv is the environment variable holding the DSN of a migrated
// Postgres database for integration tests. See docs/api/TESTING.md.
const testDBEnv = "GODO_TEST_DB_DSN"

// newTestDB opens a connection pool to the integration test database. If
// GODO_TEST_DB_DSN is unset, the test is skipped. The pool is closed when the
// test finishes.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv(testDBEnv)
	if dsn == "" {
		t.Skipf("%s not set; skipping integration test", testDBEnv)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	// The concurrency tests start dozens of goroutines. Capping the pool keeps
	// the total number of connections well under Postgres's default limit of
	// 100 when several tests and packages run at once.
	db.SetMaxOpenConns(20)

	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

// newTestUser inserts an activated user with a unique email, so tests can run
// in parallel against the same database without seeing each other's todos.
// The user, and so their todos, are deleted when the test finishes.
func newTestUser(t *testing.T, db *sql.DB) *User {
	t.Helper()

	user := &User{
		Name:      "test",
		Email:     strings.ToLower(newUID()) + "@example.com",
		Activated: true,
	}
	err := user.Password.Set("pa55word1234")
	if err != nil {
		t.Fatal(err)
	}

	err = UserModel{DB: db}.Insert(user)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Exec(`DELETE FROM users WHERE id = $1`, user.ID)
	})

	return user
}

// syncSeqFor returns the user's current users.sync_seq value.
func syncSeqFor(t *testing.T, db *sql.DB, userID int64) int64 {
	t.Helper()

	var seq int64
	err := db.QueryRow(`SELECT sync_seq FROM users WHERE id = $1`, userID).Scan(&seq)
	if err != nil {
		t.Fatal(err)
	}
	return seq
}

// insertTestTodo inserts a todo with the given text for the user.
func insertTestTodo(t *testing.T, m TodoModel, userID int64, text string) *Todo {
	t.Helper()

	todo := &Todo{UserID: userID, Text: text}
	err := m.Insert(todo)
	if err != nil {
		t.Fatal(err)
	}
	return todo
}
