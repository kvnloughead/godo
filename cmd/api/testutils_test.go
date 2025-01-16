package main

import (
	"log/slog"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kvnloughead/godo/internal/data"
	"github.com/kvnloughead/godo/internal/injector"
)

type testWriter struct {
	t *testing.T
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
}

func newTestApplication(t *testing.T) *APIApplication {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	// The first argument of NewResult (lastInsertId) is set to 0 because we
	// don't need it. Success/failure is determined by the number of rows
	// affected.

	// Valid batch delete (IDs 1, 2, 3)
	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Invalid ID format test (1, invalid, 3)
	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Note: "invalid" ID won't reach the database

	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Not found test (1, 999, 3)
	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))

	logger := slog.New(slog.NewTextHandler(testWriter{t}, nil))

	injectorApp := &injector.Application{
		Config: injector.Config{},
		Logger: logger,
		Models: data.NewModels(db),
	}

	return &APIApplication{
		Application: injectorApp,
	}
}
