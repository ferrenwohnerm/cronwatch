package api

import "net/http"

// RegisterTimeoutRoutes wires the timeout endpoints onto mux.
func RegisterTimeoutRoutes(mux *http.ServeMux, m *TimeoutManager) {
	mux.HandleFunc("/timeouts", handleTimeouts(m))
}
