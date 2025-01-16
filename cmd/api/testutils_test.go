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
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	// Set up expectations for multiple DELETE operations
	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM todos WHERE id = \\$1").
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))

	logger := slog.New(slog.NewTextHandler(testWriter{t}, nil))

	// Create an injector.Application first
	injectorApp := &injector.Application{
		Config: injector.Config{},
		Logger: logger,
		Models: data.NewModels(db),
	}

	// Return an APIApplication that embeds the injector.Application
	return &APIApplication{
		Application: injectorApp,
	}
}
