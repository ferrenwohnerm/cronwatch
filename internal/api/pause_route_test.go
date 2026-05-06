package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newPauseRouter() (*http.ServeMux, *PauseManager) {
	mux := http.NewServeMux()
	pm := NewPauseManager()
	RegisterPauseRoutes(mux, pm)
	return mux, pm
}

func TestPauseRoute_PauseEndpoint(t *testing.T) {
	mux, pm := newPauseRouter()
	body := bytes.NewBufferString(`{"job":"nightly","duration":"2h"}`)
	req := httptest.NewRequest(http.MethodPost, "/pause", body)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)
	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}
	if !pm.IsPaused("nightly") {
		t.Fatal("expected job to be paused via route")
	}
}

func TestPauseRoute_ResumeEndpoint(t *testing.T) {
	mux, pm := newPauseRouter()
	pm.Pause("nightly", 0)
	body := bytes.NewBufferString(`{"job":"nightly"}`)
	req := httptest.NewRequest(http.MethodPost, "/resume", body)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)
	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}
	if pm.IsPaused("nightly") {
		t.Fatal("expected job to be resumed via route")
	}
}

func TestPauseRoute_IndefinitePause(t *testing.T) {
	mux, pm := newPauseRouter()
	body := bytes.NewBufferString(`{"job":"cleanup"}`)
	req := httptest.NewRequest(http.MethodPost, "/pause", body)
	rw := httptest.NewRecorder()
	mux.ServeHTTP(rw, req)
	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}
	if !pm.IsPaused("cleanup") {
		t.Fatal("expected job to be paused indefinitely via route")
	}
}
