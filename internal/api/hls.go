package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/artemkoru/rtsprec/internal/hls"
)

type HLSHandler struct {
	generators map[string]*hls.HLSGenerator
}

func NewHLSHandler() *HLSHandler {
	return &HLSHandler{
		generators: make(map[string]*hls.HLSGenerator),
	}
}

func (h *HLSHandler) RegisterGenerator(cameraID string, gen *hls.HLSGenerator) {
	h.generators[cameraID] = gen
}

func (h *HLSHandler) HandlePlaylist(w http.ResponseWriter, r *http.Request) {
	cameraID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/hls/"), "/playlist.m3u8")

	gen, ok := h.generators[cameraID]
	if !ok {
		http.Error(w, "camera not found", http.StatusNotFound)
		return
	}

	playlist, err := gen.GeneratePlaylist()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate playlist: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(playlist))
}

func (h *HLSHandler) HandleSegment(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/hls/"), "/")
	if len(parts) < 2 {
		http.Error(w, "invalid segment path", http.StatusBadRequest)
		return
	}

	cameraID := parts[0]
	segmentFilename := parts[1]

	gen, ok := h.generators[cameraID]
	if !ok {
		http.Error(w, "camera not found", http.StatusNotFound)
		return
	}

	data, err := gen.GetSegmentFile(segmentFilename)
	if err != nil {
		http.Error(w, fmt.Sprintf("segment not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
