package api

import (
	"log"
	"net/http"

	"github.com/user/cronwatch/internal/tracker"
)

// NewRouter builds and returns the root HTTP mux with all routes registered
// and middleware applied.
func NewRouter(tr *tracker.Tracker, logger *log.Logger) http.Handler {
	h := NewHandler(tr)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealthz)
	mux.HandleFunc("/status", h.handleStatus)

	var handler http.Handler = mux
	handler = LoggingMiddleware(logger, handler)
	handler = RecoveryMiddleware(logger, handler)

	return handler
}
