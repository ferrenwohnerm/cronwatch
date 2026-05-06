package api

import (
	"io"
	"log"
	"net/http"

	"github.com/example/cronwatch/internal/alert"
	"github.com/example/cronwatch/internal/tracker"
)

// NewRouter wires all API routes and returns a ready http.Handler.
func NewRouter(t *tracker.Tracker, m *alert.Manager, logger *log.Logger, out io.Writer) http.Handler {
	if logger == nil {
		logger = log.New(out, "", log.LstdFlags)
	}

	mux := http.NewServeMux()

	h := NewHandler(t)
	mux.Handle("/healthz", h)
	mux.Handle("/status", h)

	hist := NewHistory(64)
	mux.Handle("/history", hist)

	met := NewMetrics()
	mux.Handle("/metrics", met)

	nfy := NewNotifyHandler(t, m)
	mux.Handle("/notify", nfy)

	return RecoveryMiddleware(LoggingMiddleware(mux, logger))
}
