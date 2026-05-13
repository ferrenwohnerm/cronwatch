package api

import "net/http"

// RegisterStreakRoutes attaches streak endpoints to the given mux.
func RegisterStreakRoutes(mux *http.ServeMux, sm *StreakManager) {
	mux.HandleFunc("/streaks", sm.handleStreaks)
}
