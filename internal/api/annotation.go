package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Annotation represents a timestamped note attached to a job.
type Annotation struct {
	JobName   string    `json:"job_name"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// AnnotationManager stores per-job annotations.
type AnnotationManager struct {
	mu          sync.RWMutex
	annotations map[string][]Annotation
	maxPerJob   int
}

// NewAnnotationManager creates an AnnotationManager with a per-job cap.
func NewAnnotationManager(maxPerJob int) *AnnotationManager {
	if maxPerJob <= 0 {
		maxPerJob = 20
	}
	return &AnnotationManager{
		annotations: make(map[string][]Annotation),
		maxPerJob:   maxPerJob,
	}
}

// Add appends an annotation for the given job, evicting the oldest if at cap.
func (m *AnnotationManager) Add(jobName, note string) Annotation {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := Annotation{JobName: jobName, Note: note, CreatedAt: time.Now().UTC()}
	list := m.annotations[jobName]
	if len(list) >= m.maxPerJob {
		list = list[1:]
	}
	m.annotations[jobName] = append(list, a)
	return a
}

// Get returns all annotations for a job.
func (m *AnnotationManager) Get(jobName string) []Annotation {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := m.annotations[jobName]
	out := make([]Annotation, len(list))
	copy(out, list)
	return out
}

// Snapshot returns a copy of all annotations keyed by job name.
func (m *AnnotationManager) Snapshot() map[string][]Annotation {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string][]Annotation, len(m.annotations))
	for k, v := range m.annotations {
		cp := make([]Annotation, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

// handleAnnotations serves GET and POST for /annotations?job=<name>.
func (m *AnnotationManager) handleAnnotations(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		list := m.Get(job)
		if list == nil {
			list = []Annotation{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	case http.MethodPost:
		var body struct {
			Note string `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Note == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		a := m.Add(job, body.Note)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(a)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
