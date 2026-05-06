package api

import (
	"testing"
)

func TestLabelManager_SetAndGet(t *testing.T) {
	lm := NewLabelManager()
	lm.Set("backup", "env", "prod")
	lm.Set("backup", "team", "ops")

	labels, ok := lm.Get("backup")
	if !ok {
		t.Fatal("expected labels to exist")
	}
	if labels["env"] != "prod" {
		t.Errorf("expected env=prod, got %s", labels["env"])
	}
	if labels["team"] != "ops" {
		t.Errorf("expected team=ops, got %s", labels["team"])
	}
}

func TestLabelManager_GetUnknown(t *testing.T) {
	lm := NewLabelManager()
	_, ok := lm.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestLabelManager_Delete(t *testing.T) {
	lm := NewLabelManager()
	lm.Set("job1", "k", "v")
	lm.Delete("job1")
	_, ok := lm.Get("job1")
	if ok {
		t.Error("expected labels to be deleted")
	}
}

func TestLabelManager_Snapshot(t *testing.T) {
	lm := NewLabelManager()
	lm.Set("alpha", "x", "1")
	lm.Set("beta", "y", "2")

	snap := lm.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(snap))
	}
	if snap["alpha"]["x"] != "1" {
		t.Errorf("unexpected snapshot value for alpha.x")
	}
}

func TestLabelManager_IsolatedCopy(t *testing.T) {
	lm := NewLabelManager()
	lm.Set("job", "k", "original")

	labels, _ := lm.Get("job")
	labels["k"] = "mutated"

	again, _ := lm.Get("job")
	if again["k"] != "original" {
		t.Error("Get should return an isolated copy")
	}
}
