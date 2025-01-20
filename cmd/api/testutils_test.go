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

func newTestApplication(t *testing.T, testName string) *APIApplication {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	switch testName {
	case "Valid batch delete":
		// Expect all deletes to succeed
		mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
			WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
			WithArgs(3).WillReturnResult(sqlmock.NewResult(0, 1))

	case "Invalid ID format":
		// Expect successful deletes for valid IDs (1 and 3)
		// "invalid ID" won't reach the database
		mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
			WithArgs(3).WillReturnResult(sqlmock.NewResult(0, 1))

	case "Not found ID":
		// Expect successful delete for IDs 1 and 3 and not found for 999.
		mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
			WithArgs(999).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
			WithArgs(3).WillReturnResult(sqlmock.NewResult(0, 1))
	}

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
