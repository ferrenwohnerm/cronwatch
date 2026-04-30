package tracker

import (
	"testing"
	"time"
)

func TestTracker_StartAndFinish(t *testing.T) {
	tr := New()
	tr.Start("backup")

	time.Sleep(10 * time.Millisecond)

	ok := tr.Finish("backup")
	if !ok {
		t.Fatal("expected Finish to return true for a started job")
	}

	rec, found := tr.Get("backup")
	if !found {
		t.Fatal("expected record to exist after finish")
	}
	if rec.Duration == nil {
		t.Fatal("expected duration to be set")
	}
	if *rec.Duration < 10*time.Millisecond {
		t.Errorf("duration %s shorter than expected", *rec.Duration)
	}
}

func TestTracker_FinishWithoutStart(t *testing.T) {
	tr := New()
	if ok := tr.Finish("nonexistent"); ok {
		t.Error("expected Finish to return false for unknown job")
	}
}

func TestTracker_Delete(t *testing.T) {
	tr := New()
	tr.Start("cleanup")
	tr.Delete("cleanup")
	_, found := tr.Get("cleanup")
	if found {
		t.Error("expected record to be removed after Delete")
	}
}

func TestCheckDrift_TooFast(t *testing.T) {
	d := 5 * time.Millisecond
	rec := JobRecord{JobName: "fast-job", Duration: &d}
	res, err := CheckDrift(rec, 10*time.Millisecond, 60*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Drifted {
		t.Error("expected drift to be detected for fast job")
	}
}

func TestCheckDrift_TooSlow(t *testing.T) {
	d := 2 * time.Minute
	rec := JobRecord{JobName: "slow-job", Duration: &d}
	res, err := CheckDrift(rec, 1*time.Second, 60*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Drifted {
		t.Error("expected drift to be detected for slow job")
	}
}

func TestCheckDrift_Normal(t *testing.T) {
	d := 30 * time.Second
	rec := JobRecord{JobName: "normal-job", Duration: &d}
	res, err := CheckDrift(rec, 10*time.Second, 60*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Drifted {
		t.Error("expected no drift for a normally completing job")
	}
}

func TestCheckDrift_NotFinished(t *testing.T) {
	rec := JobRecord{JobName: "pending-job"}
	_, err := CheckDrift(rec, time.Second, time.Minute)
	if err == nil {
		t.Error("expected error for unfinished job")
	}
}
