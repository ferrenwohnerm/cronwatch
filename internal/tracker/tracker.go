package tracker

import (
	"errors"
	"sync"
	"time"
)

// run holds timing data for a single job execution.
type run struct {
	start    time.Time
	duration time.Duration
	done     bool
}

// Tracker records start/finish times for named cron jobs.
type Tracker struct {
	mu   sync.Mutex
	runs map[string]*run
}

// New returns an initialised Tracker.
func New() *Tracker {
	return &Tracker{runs: make(map[string]*run)}
}

// Start records the beginning of a job execution.
func (t *Tracker) Start(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.runs[name] = &run{start: time.Now()}
}

// Finish marks a job as complete. An optional override duration may be
// supplied (non-zero) to inject a duration directly (useful for testing).
func (t *Tracker) Finish(name string, override time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.runs[name]
	if !ok {
		return errors.New("tracker: no active run for job " + name)
	}
	if override > 0 {
		r.duration = override
	} else {
		r.duration = time.Since(r.start)
	}
	r.done = true
	return nil
}

// Delete removes all tracking state for a job.
func (t *Tracker) Delete(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.runs, name)
}

// LastDuration returns the recorded duration for the most recent completed
// run of name, or an error if no completed run exists.
func (t *Tracker) LastDuration(name string) (time.Duration, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.runs[name]
	if !ok || !r.done {
		return 0, errors.New("tracker: no completed run for job " + name)
	}
	return r.duration, nil
}
