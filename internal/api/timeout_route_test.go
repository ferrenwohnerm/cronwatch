package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTimeoutRouter() (*http.ServeMux, *TimeoutManager) {
	mux := http.NewServeMux()
	m := NewTimeoutManager()
	RegisterTimeoutRoutes(mux, m)
	return mux, m
}

func TestTimeoutRoute_GetEmpty(t *testing.T) {
	mux, _ := newTimeoutRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/timeouts", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out []TimeoutEntry
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty slice, got %v", out)
	}
}

func TestTimeoutRoute_PostAndGet(t *testing.T) {
	mux, m := newTimeoutRouter()

	body := `{"job_name":"nightly","timeout_seconds":60}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/timeouts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	e, ok := m.Get("nightly")
	if !ok {
		t.Fatal("expected entry after POST")
	}
	if e.Timeout.Seconds() != 60 {
		t.Errorf("expected 60s, got %v", e.Timeout)
	}
}

func TestTimeoutRoute_MethodNotAllowed(t *testing.T) {
	mux, _ := newTimeoutRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/timeouts", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
