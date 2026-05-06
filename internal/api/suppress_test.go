package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSuppressManager_SuppressAndCheck(t *testing.T) {
	sm := NewSuppressManager()
	if sm.IsSuppressed("job1") {
		t.Fatal("expected job1 to not be suppressed")
	}
	sm.Suppress("job1", 5*time.Minute)
	if !sm.IsSuppressed("job1") {
		t.Fatal("expected job1 to be suppressed")
	}
}

func TestSuppressManager_ExpiredSuppression(t *testing.T) {
	sm := NewSuppressManager()
	sm.Suppress("job2", -1*time.Second)
	if sm.IsSuppressed("job2") {
		t.Fatal("expected expired suppression to not be active")
	}
}

func TestSuppressManager_Snapshot(t *testing.T) {
	sm := NewSuppressManager()
	sm.Suppress("job3", 10*time.Minute)
	sm.Suppress("job4", -1*time.Second) // expired
	snap := sm.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 active suppression, got %d", len(snap))
	}
	if snap[0].JobName != "job3" {
		t.Errorf("expected job3, got %s", snap[0].JobName)
	}
}

func TestHandleSuppress_Success(t *testing.T) {
	sm := NewSuppressManager()
	body := `{"job_name":"myjob","duration":"30m"}`
	req := httptest.NewRequest(http.MethodPost, "/suppress", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	sm.handleSuppress(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !sm.IsSuppressed("myjob") {
		t.Error("expected myjob to be suppressed after POST")
	}
}

func TestHandleSuppress_MethodNotAllowed(t *testing.T) {
	sm := NewSuppressManager()
	req := httptest.NewRequest(http.MethodGet, "/suppress", nil)
	w := httptest.NewRecorder()
	sm.handleSuppress(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleSuppress_InvalidDuration(t *testing.T) {
	sm := NewSuppressManager()
	body := `{"job_name":"myjob","duration":"notaduration"}`
	req := httptest.NewRequest(http.MethodPost, "/suppress", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	sm.handleSuppress(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleSuppressList_ReturnsActive(t *testing.T) {
	sm := NewSuppressManager()
	sm.Suppress("job5", 1*time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/suppress", nil)
	w := httptest.NewRecorder()
	sm.handleSuppressList(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []SuppressedJob
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 1 || result[0].JobName != "job5" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestHandleSuppressList_MethodNotAllowed(t *testing.T) {
	sm := NewSuppressManager()
	req := httptest.NewRequest(http.MethodPost, "/suppress", nil)
	w := httptest.NewRecorder()
	sm.handleSuppressList(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
