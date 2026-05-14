package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnvironmentManager_SetAndGet(t *testing.T) {
	m := NewEnvironmentManager()
	m.Set("backup", "S3_BUCKET", "my-bucket")
	m.Set("backup", "REGION", "us-east-1")

	env, ok := m.Get("backup")
	if !ok {
		t.Fatal("expected env to exist")
	}
	if env["S3_BUCKET"] != "my-bucket" {
		t.Errorf("expected my-bucket, got %s", env["S3_BUCKET"])
	}
	if env["REGION"] != "us-east-1" {
		t.Errorf("expected us-east-1, got %s", env["REGION"])
	}
}

func TestEnvironmentManager_GetUnknown(t *testing.T) {
	m := NewEnvironmentManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected false for unknown job")
	}
}

func TestEnvironmentManager_Delete(t *testing.T) {
	m := NewEnvironmentManager()
	m.Set("job1", "KEY", "val")
	m.Delete("job1")
	_, ok := m.Get("job1")
	if ok {
		t.Error("expected env to be deleted")
	}
}

func TestEnvironmentManager_Snapshot(t *testing.T) {
	m := NewEnvironmentManager()
	m.Set("j1", "A", "1")
	m.Set("j2", "B", "2")

	snap := m.Snapshot()
	if len(snap) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(snap))
	}
	// Mutating snapshot must not affect manager
	snap["j1"]["A"] = "changed"
	env, _ := m.Get("j1")
	if env["A"] != "1" {
		t.Error("snapshot mutation affected manager")
	}
}

func TestHandleEnvironment_PostAndGet(t *testing.T) {
	m := NewEnvironmentManager()

	body, _ := json.Marshal(map[string]string{"FOO": "bar"})
	req := httptest.NewRequest(http.MethodPost, "/environment?job=myjob", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	m.HandleEnvironment(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/environment?job=myjob", nil)
	rr2 := httptest.NewRecorder()
	m.HandleEnvironment(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr2.Code)
	}
	var result map[string]string
	json.NewDecoder(rr2.Body).Decode(&result)
	if result["FOO"] != "bar" {
		t.Errorf("expected bar, got %s", result["FOO"])
	}
}

func TestHandleEnvironment_MissingJob(t *testing.T) {
	m := NewEnvironmentManager()
	req := httptest.NewRequest(http.MethodGet, "/environment", nil)
	rr := httptest.NewRecorder()
	m.HandleEnvironment(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleEnvironment_MethodNotAllowed(t *testing.T) {
	m := NewEnvironmentManager()
	req := httptest.NewRequest(http.MethodPatch, "/environment?job=x", nil)
	rr := httptest.NewRecorder()
	m.HandleEnvironment(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHandleEnvironment_Delete(t *testing.T) {
	m := NewEnvironmentManager()
	m.Set("job", "K", "V")
	req := httptest.NewRequest(http.MethodDelete, "/environment?job=job", nil)
	rr := httptest.NewRecorder()
	m.HandleEnvironment(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	_, ok := m.Get("job")
	if ok {
		t.Error("expected job to be deleted")
	}
}
