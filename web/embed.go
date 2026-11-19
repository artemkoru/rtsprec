// Package web embeds the production build of the web UI.
//
// dist/ holds the output of `npm run build` (see `make web-build`). In a clean
// checkout only the committed placeholder dist/.gitkeep exists, so the Go build
// works without Node; the UI routes then return 404 until the frontend is built.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Dist returns the built UI files rooted at the dist directory.
func Dist() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
