package api

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestLogger(buf *bytes.Buffer) *log.Logger {
	return log.New(buf, "", 0)
}

func TestLoggingMiddleware_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := LoggingMiddleware(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Errorf("expected log to contain method GET, got: %s", output)
	}
	if !strings.Contains(output, "/healthz") {
		t.Errorf("expected log to contain path /healthz, got: %s", output)
	}
	if !strings.Contains(output, "200") {
		t.Errorf("expected log to contain status 200, got: %s", output)
	}
}

func TestLoggingMiddleware_CapturesNon200(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := LoggingMiddleware(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !strings.Contains(buf.String(), "404") {
		t.Errorf("expected log to contain status 404, got: %s", buf.String())
	}
}

func TestRecoveryMiddleware_HandlesPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := RecoveryMiddleware(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}))

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if !strings.Contains(buf.String(), "recovered from panic") {
		t.Errorf("expected panic log, got: %s", buf.String())
	}
}

func TestRecoveryMiddleware_PassthroughNoPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := RecoveryMiddleware(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
