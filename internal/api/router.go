package api

import (
	"io"
	"log"
	"net/http"

	"github.com/user/cronwatch/internal/tracker"
)

// NewRouter builds the HTTP mux for cronwatch, wiring all routes.
func NewRouter(t *tracker.Tracker, h *History, logger *log.Logger, out io.Writer) http.Handler {
	if logger == nil {
		logger = log.New(out, "", log.LstdFlags)
	}

	handler := NewHandler(t)
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/status", handler.handleStatus)
	mux.HandleFunc("/history", handleHistory(h))

	return RecoveryMiddleware(LoggingMiddleware(mux, logger))
}
