package api

import "net/http"

// RegisterRunLogRoutes registers the /runlog endpoint on the given mux.
// It supports the following HTTP methods:
//   - GET  /runlog        - list recent run log entries
//   - POST /runlog        - create a new run log entry
func RegisterRunLogRoutes(mux *http.ServeMux, mgr *RunLogManager) {
	mux.HandleFunc("/runlog", mgr.handleRunLog)
	mux.HandleFunc("/runlog/", mgr.handleRunLogByID)
}
