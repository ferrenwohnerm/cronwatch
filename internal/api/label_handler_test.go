package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleLabels_MethodNotAllowed(t *testing.T) {
	lm := NewLabelManager()
	h := handleLabels(lm)

	req := httptest.NewRequest(http.MethodPatch, "/labels?job=x", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleLabels_PostBadJSON(t *testing.T) {
	lm := NewLabelManager()
	h := handleLabels(lm)

	req := httptest.NewRequest(http.MethodPost, "/labels?job=x", bytes.NewBufferString("not-json"))
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad JSON, got %d", rec.Code)
	}
}

func TestHandleLabels_GetReturnsJSON(t *testing.T) {
	lm := NewLabelManager()
	lm.Set("nightly", "region", "us-east-1")
	h := handleLabels(lm)

	req := httptest.NewRequest(http.MethodGet, "/labels?job=nightly", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if result["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %s", result["region"])
	}
}

func TestHandleLabels_ContentTypeJSON(t *testing.T) {
	lm := NewLabelManager()
	h := handleLabels(lm)

	req := httptest.NewRequest(http.MethodGet, "/labels?job=any", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}
