package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookNotifier_SendSuccess(t *testing.T) {
	var received webhookPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected content-type: %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wn := NewWebhookNotifier(ts.URL)
	a := Alert{
		JobName:    "nightly-export",
		Level:      LevelError,
		Message:    "ran too long",
		OccurredAt: time.Now(),
	}
	if err := wn.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.JobName != "nightly-export" {
		t.Errorf("expected job_name 'nightly-export', got %q", received.JobName)
	}
	if received.Level != "ERROR" {
		t.Errorf("expected level 'ERROR', got %q", received.Level)
	}
}

func TestWebhookNotifier_SendNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	wn := NewWebhookNotifier(ts.URL)
	a := Alert{JobName: "job", Level: LevelWarn, Message: "slow", OccurredAt: time.Now()}
	err := wn.Send(a)
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestWebhookNotifier_SendConnectionRefused(t *testing.T) {
	wn := NewWebhookNotifier("http://127.0.0.1:1")
	a := Alert{JobName: "job", Level: LevelError, Message: "gone", OccurredAt: time.Now()}
	if err := wn.Send(a); err == nil {
		t.Fatal("expected error on refused connection, got nil")
	}
}
