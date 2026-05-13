package api

import "net/http"

// RegisterJobGroupRoutes registers job group endpoints on the given mux.
func RegisterJobGroupRoutes(mux *http.ServeMux, mgr *JobGroupManager) {
	mux.Handle("/api/v1/groups", mgr)
}
