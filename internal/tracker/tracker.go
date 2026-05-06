package tracker

import (
	"errors"
	"sync"
	"time"
)

// Entry holds the state of a tracked cron job.
type Entry struct {
	JobName   string
	StartedAt time.Time
	Expected  time.Duration
	Running   bool
}

// Tracker stores in-flight job state.
type Tracker struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// New returns an initialised Tracker.
func New() *Tracker {
	return &Tracker{entries: make(map[string]*Entry)}
}

// Start records that a job has begun.
func (t *Tracker) Start(name string, expected time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries[name] = &Entry{
		JobName:   name,
		StartedAt: time.Now(),
		Expected:  expected,
		Running:   true,
	}
}

// Finish marks a job complete and returns elapsed duration.
func (t *Tracker) Finish(name string) (time.Duration, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[name]
	if !ok || !e.Running {
		return 0, errors.New("no running entry for job: " + name)
	}
	elapsed := time.Since(e.StartedAt)
	e.Running = false
	return elapsed, nil
}

// Get retrieves a job entry (running or finished).
func (t *Tracker) Get(name string) (*Entry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	e, ok := t.entries[name]
	return e, ok
}

// Delete removes a job entry.
func (t *Tracker) Delete(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, name)
}

// Snapshot returns a copy of all current entries.
func (t *Tracker) Snapshot() []Entry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]Entry, 0, len(t.entries))
	for _, e := range t.entries {
		out = append(out, *e)
	}
	return out
}
