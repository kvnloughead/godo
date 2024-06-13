package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kvnloughead/godo/internal/assert"
	"github.com/kvnloughead/godo/internal/injector"
)

func TestHealthcheck(t *testing.T) {

	baseApp := injector.NewApplication(
		injector.Config{Env: "testing"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
	)

	app := NewAPIApplication(baseApp)

	ts := httptest.NewServer(app.Routes())
	defer ts.Close()

	rs, err := ts.Client().Get(ts.URL + "/v1/healthcheck")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, rs.StatusCode, http.StatusOK)

	var response struct {
		Status     string `json:"status"`
		SystemInfo struct {
			Environment string `json:"environment"`
			Version     string `json:"version"`
		} `json:"system_info"`
	}

	defer rs.Body.Close()
	err = json.NewDecoder(rs.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, response.Status, "available")
	assert.Equal(t, response.SystemInfo.Environment, "testing")
	assert.Equal(t, response.SystemInfo.Version, version)
}
