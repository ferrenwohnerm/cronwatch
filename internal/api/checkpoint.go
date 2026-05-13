package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// CheckpointEntry records the last known checkpoint for a job.
type CheckpointEntry struct {
	JobName     string    `json:"job_name"`
	Checkpoint  string    `json:"checkpoint"`
	RecordedAt  time.Time `json:"recorded_at"`
}

// CheckpointManager stores the latest checkpoint value per job.
type CheckpointManager struct {
	mu   sync.RWMutex
	data map[string]CheckpointEntry
}

// NewCheckpointManager returns an initialised CheckpointManager.
func NewCheckpointManager() *CheckpointManager {
	return &CheckpointManager{
		data: make(map[string]CheckpointEntry),
	}
}

// Set records a checkpoint for the given job.
func (m *CheckpointManager) Set(jobName, checkpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[jobName] = CheckpointEntry{
		JobName:    jobName,
		Checkpoint: checkpoint,
		RecordedAt: time.Now().UTC(),
	}
}

// Get returns the checkpoint entry for a job and whether it exists.
func (m *CheckpointManager) Get(jobName string) (CheckpointEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.data[jobName]
	return e, ok
}

// Delete removes the checkpoint for a job.
func (m *CheckpointManager) Delete(jobName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, jobName)
}

// Snapshot returns a copy of all checkpoint entries.
func (m *CheckpointManager) Snapshot() []CheckpointEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]CheckpointEntry, 0, len(m.data))
	for _, e := range m.data {
		out = append(out, e)
	}
	return out
}

// HandleCheckpoints handles GET/POST/DELETE for /checkpoints?job=<name>.
func (m *CheckpointManager) HandleCheckpoints(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")

	switch r.Method {
	case http.MethodGet:
		if job == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(m.Snapshot())
			return
		}
		e, ok := m.Get(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(e)

	case http.MethodPost:
		if job == "" {
			http.Error(w, "job parameter required", http.StatusBadRequest)
			return
		}
		var body struct {
			Checkpoint string `json:"checkpoint"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Checkpoint == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		m.Set(job, body.Checkpoint)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if job == "" {
			http.Error(w, "job parameter required", http.StatusBadRequest)
			return
		}
		m.Delete(job)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
