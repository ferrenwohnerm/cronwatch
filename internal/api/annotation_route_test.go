package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newAnnotationRouter() (*http.ServeMux, *AnnotationManager) {
	mux := http.NewServeMux()
	m := NewAnnotationManager(10)
	RegisterAnnotationRoutes(mux, m)
	return mux, m
}

func TestAnnotationRoute_GetEmpty(t *testing.T) {
	mux, _ := newAnnotationRouter()
	req := httptest.NewRequest(http.MethodGet, "/annotations?job=myjob", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var list []Annotation
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list")
	}
}

func TestAnnotationRoute_PostAndGet(t *testing.T) {
	mux, _ := newAnnotationRouter()

	body := bytes.NewBufferString(`{"note":"maintenance window"}`)
	postReq := httptest.NewRequest(http.MethodPost, "/annotations?job=cleanup", body)
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", postRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/annotations?job=cleanup", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	var list []Annotation
	json.NewDecoder(getRec.Body).Decode(&list)
	if len(list) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(list))
	}
	if list[0].JobName != "cleanup" {
		t.Errorf("unexpected job name: %s", list[0].JobName)
	}
}
