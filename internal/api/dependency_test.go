package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDependencyManager_AddAndGet(t *testing.T) {
	dm := NewDependencyManager()
	dm.Add("job-b", "job-a")
	deps := dm.Get("job-b")
	if len(deps) != 1 || deps[0] != "job-a" {
		t.Fatalf("expected [job-a], got %v", deps)
	}
}

func TestDependencyManager_NoDuplicates(t *testing.T) {
	dm := NewDependencyManager()
	dm.Add("job-b", "job-a")
	dm.Add("job-b", "job-a")
	if len(dm.Get("job-b")) != 1 {
		t.Fatal("expected deduplication")
	}
}

func TestDependencyManager_Remove(t *testing.T) {
	dm := NewDependencyManager()
	dm.Add("job-b", "job-a")
	dm.Remove("job-b", "job-a")
	if len(dm.Get("job-b")) != 0 {
		t.Fatal("expected empty after remove")
	}
}

func TestDependencyManager_Snapshot(t *testing.T) {
	dm := NewDependencyManager()
	dm.Add("job-b", "job-a")
	dm.Add("job-c", "job-a")
	snap := dm.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(snap))
	}
}

func TestHandleDependencies_MissingJob(t *testing.T) {
	dm := NewDependencyManager()
	req := httptest.NewRequest(http.MethodGet, "/deps", nil)
	w := httptest.NewRecorder()
	dm.handleDependencies(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleDependencies_GetReturnsJSON(t *testing.T) {
	dm := NewDependencyManager()
	dm.Add("job-b", "job-a")
	req := httptest.NewRequest(http.MethodGet, "/deps?job=job-b", nil)
	w := httptest.NewRecorder()
	dm.handleDependencies(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	deps, ok := resp["depends_on"].([]interface{})
	if !ok || len(deps) != 1 {
		t.Fatalf("unexpected depends_on: %v", resp["depends_on"])
	}
}

func TestHandleDependencies_PostAdds(t *testing.T) {
	dm := NewDependencyManager()
	body, _ := json.Marshal(map[string]string{"depends_on": "job-a"})
	req := httptest.NewRequest(http.MethodPost, "/deps?job=job-b", bytes.NewReader(body))
	w := httptest.NewRecorder()
	dm.handleDependencies(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if len(dm.Get("job-b")) != 1 {
		t.Fatal("expected dependency to be recorded")
	}
}

func TestHandleDependencies_MethodNotAllowed(t *testing.T) {
	dm := NewDependencyManager()
	req := httptest.NewRequest(http.MethodPut, "/deps?job=job-b", nil)
	w := httptest.NewRecorder()
	dm.handleDependencies(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
