package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookNotifier sends alerts as JSON POST requests to a URL.
type WebhookNotifier struct {
	URL    string
	Client *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with a default timeout.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL: url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

type webhookPayload struct {
	JobName    string `json:"job_name"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	OccurredAt string `json:"occurred_at"`
}

// Send implements Notifier by POSTing the alert as JSON.
func (w *WebhookNotifier) Send(a Alert) error {
	payload := webhookPayload{
		JobName:    a.JobName,
		Level:      string(a.Level),
		Message:    a.Message,
		OccurredAt: a.OccurredAt.Format(time.RFC3339),
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	resp, err := w.Client.Post(w.URL, "application/json", bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d from webhook", resp.StatusCode)
	}
	return nil
}
