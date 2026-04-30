package tracker

import (
	"fmt"
	"time"
)

// DriftResult holds the outcome of a drift check for a single job.
type DriftResult struct {
	JobName  string
	Duration time.Duration
	Min      time.Duration
	Max      time.Duration
	Drifted  bool
	Message  string
}

// CheckDrift compares the recorded duration of a job against its expected
// min/max window and returns a DriftResult.
func CheckDrift(rec JobRecord, min, max time.Duration) (DriftResult, error) {
	if rec.Duration == nil {
		return DriftResult{}, fmt.Errorf("job %q has not finished yet", rec.JobName)
	}

	result := DriftResult{
		JobName:  rec.JobName,
		Duration: *rec.Duration,
		Min:      min,
		Max:      max,
	}

	switch {
	case *rec.Duration < min:
		result.Drifted = true
		result.Message = fmt.Sprintf(
			"job %q finished too quickly: %s (min %s)",
			rec.JobName, rec.Duration.Round(time.Millisecond), min,
		)
	case *rec.Duration > max:
		result.Drifted = true
		result.Message = fmt.Sprintf(
			"job %q ran too long: %s (max %s)",
			rec.JobName, rec.Duration.Round(time.Millisecond), max,
		)
	default:
		result.Message = fmt.Sprintf(
			"job %q completed normally in %s",
			rec.JobName, rec.Duration.Round(time.Millisecond),
		)
	}

	return result, nil
}
