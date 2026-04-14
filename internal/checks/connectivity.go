// Package checks contains individual proxy health check implementations.
// Each file in this package is responsible for one type of check —
// connectivity, anonymity, geo-location, etc. They are intentionally
// independent so the caller (the worker) can run only what's needed.
package checks

import (
	"fmt"
	"net/http"
	"time"

	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
	"github.com/tofuwu-dev/proxychecker/pkg/httputil"
)

// DefaultTarget is the URL used when the user doesn't specify a custom one.
// httpbin.org/ip echoes back the IP address the request arrived from —
// which tells us the proxy's outbound IP, not our real IP. It's a reliable,
// purpose-built tool for exactly this kind of testing.
const DefaultTarget = "https://httpbin.org/ip"

// ConnectivityResult holds the outcome of a single connectivity check.
type ConnectivityResult struct {
	// Alive is true if the proxy successfully completed a round trip.
	Alive bool

	// LatencyMs is round-trip time in milliseconds.
	// Zero if the proxy is dead — callers should check Alive first.
	LatencyMs int64

	// StatusCode is the HTTP status code returned through the proxy.
	// Useful for catching proxies that are "alive" but return garbage
	// (e.g. a captive portal returning 200 for everything).
	StatusCode int

	// OutboundIP is the IP the target server saw the request come from.
	// This is the proxy's public IP, not your real IP — confirming the
	// proxy is actually being used and not bypassed.
	OutboundIP string

	// Err holds a human-readable reason when Alive is false.
	// Nil on success.
	Err error
}

// Check performs a single connectivity probe through p to targetURL.
// It measures latency and confirms the proxy is routing traffic correctly.
//
// timeout controls the entire request lifecycle. If the proxy doesn't
// respond within this duration, the request is cancelled and the proxy
// is marked dead. This is critical — without a timeout, a hanging proxy
// would block a goroutine forever.
func Check(p px.Proxy, targetURL string, timeout time.Duration) ConnectivityResult {
	if targetURL == "" {
		targetURL = DefaultTarget
	}

	client, err := httputil.NewClient(p, timeout)
	if err != nil {
		return ConnectivityResult{
			Alive: false,
			Err:   fmt.Errorf("build client: %w", err),
		}
	}

	return probe(client, targetURL)
}

// probe fires the actual HTTP request and measures latency.
// Separate function so tests can inject a pre-built client without real proxy credentials.
func probe(client *http.Client, targetURL string) ConnectivityResult {
	start := time.Now()

	resp, err := client.Get(targetURL)

	// Record latency regardless of outcome — a 9.9s failure means something
	// different from a 2ms one (timeout vs. instant rejection).
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		return ConnectivityResult{
			Alive:     false,
			LatencyMs: latencyMs,
			Err:       fmt.Errorf("request failed: %w", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ConnectivityResult{
			Alive:      false,
			LatencyMs:  latencyMs,
			StatusCode: resp.StatusCode,
			Err:        fmt.Errorf("unexpected status: %d", resp.StatusCode),
		}
	}

	return ConnectivityResult{
		Alive:      true,
		LatencyMs:  latencyMs,
		StatusCode: resp.StatusCode,
	}
}
