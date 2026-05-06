package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// DriftEvent records a single drift alert occurrence.
type DriftEvent struct {
	JobName   string        `json:"job_name"`
	OccurredAt time.Time    `json:"occurred_at"`
	Actual    time.Duration `json:"actual_ms"`
	Min       time.Duration `json:"min_ms"`
	Max       time.Duration `json:"max_ms"`
}

// History stores a bounded list of recent drift events.
type History struct {
	mu     sync.RWMutex
	events []DriftEvent
	limit  int
}

// NewHistory creates a History that retains at most limit events.
func NewHistory(limit int) *History {
	if limit <= 0 {
		limit = 100
	}
	return &History{limit: limit}
}

// Record appends a drift event, evicting the oldest if the limit is reached.
func (h *History) Record(e DriftEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.events) >= h.limit {
		h.events = h.events[1:]
	}
	h.events = append(h.events, e)
}

// Snapshot returns a copy of all stored events.
func (h *History) Snapshot() []DriftEvent {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]DriftEvent, len(h.events))
	copy(out, h.events)
	return out
}

// handleHistory writes the drift event history as JSON.
func handleHistory(h *History) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		events := h.Snapshot()
		if events == nil {
			events = []DriftEvent{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}
}
