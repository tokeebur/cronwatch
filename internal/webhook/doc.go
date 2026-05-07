// Package webhook implements outbound HTTP webhook notifications.
//
// When a monitored cron job fails, the Notifier marshals a Payload
// describing the failure and POSTs it as JSON to the configured URL.
// Custom HTTP headers (e.g. Authorization) and a configurable timeout
// are supported.
//
// Example configuration (cronwatch.yaml):
//
//	webhook:
//	  url: "https://hooks.example.com/cronwatch"
//	  timeout: 5s
//	  headers:
//	    Authorization: "Bearer <token>"
package webhook
