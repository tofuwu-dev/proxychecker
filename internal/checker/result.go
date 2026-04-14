// Package checker orchestrates concurrent proxy health checks.
package checker

import (
	"time"

	"github.com/tofuwu-dev/proxychecker/internal/checks"
	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
)

// Status represents the outcome of a proxy check.
type Status string

const (
	StatusAlive   Status = "alive"
	StatusDead    Status = "dead"
	StatusTimeout Status = "timeout"
)

// Result holds the complete health check outcome for a single proxy.
// This is the central data structure that flows from workers through
// the results channel to the output formatters.
type Result struct {
	Proxy      px.Proxy              `json:"proxy"`
	Status     Status                `json:"status"`
	LatencyMs  int64                 `json:"latency_ms"`
	StatusCode int                   `json:"status_code,omitempty"`
	Anonymity  checks.AnonymityLevel `json:"anonymity,omitempty"`
	Geo        *checks.GeoResult     `json:"geo,omitempty"`
	Error      string                `json:"error,omitempty"`
	CheckedAt  time.Time             `json:"checked_at"`
}
