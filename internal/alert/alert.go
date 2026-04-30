package alert

import (
	"fmt"
	"log"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Alert holds information about a drift event.
type Alert struct {
	JobName   string
	Level     Level
	Message   string
	OccurredAt time.Time
}

// Notifier is the interface for sending alerts.
type Notifier interface {
	Send(a Alert) error
}

// LogNotifier writes alerts to the standard logger.
type LogNotifier struct{}

// Send implements Notifier for LogNotifier.
func (l *LogNotifier) Send(a Alert) error {
	log.Printf("[%s] cronwatch alert for job %q at %s: %s",
		a.Level, a.JobName, a.OccurredAt.Format(time.RFC3339), a.Message)
	return nil
}

// Manager dispatches alerts through one or more Notifiers.
type Manager struct {
	Notifiers []Notifier
}

// NewManager creates a Manager with the provided notifiers.
func NewManager(notifiers ...Notifier) *Manager {
	return &Manager{Notifiers: notifiers}
}

// Send dispatches the alert to all registered notifiers.
// It returns the first error encountered, but attempts all notifiers.
func (m *Manager) Send(a Alert) error {
	var firstErr error
	for _, n := range m.Notifiers {
		if err := n.Send(a); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("notifier %T: %w", n, err)
		}
	}
	return firstErr
}

// NewDriftAlert builds an Alert for a drift condition.
func NewDriftAlert(jobName string, level Level, driftSeconds float64) Alert {
	msg := fmt.Sprintf("job duration drifted by %.2fs outside expected window", driftSeconds)
	return Alert{
		JobName:    jobName,
		Level:      level,
		Message:    msg,
		OccurredAt: time.Now(),
	}
}
