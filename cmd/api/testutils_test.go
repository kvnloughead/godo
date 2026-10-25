package main

import (
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/kvnloughead/godo/internal/injector"
)

// testDBEnv is the environment variable holding the DSN of a migrated
// Postgres database for integration tests. See docs/api/TESTING.md.
const testDBEnv = "GODO_TEST_DB_DSN"

var (
	testAppOnce    sync.Once
	testApp        *APIApplication
	testAppHandler http.Handler
	testAppErr     error
)

// newTestApp returns an application and its routes for use in tests. If
// GODO_TEST_DB_DSN is set, the application is connected to that database.
// Otherwise its database is nil, and only routes that don't use it will work.
//
// Every test shares one application, because app.Routes() can only be called
// once per process: the middleware it builds publishes expvar metrics, and
// expvar panics if a name is published twice.
func newTestApp(t *testing.T) (*APIApplication, http.Handler) {
	t.Helper()

	testAppOnce.Do(func() {
		var db *sql.DB
		if dsn := os.Getenv(testDBEnv); dsn != "" {
			db, testAppErr = sql.Open("postgres", dsn)
			if testAppErr != nil {
				return
			}
		}

		testApp = NewAPIApplication(injector.NewApplication(
			injector.Config{Env: "testing"},
			slog.New(slog.NewTextHandler(io.Discard, nil)),
			db,
		))
		testAppHandler = testApp.Routes()
	})

	if testAppErr != nil {
		t.Fatal(testAppErr)
	}

	return testApp, testAppHandler
}
