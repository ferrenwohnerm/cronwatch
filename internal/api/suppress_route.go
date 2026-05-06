package api

import (
	"log"
	"net/http"
)

// RegisterSuppressRoutes registers suppression management routes on the given mux.
// POST /suppress  — add a suppression window for a job
// GET  /suppress  — list currently active suppressions
func RegisterSuppressRoutes(mux *http.ServeMux, sm *SuppressManager, logger *log.Logger) {
	mux.Handle("/suppress", LoggingMiddleware(logger)(
		RecoveryMiddleware(logger)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost:
					sm.handleSuppress(w, r)
				case http.MethodGet:
					sm.handleSuppressList(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		),
	))
}
