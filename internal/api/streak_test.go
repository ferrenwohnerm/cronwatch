package api

import (
	"testing"
)

func TestStreakManager_SuccessStreak(t *testing.T) {
	sm := NewStreakManager()
	sm.Record("backup", "success")
	sm.Record("backup", "success")
	sm.Record("backup", "success")

	e, ok := sm.Get("backup")
	if !ok {
		t.Fatal("expected entry for backup")
	}
	if e.SuccessStreak != 3 {
		t.Errorf("expected success streak 3, got %d", e.SuccessStreak)
	}
	if e.FailureStreak != 0 {
		t.Errorf("expected failure streak 0, got %d", e.FailureStreak)
	}
	if e.LastOutcome != "success" {
		t.Errorf("expected last outcome success, got %s", e.LastOutcome)
	}
}

func TestStreakManager_FailureResetsSuccess(t *testing.T) {
	sm := NewStreakManager()
	sm.Record("sync", "success")
	sm.Record("sync", "success")
	sm.Record("sync", "failure")

	e, _ := sm.Get("sync")
	if e.SuccessStreak != 0 {
		t.Errorf("expected success streak reset to 0, got %d", e.SuccessStreak)
	}
	if e.FailureStreak != 1 {
		t.Errorf("expected failure streak 1, got %d", e.FailureStreak)
	}
}

func TestStreakManager_GetUnknown(t *testing.T) {
	sm := NewStreakManager()
	_, ok := sm.Get("ghost")
	if ok {
		t.Error("expected false for unknown job")
	}
}

func TestStreakManager_Snapshot(t *testing.T) {
	sm := NewStreakManager()
	sm.Record("jobA", "success")
	sm.Record("jobB", "failure")

	snap := sm.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 entries, got %d", len(snap))
	}
}

func TestStreakManager_MultipleFailures(t *testing.T) {
	sm := NewStreakManager()
	sm.Record("report", "failure")
	sm.Record("report", "failure")
	sm.Record("report", "failure")

	e, _ := sm.Get("report")
	if e.FailureStreak != 3 {
		t.Errorf("expected failure streak 3, got %d", e.FailureStreak)
	}
	if e.SuccessStreak != 0 {
		t.Errorf("expected success streak 0, got %d", e.SuccessStreak)
	}
}
