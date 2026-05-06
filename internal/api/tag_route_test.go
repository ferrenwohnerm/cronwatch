package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTagRouter() http.Handler {
	mux := http.NewServeMux()
	tm := NewTagManager()
	RegisterTagRoutes(mux, tm)
	return mux
}

func TestTagRoute_GetEmpty(t *testing.T) {
	router := newTagRouter()
	req := httptest.NewRequest(http.MethodGet, "/tags", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out map[string][]string
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty map, got %v", out)
	}
}

func TestTagRoute_PostAndGet(t *testing.T) {
	mux := http.NewServeMux()
	tm := NewTagManager()
	RegisterTagRoutes(mux, tm)

	body, _ := json.Marshal(map[string]interface{}{"job": "report", "tags": []string{"weekly", "email"}})
	postReq := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewReader(body))
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusOK {
		t.Fatalf("POST expected 200, got %d", postRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/tags", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)

	var out map[string][]string
	json.NewDecoder(getRec.Body).Decode(&out)
	if len(out["report"]) != 2 {
		t.Fatalf("expected 2 tags for report, got %v", out)
	}
}
