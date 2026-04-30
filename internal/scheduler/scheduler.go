package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/example/cronwatch/internal/alert"
	"github.com/example/cronwatch/internal/config"
	"github.com/example/cronwatch/internal/tracker"
)

// Scheduler periodically checks for overdue cron jobs and fires alerts.
type Scheduler struct {
	cfg     *config.Config
	tracker *tracker.Tracker
	alerts  *alert.Manager
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// New creates a new Scheduler.
func New(cfg *config.Config, t *tracker.Tracker, am *alert.Manager) *Scheduler {
	return &Scheduler{
		cfg:     cfg,
		tracker: t,
		alerts:  am,
		stopCh:  make(chan struct{}),
	}
}

// Start begins the polling loop in a background goroutine.
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.cfg.PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.check()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop signals the polling loop to exit and waits for it to finish.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

// check inspects every configured job for drift.
func (s *Scheduler) check() {
	for _, job := range s.cfg.Jobs {
		drift, err := s.tracker.CheckDrift(job.Name, job.ExpectedDuration, job.Tolerance)
		if err != nil {
			log.Printf("[scheduler] drift check skipped for %q: %v", job.Name, err)
			continue
		}
		if drift != 0 {
			a := alert.NewDriftAlert(job.Name, drift, job.ExpectedDuration)
			if err := s.alerts.Send(a); err != nil {
				log.Printf("[scheduler] alert send failed for %q: %v", job.Name, err)
			}
		}
	}
}
