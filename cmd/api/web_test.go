package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/kvnloughead/godo/internal/assert"
	"github.com/kvnloughead/godo/internal/injector"
)

// TestFrontendHandler tests that the web app's files are served, unknown paths
// fall back to index.html, and API paths still get the JSON 404.
func TestFrontendHandler(t *testing.T) {
	dist := fstest.MapFS{
		"index.html":       {Data: []byte("index")},
		"favicon.svg":      {Data: []byte("favicon")},
		"assets/app-1a.js": {Data: []byte("js")},
	}

	app := NewAPIApplication(injector.NewApplication(
		injector.Config{Env: "testing"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	))
	handler := app.frontendHandler(dist)

	tests := []struct {
		name         string
		method       string
		path         string
		status       int
		body         string
		cacheControl string
	}{
		{name: "Root", method: http.MethodGet, path: "/", status: http.StatusOK, body: "index", cacheControl: "no-cache"},
		{name: "Static file", method: http.MethodGet, path: "/favicon.svg", status: http.StatusOK, body: "favicon", cacheControl: "no-cache"},
		{name: "Hashed asset", method: http.MethodGet, path: "/assets/app-1a.js", status: http.StatusOK, body: "js", cacheControl: "public, max-age=31536000, immutable"},
		{name: "Client-side route", method: http.MethodGet, path: "/settings", status: http.StatusOK, body: "index", cacheControl: "no-cache"},
		{name: "Directory", method: http.MethodGet, path: "/assets/", status: http.StatusOK, body: "index", cacheControl: "no-cache"},
		{name: "Unknown API path", method: http.MethodGet, path: "/v1/nope", status: http.StatusNotFound},
		{name: "Debug path", method: http.MethodGet, path: "/debug/nope", status: http.StatusNotFound},
		{name: "Non-GET", method: http.MethodPost, path: "/settings", status: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(tt.method, tt.path, nil))

			assert.Equal(t, rr.Code, tt.status)
			if tt.status == http.StatusOK {
				assert.Equal(t, rr.Body.String(), tt.body)
				assert.Equal(t, rr.Header().Get("Cache-Control"), tt.cacheControl)
			} else {
				assert.StringContains(t, rr.Header().Get("Content-Type"), "application/json")
			}
		})
	}
}

// TestFrontendHandlerWithoutBuild tests that a binary built without the web
// app sends the JSON 404 for every unmatched path.
func TestFrontendHandlerWithoutBuild(t *testing.T) {
	app := NewAPIApplication(injector.NewApplication(
		injector.Config{Env: "testing"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	))

	rr := httptest.NewRecorder()
	app.frontendHandler(nil).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, rr.Code, http.StatusNotFound)
}
