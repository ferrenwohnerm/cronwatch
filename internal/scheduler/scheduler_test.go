package scheduler_test

import (
	"sync"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/alert"
	"github.com/example/cronwatch/internal/config"
	"github.com/example/cronwatch/internal/scheduler"
	"github.com/example/cronwatch/internal/tracker"
)

// captureNotifier records every alert it receives.
type captureNotifier struct {
	mu     sync.Mutex
	alerts []alert.Alert
}

func (c *captureNotifier) Send(a alert.Alert) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.alerts = append(c.alerts, a)
	return nil
}

func (c *captureNotifier) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.alerts)
}

func TestScheduler_StartsAndStops(t *testing.T) {
	cfg := &config.Config{
		PollInterval: 50 * time.Millisecond,
		Jobs:         []config.Job{},
	}
	t_ := tracker.New()
	am := alert.NewManager()
	s := scheduler.New(cfg, t_, am)
	s.Start()
	time.Sleep(120 * time.Millisecond)
	s.Stop() // should not block or panic
}

func TestScheduler_AlertOnDrift(t *testing.T) {
	notifier := &captureNotifier{}
	cfg := &config.Config{
		PollInterval: 30 * time.Millisecond,
		Jobs: []config.Job{
			{
				Name:             "slow-job",
				ExpectedDuration: 100 * time.Millisecond,
				Tolerance:        10 * time.Millisecond,
			},
		},
	}
	t_ := tracker.New()
	// Simulate a job that ran much longer than expected.
	t_.Start("slow-job")
	time.Sleep(5 * time.Millisecond)
	t_.Finish("slow-job", 500*time.Millisecond) // inject a long duration

	am := alert.NewManager(notifier)
	s := scheduler.New(cfg, t_, am)
	s.Start()
	time.Sleep(80 * time.Millisecond)
	s.Stop()

	if notifier.count() == 0 {
		t.Error("expected at least one drift alert, got none")
	}
}
