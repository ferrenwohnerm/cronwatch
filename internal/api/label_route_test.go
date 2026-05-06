package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newLabelRouter() (*http.ServeMux, *LabelManager) {
	mux := http.NewServeMux()
	lm := NewLabelManager()
	RegisterLabelRoutes(mux, lm)
	return mux, lm
}

func TestLabelRoute_GetEmpty(t *testing.T) {
	mux, _ := newLabelRouter()
	req := httptest.NewRequest(http.MethodGet, "/labels?job=myjob", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestLabelRoute_PostAndGet(t *testing.T) {
	mux, _ := newLabelRouter()
	body, _ := json.Marshal(map[string]string{"env": "staging", "owner": "alice"})

	req := httptest.NewRequest(http.MethodPost, "/labels?job=deploy", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/labels?job=deploy", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	var result map[string]string
	json.NewDecoder(rec2.Body).Decode(&result)
	if result["env"] != "staging" {
		t.Errorf("expected env=staging, got %s", result["env"])
	}
}

func TestLabelRoute_Delete(t *testing.T) {
	mux, lm := newLabelRouter()
	lm.Set("job", "k", "v")

	req := httptest.NewRequest(http.MethodDelete, "/labels?job=job", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, ok := lm.Get("job")
	if ok {
		t.Error("expected labels to be deleted after DELETE request")
	}
}

func TestLabelRoute_MissingJobParam(t *testing.T) {
	mux, _ := newLabelRouter()
	req := httptest.NewRequest(http.MethodGet, "/labels", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing job param, got %d", rec.Code)
	}
}
