package checker

import (
	"strings"
	"time"

	"github.com/tofuwu-dev/proxychecker/internal/checks"
	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
)

// WorkerConfig holds the settings a worker needs to do its job.
type WorkerConfig struct {
	TargetURL      string
	Timeout        time.Duration
	Retries        int
	CheckAnonymity bool
	CheckGeo       bool
}

// runWorker processes proxies from the jobs channel until it is closed,
// sending each Result to the results channel when done.
func runWorker(jobs <-chan px.Proxy, results chan<- Result, cfg WorkerConfig) {
	for proxy := range jobs {
		result := checkWithRetry(proxy, cfg)
		results <- result
	}
}

// checkWithRetry runs connectivity check with optional anonymity detection.
func checkWithRetry(proxy px.Proxy, cfg WorkerConfig) Result {
	var last checks.ConnectivityResult

	attempts := 1 + cfg.Retries
	for i := 0; i < attempts; i++ {
		last = checks.Check(proxy, cfg.TargetURL, cfg.Timeout)
		if last.Alive {
			break
		}
		if last.StatusCode != 0 {
			break
		}
	}

	result := buildResult(proxy, last)

	// Only run anonymity check if the proxy is alive — no point checking
	// anonymity on a dead proxy. This keeps dead proxy checks fast.
	if result.Status == StatusAlive && cfg.CheckAnonymity {
		ar := checks.CheckAnonymity(proxy, cfg.Timeout)
		result.Anonymity = ar.Level
	}

	if result.Status == StatusAlive && cfg.CheckGeo {
		geo := checks.CheckGeo(proxy, cfg.Timeout)
		result.Geo = &geo
	}

	return result
}

// buildResult converts a ConnectivityResult into a checker Result,
// determining the Status and formatting the error string.
func buildResult(proxy px.Proxy, cr checks.ConnectivityResult) Result {
	result := Result{
		Proxy:      proxy,
		LatencyMs:  cr.LatencyMs,
		StatusCode: cr.StatusCode,
		CheckedAt:  time.Now(),
	}

	if cr.Alive {
		result.Status = StatusAlive
		return result
	}

	// Distinguish timeouts from other failures — useful for the user to know
	// whether their timeout setting is too aggressive.
	if cr.Err != nil && isTimeout(cr.Err) {
		result.Status = StatusTimeout
	} else {
		result.Status = StatusDead
	}

	if cr.Err != nil {
		result.Error = cr.Err.Error()
	}

	return result
}

// isTimeout checks if an error is a timeout by inspecting the error message.
// Go doesn't have a single timeout error type across all packages, so string
// inspection is the pragmatic approach here.
func isTimeout(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "context deadline")
}
