package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// SuppressedJob tracks a job name and the time until which alerts are suppressed.
type SuppressedJob struct {
	JobName   string    `json:"job_name"`
	Until     time.Time `json:"until"`
	CreatedAt time.Time `json:"created_at"`
}

// SuppressManager manages alert suppression windows for jobs.
type SuppressManager struct {
	mu          sync.RWMutex
	suppressed  map[string]SuppressedJob
}

// NewSuppressManager creates a new SuppressManager.
func NewSuppressManager() *SuppressManager {
	return &SuppressManager{
		suppressed: make(map[string]SuppressedJob),
	}
}

// Suppress adds or updates a suppression window for the given job.
func (s *SuppressManager) Suppress(jobName string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.suppressed[jobName] = SuppressedJob{
		JobName:   jobName,
		Until:     now.Add(duration),
		CreatedAt: now,
	}
}

// IsSuppressed reports whether alerts for the given job are currently suppressed.
func (s *SuppressManager) IsSuppressed(jobName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sj, ok := s.suppressed[jobName]
	if !ok {
		return false
	}
	return time.Now().Before(sj.Until)
}

// Snapshot returns a copy of all active suppressions.
func (s *SuppressManager) Snapshot() []SuppressedJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	result := make([]SuppressedJob, 0, len(s.suppressed))
	for _, sj := range s.suppressed {
		if now.Before(sj.Until) {
			result = append(result, sj)
		}
	}
	return result
}

// handleSuppress handles POST /suppress to add a suppression window.
func (s *SuppressManager) handleSuppress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		JobName  string `json:"job_name"`
		Duration string `json:"duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.JobName == "" {
		http.Error(w, "job_name is required", http.StatusBadRequest)
		return
	}
	dur, err := time.ParseDuration(req.Duration)
	if err != nil || dur <= 0 {
		http.Error(w, "invalid duration", http.StatusBadRequest)
		return
	}
	s.Suppress(req.JobName, dur)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "suppressed", "job_name": req.JobName})
}

// handleSuppressList handles GET /suppress to list active suppressions.
func (s *SuppressManager) handleSuppressList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	snap := s.Snapshot()
	if snap == nil {
		snap = []SuppressedJob{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}
