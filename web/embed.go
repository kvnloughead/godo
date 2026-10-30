//go:build embedweb

package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Dist returns the built frontend, rooted at web/dist.
func Dist() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err) // unreachable: "dist" is a valid path
	}
	return sub
}
