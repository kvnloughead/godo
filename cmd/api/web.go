package main

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// frontendHandler serves the built web app from dist. It is the router's
// NotFound handler, so it only sees requests that no API route matched.
//
// API paths (/v1/, /debug/) and non-GET requests get the usual JSON 404. Any
// other path is served from dist if the file exists. Otherwise index.html is
// served, so client-side routes like /settings load the app. If dist is nil
// (the binary was built without -tags embedweb), every request gets the JSON
// 404.
func (app *APIApplication) frontendHandler(dist fs.FS) http.Handler {
	if dist == nil {
		return http.HandlerFunc(app.notFoundResponse)
	}

	fileServer := http.FileServerFS(dist)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead ||
			strings.HasPrefix(r.URL.Path, "/v1/") ||
			strings.HasPrefix(r.URL.Path, "/debug/") {
			app.notFoundResponse(w, r)
			return
		}

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if info, err := fs.Stat(dist, name); err != nil || info.IsDir() {
			// Unknown path: let the app's client-side router handle it. This
			// also keeps the file server from listing directories.
			r = r.Clone(r.Context())
			r.URL.Path = "/"
			name = "index.html"
		}

		// Vite puts content hashes in the names of everything under assets/,
		// so those files never change. index.html must always be revalidated
		// so a deploy is picked up.
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}

		fileServer.ServeHTTP(w, r)
	})
}
