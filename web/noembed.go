//go:build !embedweb

// Package web holds the GoDo frontend (a Vue app; see web/README.md) and
// embeds its build output for the API to serve.
//
// The files are embedded only when building with -tags embedweb, after
// running npm run build. Without the tag, Dist returns nil and the API serves
// no frontend, so go build ./... works on a checkout with no web/dist.
package web

import "io/fs"

// Dist returns nil, because this binary was built without -tags embedweb.
func Dist() fs.FS {
	return nil
}
