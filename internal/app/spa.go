package app

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// spaHandler serves the embedded web UI. A request for an existing file gets that
// file; any other path gets index.html so that client-side routes work on reload.
// API and media paths are never answered with index.html.
type spaHandler struct {
	files      fs.FS
	fileServer http.Handler
}

func newSPAHandler(files fs.FS) http.Handler {
	return &spaHandler{files: files, fileServer: http.FileServer(http.FS(files))}
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/hls/") {
		http.NotFound(w, r)
		return
	}

	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if name == "" {
		name = "index.html"
	}

	if info, err := fs.Stat(h.files, name); err == nil && !info.IsDir() {
		h.fileServer.ServeHTTP(w, r)
		return
	}

	// A missing file with an extension (for example /assets/app.js) is an error,
	// not a client route: answer 404 instead of HTML the browser cannot execute.
	if path.Ext(name) != "" {
		http.NotFound(w, r)
		return
	}

	index, err := fs.ReadFile(h.files, "index.html")
	if err != nil {
		// Only the placeholder is embedded: the UI has not been built yet.
		http.Error(w, "web UI is not built; run make web-build", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(index)
}
