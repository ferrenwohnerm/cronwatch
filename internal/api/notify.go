package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/cronwatch/internal/alert"
	"github.com/example/cronwatch/internal/tracker"
)

// NotifyHandler handles manual drift-check trigger requests via HTTP.
type NotifyHandler struct {
	tracker *tracker.Tracker
	alerts  *alert.Manager
}

// NewNotifyHandler creates a NotifyHandler.
func NewNotifyHandler(t *tracker.Tracker, m *alert.Manager) *NotifyHandler {
	return &NotifyHandler{tracker: t, alerts: m}
}

type notifyRequest struct {
	JobName string        `json:"job_name"`
	Actual  time.Duration `json:"actual_ms"`
}

type notifyResponse struct {
	Triggered bool   `json:"triggered"`
	Message   string `json:"message"`
}

func (h *NotifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req notifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.JobName == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	entry, ok := h.tracker.Get(req.JobName)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	a := alert.NewDriftAlert(req.JobName, req.Actual, entry.Expected)
	if err := h.alerts.Send(r.Context(), a); err != nil {
		http.Error(w, "failed to send alert: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifyResponse{Triggered: true, Message: "alert dispatched"})
}
