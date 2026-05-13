package api

import "net/http"

// RegisterLastSeenRoutes wires the last-seen endpoint into the provided mux.
func RegisterLastSeenRoutes(mux *http.ServeMux, m *LastSeenManager) {
	mux.HandleFunc("/last-seen", m.handleLastSeen)
}
