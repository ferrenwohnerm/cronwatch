package api

import (
	"io"
	"log"
	"net/http"
)

// NewRouter constructs the HTTP mux wiring all API routes together.
// It accepts a Handler (status/healthz), a History recorder, and a Metrics
// recorder so each subsystem registers its own route.
func NewRouter(h *Handler, hist *History, m *Metrics, logger *log.Logger, out io.Writer) http.Handler {
	if logger == nil {
		logger = log.New(out, "", log.LstdFlags)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/status", h.handleStatus)
	mux.HandleFunc("/history", hist.handleHistory)
	mux.HandleFunc("/metrics", m.handleMetrics)

	return RecoveryMiddleware(LoggingMiddleware(mux, logger))
}
