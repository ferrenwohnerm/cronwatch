package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/alert"
	"github.com/example/cronwatch/internal/tracker"
)

func newNotifySetup(t *testing.T) (*NotifyHandler, *tracker.Tracker) {
	t.Helper()
	tr := tracker.New()
	mgr := alert.NewManager()
	h := NewNotifyHandler(tr, mgr)
	return h, tr
}

func TestHandleNotify_MethodNotAllowed(t *testing.T) {
	h, _ := newNotifySetup(t)
	req := httptest.NewRequest(http.MethodGet, "/notify", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleNotify_BadBody(t *testing.T) {
	h, _ := newNotifySetup(t)
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBufferString(`{bad}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleNotify_JobNotFound(t *testing.T) {
	h, _ := newNotifySetup(t)
	body, _ := json.Marshal(map[string]interface{}{"job_name": "missing", "actual_ms": 1000})
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleNotify_Success(t *testing.T) {
	h, tr := newNotifySetup(t)
	tr.Start("backup", 5*time.Second)

	body, _ := json.Marshal(map[string]interface{}{"job_name": "backup", "actual_ms": 8000})
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(body))
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp notifyResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !resp.Triggered {
		t.Error("expected triggered=true")
	}
}
