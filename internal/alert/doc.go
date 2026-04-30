// Package alert provides the alerting subsystem for cronwatch.
//
// It defines the Alert type, the Notifier interface, and a Manager that
// fans out alerts to one or more registered Notifier implementations.
//
// Built-in implementations:
//
//   - LogNotifier  — writes alerts to the standard logger (useful as a
//     fallback or during development).
//   - WebhookNotifier — POSTs a JSON payload to a configurable HTTP endpoint,
//     suitable for integrating with Slack incoming webhooks, PagerDuty, or
//     any custom HTTP receiver.
//
// Usage:
//
//	log := &alert.LogNotifier{}
//	wh  := alert.NewWebhookNotifier("https://hooks.example.com/cronwatch")
//	mgr := alert.NewManager(log, wh)
//
//	a := alert.NewDriftAlert("daily-backup", alert.LevelWarn, 42.3)
//	if err := mgr.Send(a); err != nil {
//		log.Printf("alert delivery error: %v", err)
//	}
package alert
