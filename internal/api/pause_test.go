package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPauseManager_PauseAndCheck(t *testing.T) {
	pm := NewPauseManager()
	pm.Pause("backup", 0)
	if !pm.IsPaused("backup") {
		t.Fatal("expected job to be paused indefinitely")
	}
}

func TestPauseManager_PauseExpires(t *testing.T) {
	pm := NewPauseManager()
	pm.Pause("backup", 10*time.Millisecond)
	if !pm.IsPaused("backup") {
		t.Fatal("expected job to be paused immediately after pause")
	}
	time.Sleep(20 * time.Millisecond)
	if pm.IsPaused("backup") {
		t.Fatal("expected pause to have expired")
	}
}

func TestPauseManager_Resume(t *testing.T) {
	pm := NewPauseManager()
	pm.Pause("sync", 0)
	pm.Resume("sync")
	if pm.IsPaused("sync") {
		t.Fatal("expected job to be resumed")
	}
}

func TestPauseManager_UnknownJobNotPaused(t *testing.T) {
	pm := NewPauseManager()
	if pm.IsPaused("nonexistent") {
		t.Fatal("expected unknown job to not be paused")
	}
}

func TestHandlePause_Success(t *testing.T) {
	pm := NewPauseManager()
	body := bytes.NewBufferString(`{"job":"backup","duration":"1h"}`)
	req := httptest.NewRequest(http.MethodPost, "/pause", body)
	rw := httptest.NewRecorder()
	pm.handlePause(rw, req)
	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}
	if !pm.IsPaused("backup") {
		t.Fatal("expected job to be paused after request")
	}
}

func TestHandlePause_MethodNotAllowed(t *testing.T) {
	pm := NewPauseManager()
	req := httptest.NewRequest(http.MethodGet, "/pause", nil)
	rw := httptest.NewRecorder()
	pm.handlePause(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}

func TestHandlePause_InvalidDuration(t *testing.T) {
	pm := NewPauseManager()
	body := bytes.NewBufferString(`{"job":"backup","duration":"not-a-duration"}`)
	req := httptest.NewRequest(http.MethodPost, "/pause", body)
	rw := httptest.NewRecorder()
	pm.handlePause(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestHandleResume_Success(t *testing.T) {
	pm := NewPauseManager()
	pm.Pause("sync", 0)
	body := bytes.NewBufferString(`{"job":"sync"}`)
	req := httptest.NewRequest(http.MethodPost, "/resume", body)
	rw := httptest.NewRecorder()
	pm.handleResume(rw, req)
	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}
	if pm.IsPaused("sync") {
		t.Fatal("expected job to be resumed after request")
	}
}

func TestHandleResume_MethodNotAllowed(t *testing.T) {
	pm := NewPauseManager()
	req := httptest.NewRequest(http.MethodGet, "/resume", nil)
	rw := httptest.NewRecorder()
	pm.handleResume(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}
