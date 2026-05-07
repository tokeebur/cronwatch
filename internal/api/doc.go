// Package api exposes a lightweight HTTP API for cronwatch.
//
// Endpoints:
//
//	GET /healthz        — liveness probe, returns {"status":"ok"}
//	GET /status         — returns a 24-hour summary for every known job
//	GET /history?job=X  — returns the last 20 run entries for job X
//
// The server is intentionally minimal; it is meant for dashboards and
// simple health checks, not as a full management API.
package api
