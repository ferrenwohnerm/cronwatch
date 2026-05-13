package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJobGroupManager_SetAndGet(t *testing.T) {
	m := NewJobGroupManager()
	m.Set("backup", "infra")
	g, ok := m.Get("backup")
	if !ok || g != "infra" {
		t.Fatalf("expected group infra, got %q ok=%v", g, ok)
	}
}

func TestJobGroupManager_GetUnknown(t *testing.T) {
	m := NewJobGroupManager()
	_, ok := m.Get("ghost")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestJobGroupManager_Delete(t *testing.T) {
	m := NewJobGroupManager()
	m.Set("backup", "infra")
	m.Delete("backup")
	_, ok := m.Get("backup")
	if ok {
		t.Fatal("expected job to be deleted")
	}
}

func TestJobGroupManager_Snapshot(t *testing.T) {
	m := NewJobGroupManager()
	m.Set("a", "g1")
	m.Set("b", "g2")
	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	snap["c"] = "g3"
	if _, ok := m.Get("c"); ok {
		t.Fatal("snapshot mutation should not affect manager")
	}
}

func TestHandleJobGroup_PostAndGet(t *testing.T) {
	m := NewJobGroupManager()

	body, _ := json.Marshal(map[string]string{"job": "deploy", "group": "release"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/groups?job=deploy", nil)
	rec2 := httptest.NewRecorder()
	m.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var resp map[string]string
	json.NewDecoder(rec2.Body).Decode(&resp)
	if resp["group"] != "release" {
		t.Fatalf("expected group release, got %q", resp["group"])
	}
}

func TestHandleJobGroup_GetAll(t *testing.T) {
	m := NewJobGroupManager()
	m.Set("a", "g1")
	m.Set("b", "g2")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(resp))
	}
}

func TestHandleJobGroup_MethodNotAllowed(t *testing.T) {
	m := NewJobGroupManager()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/groups", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleJobGroup_DeleteMissingParam(t *testing.T) {
	m := NewJobGroupManager()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
