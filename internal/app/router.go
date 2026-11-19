package app

import (
	"fmt"
	"net/http"
)

// registerPublicRoutes registers routes that do not require a session.
func registerPublicRoutes(mux *http.ServeMux, a *App) {
	mux.HandleFunc("POST /api/auth/login", a.authManager.HandleBasicLogin)
	mux.HandleFunc("POST /api/auth/logout", handleLogout)
}

// registerRoutes registers the API and media routes. They are wrapped in
// AuthMiddleware by Init; the web UI is served separately (see spa.go).
// Handlers come from internal/api; the paths match the previous binary.
func registerRoutes(mux *http.ServeMux, a *App) {
	apiServer := a.apiServer

	// API routes
	mux.HandleFunc("GET /api/cameras", apiServer.HandleGetCameras)
	mux.HandleFunc("POST /api/cameras", apiServer.HandleAddCamera)
	mux.HandleFunc("GET /api/cameras/{id}/status", apiServer.HandleCameraStatus)
	mux.HandleFunc("GET /api/config/{type}", apiServer.HandleGetConfig)
	mux.HandleFunc("POST /api/config/{type}", apiServer.HandleUpdateConfig)
	mux.HandleFunc("GET /api/archive", apiServer.HandleTimeline)
	mux.HandleFunc("GET /api/events", apiServer.HandleGetEvents)
	mux.HandleFunc("GET /api/diagnostics", apiServer.HandleDiagnostics)

	// HLS routes
	mux.HandleFunc("GET /hls/{camera}/playlist.m3u8", a.hlsHandler.HandlePlaylist)
	mux.HandleFunc("GET /hls/{camera}/{file}", a.hlsHandler.HandleSegment)
}

// handleLogout returns the same fixed response as the previous binary.
func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprint(w, `{"status": "logged out"}`)
}
