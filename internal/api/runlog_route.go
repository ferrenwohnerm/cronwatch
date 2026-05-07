package api

import "net/http"

// RegisterRunLogRoutes registers the /runlog endpoint on the given mux.
func RegisterRunLogRoutes(mux *http.ServeMux, mgr *RunLogManager) {
	mux.HandleFunc("/runlog", mgr.handleRunLog)
}
