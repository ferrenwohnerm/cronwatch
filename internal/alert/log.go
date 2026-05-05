package alert

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// LogNotifier sends alerts to a log output (stdout or a file).
type LogNotifier struct {
	logger *log.Logger
}

// NewLogNotifier creates a LogNotifier writing to the given writer.
// If w is nil, os.Stdout is used.
func NewLogNotifier(w io.Writer) *LogNotifier {
	if w == nil {
		w = os.Stdout
	}
	return &LogNotifier{
		logger: log.New(w, "[cronwatch] ", 0),
	}
}

// Send writes the alert to the log output.
func (l *LogNotifier) Send(a Alert) error {
	l.logger.Printf(
		"ALERT job=%q severity=%s drift=%s expected_min=%s expected_max=%s occurred_at=%s message=%q",
		a.JobName,
		a.Severity,
		formatDuration(a.ActualDuration),
		formatDuration(a.ExpectedMin),
		formatDuration(a.ExpectedMax),
		a.OccurredAt.Format(time.RFC3339),
		a.Message,
	)
	return nil
}

func formatDuration(d interface{ String() string }) string {
	if d == nil {
		return "N/A"
	}
	return fmt.Sprintf("%v", d)
}
