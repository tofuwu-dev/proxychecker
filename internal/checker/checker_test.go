package checker

import (
	"testing"
	"time"

	"github.com/tofuwu-dev/proxychecker/internal/checks"
	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
)

// defaultConfig returns a WorkerConfig suitable for tests —
// short timeout, no retries, no real network calls needed for unit tests.
func defaultConfig() Config {
	return Config{
		Concurrency: 3,
		Worker: WorkerConfig{
			TargetURL: "http://127.0.0.1:1", // always fails fast — port 1 is reserved
			Timeout:   1 * time.Second,
			Retries:   0,
		},
	}
}

func TestRun_ReturnsResultForEachProxy(t *testing.T) {
	proxies := []px.Proxy{
		{Host: "1.2.3.4", Port: "8080", Protocol: px.HTTP},
		{Host: "1.2.3.5", Port: "8080", Protocol: px.HTTP},
		{Host: "1.2.3.6", Port: "8080", Protocol: px.HTTP},
	}

	results := Run(proxies, defaultConfig())

	// Collect all results from the channel.
	// The range exits when the channel is closed — which happens after
	// all workers finish. If Run() has a bug and never closes the channel,
	// this test will hang and timeout, which is itself a useful signal.
	var got []Result
	for r := range results {
		got = append(got, r)
	}

	if len(got) != len(proxies) {
		t.Errorf("expected %d results, got %d", len(proxies), len(got))
	}
}

func TestRun_AllDeadWithBadTarget(t *testing.T) {
	proxies := []px.Proxy{
		{Host: "1.2.3.4", Port: "8080", Protocol: px.HTTP},
		{Host: "1.2.3.5", Port: "8080", Protocol: px.HTTP},
	}

	results := Run(proxies, defaultConfig())

	for r := range results {
		if r.Status == StatusAlive {
			t.Errorf("proxy %s should be dead with unreachable target", r.Proxy.Address())
		}
	}
}

func TestRun_EmptyProxyList(t *testing.T) {
	results := Run([]px.Proxy{}, defaultConfig())

	var got []Result
	for r := range results {
		got = append(got, r)
	}

	if len(got) != 0 {
		t.Errorf("expected 0 results for empty input, got %d", len(got))
	}
}

func TestRun_ConcurrencyHigherThanProxyCount(t *testing.T) {
	// 2 proxies, 10 workers — workers outnumber jobs.
	// The extra workers should exit cleanly when jobs channel is closed.
	proxies := []px.Proxy{
		{Host: "1.2.3.4", Port: "8080", Protocol: px.HTTP},
		{Host: "1.2.3.5", Port: "8080", Protocol: px.HTTP},
	}

	cfg := Config{
		Concurrency: 10,
		Worker: WorkerConfig{
			TargetURL: "http://127.0.0.1:1",
			Timeout:   1 * time.Second,
			Retries:   0,
		},
	}

	results := Run(proxies, cfg)

	var got []Result
	for r := range results {
		got = append(got, r)
	}

	if len(got) != 2 {
		t.Errorf("expected 2 results, got %d", len(got))
	}
}

func TestBuildResult_AliveProxy(t *testing.T) {
	proxy := px.Proxy{Host: "1.2.3.4", Port: "8080", Protocol: px.HTTP}
	cr := checks_ConnectivityResult(true, 42, 200, nil)

	result := buildResult(proxy, cr)

	if result.Status != StatusAlive {
		t.Errorf("Status: got %q, want %q", result.Status, StatusAlive)
	}
	if result.LatencyMs != 42 {
		t.Errorf("LatencyMs: got %d, want 42", result.LatencyMs)
	}
}

// checks_ConnectivityResult is a test helper that constructs a
// checks.ConnectivityResult inline. Named with package prefix for clarity.
func checks_ConnectivityResult(alive bool, latencyMs int64, statusCode int, err error) checks.ConnectivityResult {
	return checks.ConnectivityResult{
		Alive:      alive,
		LatencyMs:  latencyMs,
		StatusCode: statusCode,
		Err:        err,
	}
}
