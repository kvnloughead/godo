package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kvnloughead/godo/internal/assert"
	"github.com/kvnloughead/godo/internal/injector"
	"github.com/kvnloughead/godo/internal/server/data"
)

// TestContextualizeRequestLogging tests that successful healthchecks aren't
// logged, while failed healthchecks and other requests are.
func TestContextualizeRequestLogging(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		status      int
		expectedLog bool
	}{
		{name: "Successful healthcheck", path: "/v1/healthcheck", status: http.StatusOK, expectedLog: false},
		{name: "Failed healthcheck", path: "/v1/healthcheck", status: http.StatusInternalServerError, expectedLog: true},
		{name: "Other request", path: "/v1/todos", status: http.StatusOK, expectedLog: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			app := NewAPIApplication(injector.NewApplication(
				injector.Config{Env: "testing"},
				slog.New(slog.NewTextHandler(&logs, nil)),
				nil,
			))

			handler := app.contextualizeRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))

			r := httptest.NewRequest(http.MethodGet, tt.path, nil)
			r = app.contextSetUser(r, data.AnonymousUser)
			handler.ServeHTTP(httptest.NewRecorder(), r)

			assert.Equal(t, strings.Contains(logs.String(), "request completed"), tt.expectedLog)
		})
	}
}
