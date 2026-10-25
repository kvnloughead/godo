package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kvnloughead/godo/internal/assert"
	"github.com/kvnloughead/godo/internal/server/data"
)

// todoTestClient sends authenticated requests to a test server backed by the
// integration test database.
type todoTestClient struct {
	t     *testing.T
	ts    *httptest.Server
	token string
}

// newTodoTestClient starts a test server and creates an activated user with
// todos:read and todos:write permissions. The test is skipped if
// GODO_TEST_DB_DSN is unset.
func newTodoTestClient(t *testing.T) *todoTestClient {
	t.Helper()

	if os.Getenv(testDBEnv) == "" {
		t.Skipf("%s not set; skipping integration test", testDBEnv)
	}

	app, routes := newTestApp(t)

	user := &data.User{
		Name:      "test",
		Email:     fmt.Sprintf("handlers-%d@example.com", time.Now().UnixNano()),
		Activated: true,
	}
	err := user.Password.Set("pa55word1234")
	if err != nil {
		t.Fatal(err)
	}
	err = app.Models.Users.Insert(user)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		app.Models.Users.DB.Exec(`DELETE FROM users WHERE id = $1`, user.ID)
	})

	err = app.Models.Permissions.AddForUser(user.ID, data.TodosRead, data.TodosWrite)
	if err != nil {
		t.Fatal(err)
	}

	token, err := app.Models.Tokens.New(user.ID, time.Hour, data.Authentication)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(routes)
	t.Cleanup(ts.Close)

	return &todoTestClient{t: t, ts: ts, token: token.Plaintext}
}

// do sends a request and returns the status code and decoded JSON body. If
// ifMatch is not empty, it's sent as the If-Match header.
func (c *todoTestClient) do(method, path, body, ifMatch string) (int, map[string]json.RawMessage) {
	c.t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, c.ts.URL+path, reader)
	if err != nil {
		c.t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if ifMatch != "" {
		req.Header.Set("If-Match", ifMatch)
	}

	rs, err := c.ts.Client().Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer rs.Body.Close()

	var env map[string]json.RawMessage
	err = json.NewDecoder(rs.Body).Decode(&env)
	if err != nil {
		c.t.Fatal(err)
	}

	return rs.StatusCode, env
}

// todoFrom decodes the "todo" field of a response envelope.
func (c *todoTestClient) todoFrom(env map[string]json.RawMessage) data.Todo {
	c.t.Helper()

	var todo data.Todo
	err := json.Unmarshal(env["todo"], &todo)
	if err != nil {
		c.t.Fatalf("decoding todo from %v: %v", env, err)
	}
	return todo
}

// TestTodoIfMatch checks that PATCH and DELETE honor an optional If-Match
// header, and that DELETE leaves a tombstone that reads no longer return.
func TestTodoIfMatch(t *testing.T) {
	c := newTodoTestClient(t)

	status, env := c.do(http.MethodPost, "/v1/todos", `{"text": "buy milk"}`, "")
	assert.Equal(t, status, http.StatusCreated)
	created := c.todoFrom(env)
	assert.Equal(t, created.Version, int32(1))
	if created.UID == "" {
		t.Error("response should include a uid")
	}

	path := fmt.Sprintf("/v1/todos/%d", created.ID)

	tests := []struct {
		name        string
		method      string
		body        string
		ifMatch     string
		wantStatus  int
		wantVersion int32 // checked when the request succeeds
	}{
		{"PATCH stale If-Match", http.MethodPatch, `{"completed": true}`, "2", http.StatusConflict, 0},
		{"PATCH malformed If-Match", http.MethodPatch, `{"completed": true}`, "abc", http.StatusBadRequest, 0},
		{"PATCH matching If-Match", http.MethodPatch, `{"completed": true}`, "1", http.StatusOK, 2},
		{"PATCH quoted If-Match", http.MethodPatch, `{"priority": "A"}`, `"2"`, http.StatusOK, 3},
		{"PATCH without If-Match", http.MethodPatch, `{"text": "buy oat milk"}`, "", http.StatusOK, 4},
		{"PATCH If-Match *", http.MethodPatch, `{"completed": false}`, "*", http.StatusOK, 5},
		{"DELETE stale If-Match", http.MethodDelete, "", "4", http.StatusConflict, 0},
		{"DELETE matching If-Match", http.MethodDelete, "", "5", http.StatusOK, 0},
		{"GET after DELETE", http.MethodGet, "", "", http.StatusNotFound, 0},
		{"PATCH after DELETE", http.MethodPatch, `{"completed": true}`, "", http.StatusNotFound, 0},
		{"DELETE after DELETE", http.MethodDelete, "", "", http.StatusNotFound, 0},
	}

	// The cases share one todo, so they run in order rather than in parallel.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, env := c.do(tt.method, path, tt.body, tt.ifMatch)
			assert.Equal(t, status, tt.wantStatus)
			if tt.wantVersion != 0 {
				assert.Equal(t, c.todoFrom(env).Version, tt.wantVersion)
			}
		})
	}

	// The deleted todo is gone from the list, too.
	status, env = c.do(http.MethodGet, "/v1/todos", "", "")
	assert.Equal(t, status, http.StatusOK)
	var todos []data.Todo
	err := json.Unmarshal(env["todos"], &todos)
	assert.IsNil(t, err)
	assert.Equal(t, len(todos), 0)
}

// TestDeleteOtherUsersTodo checks that a user can't delete another user's
// todo by guessing its ID.
func TestDeleteOtherUsersTodo(t *testing.T) {
	owner := newTodoTestClient(t)
	other := newTodoTestClient(t)

	status, env := owner.do(http.MethodPost, "/v1/todos", `{"text": "mine"}`, "")
	assert.Equal(t, status, http.StatusCreated)
	path := fmt.Sprintf("/v1/todos/%d", owner.todoFrom(env).ID)

	status, _ = other.do(http.MethodDelete, path, "", "")
	assert.Equal(t, status, http.StatusNotFound)

	status, _ = owner.do(http.MethodGet, path, "", "")
	assert.Equal(t, status, http.StatusOK)
}
