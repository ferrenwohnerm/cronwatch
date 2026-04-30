package tracker

import (
	"sync"
	"time"
)

// JobRecord holds timing information for a single cron job execution.
type JobRecord struct {
	JobName   string
	StartedAt time.Time
	FinishedAt *time.Time
	Duration  *time.Duration
}

// Tracker maintains in-memory state of recent job executions.
type Tracker struct {
	mu      sync.RWMutex
	records map[string]*JobRecord
}

// New returns an initialised Tracker.
func New() *Tracker {
	return &Tracker{
		records: make(map[string]*JobRecord),
	}
}

// Start records the start time for a named job.
func (t *Tracker) Start(jobName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.records[jobName] = &JobRecord{
		JobName:   jobName,
		StartedAt: time.Now(),
	}
}

// Finish records the completion time for a named job and computes its duration.
// It returns false if no corresponding Start was recorded.
func (t *Tracker) Finish(jobName string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	rec, ok := t.records[jobName]
	if !ok {
		return false
	}
	now := time.Now()
	d := now.Sub(rec.StartedAt)
	rec.FinishedAt = &now
	rec.Duration = &d
	return true
}

// Get returns a copy of the JobRecord for the given job name.
// The second return value is false when no record exists.
func (t *Tracker) Get(jobName string) (JobRecord, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	rec, ok := t.records[jobName]
	if !ok {
		return JobRecord{}, false
	}
	return *rec, true
}

// Delete removes the record for the given job name.
func (t *Tracker) Delete(jobName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.records, jobName)
}
