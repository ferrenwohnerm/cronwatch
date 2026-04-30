package alert

import (
	"errors"
	"testing"
	"time"
)

// stubNotifier records calls and optionally returns an error.
type stubNotifier struct {
	Called bool
	Last   Alert
	Err    error
}

func (s *stubNotifier) Send(a Alert) error {
	s.Called = true
	s.Last = a
	return s.Err
}

func TestManager_SendDispatchesToAll(t *testing.T) {
	s1 := &stubNotifier{}
	s2 := &stubNotifier{}
	m := NewManager(s1, s2)

	a := Alert{JobName: "backup", Level: LevelWarn, Message: "slow", OccurredAt: time.Now()}
	if err := m.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s1.Called || !s2.Called {
		t.Error("expected both notifiers to be called")
	}
}

func TestManager_SendReturnsFirstError(t *testing.T) {
	s1 := &stubNotifier{Err: errors.New("fail")}
	s2 := &stubNotifier{}
	m := NewManager(s1, s2)

	a := Alert{JobName: "backup", Level: LevelError, Message: "timeout", OccurredAt: time.Now()}
	err := m.Send(a)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !s2.Called {
		t.Error("second notifier should still be called despite first error")
	}
}

func TestNewDriftAlert(t *testing.T) {
	a := NewDriftAlert("sync", LevelWarn, 15.5)
	if a.JobName != "sync" {
		t.Errorf("expected job name 'sync', got %q", a.JobName)
	}
	if a.Level != LevelWarn {
		t.Errorf("expected level WARN, got %q", a.Level)
	}
	if a.OccurredAt.IsZero() {
		t.Error("expected OccurredAt to be set")
	}
}

func TestLogNotifier_Send(t *testing.T) {
	ln := &LogNotifier{}
	a := Alert{JobName: "test", Level: LevelError, Message: "too slow", OccurredAt: time.Now()}
	if err := ln.Send(a); err != nil {
		t.Errorf("LogNotifier.Send returned unexpected error: %v", err)
	}
}
