package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTagManager_SetAndGet(t *testing.T) {
	tm := NewTagManager()
	tm.Set("backup", []string{"infra", "nightly"})
	tags := tm.Get("backup")
	if len(tags) != 2 || tags[0] != "infra" {
		t.Fatalf("unexpected tags: %v", tags)
	}
}

func TestTagManager_GetUnknown(t *testing.T) {
	tm := NewTagManager()
	if tags := tm.Get("missing"); tags != nil {
		t.Fatalf("expected nil, got %v", tags)
	}
}

func TestTagManager_Snapshot(t *testing.T) {
	tm := NewTagManager()
	tm.Set("job1", []string{"a"})
	tm.Set("job2", []string{"b", "c"})
	snap := tm.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
}

func TestHandleTags_GetReturnsJSON(t *testing.T) {
	tm := NewTagManager()
	tm.Set("deploy", []string{"prod"})

	req := httptest.NewRequest(http.MethodGet, "/tags", nil)
	rec := httptest.NewRecorder()
	tm.handleTags(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out map[string][]string
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out["deploy"]) != 1 || out["deploy"][0] != "prod" {
		t.Fatalf("unexpected body: %v", out)
	}
}

func TestHandleTags_PostSetsTag(t *testing.T) {
	tm := NewTagManager()
	body, _ := json.Marshal(map[string]interface{}{"job": "sync", "tags": []string{"hourly"}})

	req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	tm.handleTags(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if tags := tm.Get("sync"); len(tags) != 1 || tags[0] != "hourly" {
		t.Fatalf("tag not stored: %v", tags)
	}
}

func TestHandleTags_PostBadBody(t *testing.T) {
	tm := NewTagManager()
	req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()
	tm.handleTags(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleTags_MethodNotAllowed(t *testing.T) {
	tm := NewTagManager()
	req := httptest.NewRequest(http.MethodDelete, "/tags", nil)
	rec := httptest.NewRecorder()
	tm.handleTags(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
