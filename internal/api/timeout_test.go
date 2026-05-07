package api

import (
	"testing"
	"time"
)

func TestTimeoutManager_SetAndGet(t *testing.T) {
	m := NewTimeoutManager()
	m.Set("backup", 30*time.Second)
	e, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Timeout != 30*time.Second {
		t.Errorf("expected 30s, got %v", e.Timeout)
	}
	if e.JobName != "backup" {
		t.Errorf("unexpected job name: %s", e.JobName)
	}
}

func TestTimeoutManager_GetUnknown(t *testing.T) {
	m := NewTimeoutManager()
	_, ok := m.Get("ghost")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestTimeoutManager_Delete(t *testing.T) {
	m := NewTimeoutManager()
	m.Set("cleanup", time.Minute)
	m.Delete("cleanup")
	_, ok := m.Get("cleanup")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestTimeoutManager_Snapshot(t *testing.T) {
	m := NewTimeoutManager()
	m.Set("job-a", 10*time.Second)
	m.Set("job-b", 20*time.Second)
	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
}

func TestTimeoutManager_OverwritesPrevious(t *testing.T) {
	m := NewTimeoutManager()
	m.Set("job", 5*time.Second)
	m.Set("job", 15*time.Second)
	e, _ := m.Get("job")
	if e.Timeout != 15*time.Second {
		t.Errorf("expected overwritten value 15s, got %v", e.Timeout)
	}
	if len(m.Snapshot()) != 1 {
		t.Error("expected exactly one entry after overwrite")
	}
}
