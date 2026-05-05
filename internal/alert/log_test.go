package alert_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/alert"
)

func TestLogNotifier_SendWritesFormattedAlert(t *testing.T) {
	var buf bytes.Buffer
	notifier := alert.NewLogNotifier(&buf)

	a := alert.Alert{
		JobName:        "backup",
		Severity:       "warning",
		Message:        "job ran too fast",
		ActualDuration: 5 * time.Second,
		ExpectedMin:    10 * time.Second,
		ExpectedMax:    60 * time.Second,
		OccurredAt:     time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}

	if err := notifier.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	for _, want := range []string{
		"[cronwatch]",
		"job=\"backup\"",
		"severity=warning",
		"job ran too fast",
		"2024-06-01T12:00:00Z",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got: %s", want, output)
		}
	}
}

func TestLogNotifier_DefaultsToStdout(t *testing.T) {
	// Passing nil should not panic; it defaults to os.Stdout.
	notifier := alert.NewLogNotifier(nil)
	if notifier == nil {
		t.Fatal("expected non-nil LogNotifier")
	}
}

func TestLogNotifier_SendReturnsNoError(t *testing.T) {
	var buf bytes.Buffer
	notifier := alert.NewLogNotifier(&buf)

	a := alert.NewDriftAlert("cleanup", 2*time.Second, 5*time.Second, 30*time.Second)
	if err := notifier.Send(a); err != nil {
		t.Errorf("Send should return nil error, got: %v", err)
	}
}
