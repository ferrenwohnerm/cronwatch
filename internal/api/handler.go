package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwatch/internal/tracker"
)

// JobStatusResponse represents the JSON response for a job's current status.
type JobStatusResponse struct {
	JobName   string     `json:"job_name"`
	Running   bool       `json:"running"`
	LastStart *time.Time `json:"last_start,omitempty"`
	LastEnd   *time.Time `json:"last_end,omitempty"`
	LastDrift *float64   `json:"last_drift_seconds,omitempty"`
}

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	tracker *tracker.Tracker
}

// NewHandler creates a new Handler with the given tracker.
func NewHandler(t *tracker.Tracker) *Handler {
	return &Handler{tracker: t}
}

// RegisterRoutes registers all API routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/status", h.handleStatus)
	mux.HandleFunc("/healthz", handleHealthz)
}

// handleStatus returns the status of all tracked jobs.
func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snapshot := h.tracker.Snapshot()
	responses := make([]JobStatusResponse, 0, len(snapshot))

	for name, entry := range snapshot {
		resp := JobStatusResponse{
			JobName: name,
			Running: entry.Running,
		}
		if !entry.StartTime.IsZero() {
			t := entry.StartTime
			resp.LastStart = &t
		}
		if !entry.EndTime.IsZero() {
			t := entry.EndTime
			resp.LastEnd = &t
		}
		if entry.LastDuration > 0 {
			d := entry.LastDuration.Seconds()
			resp.LastDrift = &d
		}
		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// handleHealthz returns a simple liveness check.
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
