package api

import "net/http"

// RegisterPauseRoutes registers the pause and resume endpoints on the given mux.
// POST /pause  — pause monitoring for a job
// POST /resume — resume monitoring for a job
func RegisterPauseRoutes(mux *http.ServeMux, pm *PauseManager) {
	mux.HandleFunc("/pause", pm.handlePause)
	mux.HandleFunc("/resume", pm.handleResume)
}
