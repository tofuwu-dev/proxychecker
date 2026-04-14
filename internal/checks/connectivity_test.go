package checks

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
)

// startTestServer spins up a minimal TCP server on a random local port.
// We avoid httptest.Server entirely because it has a deadlock bug on Windows
// when closing connections. This approach gives us full control over the
// server lifecycle.
//
// handler is a standard http.Handler — same as what httptest.Server uses
// internally, just wired up manually so we control the shutdown.
func startTestServer(t *testing.T, handler http.Handler) string {
	t.Helper()

	// net.Listen picks a random available port when you pass ":0"
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}

	server := &http.Server{Handler: handler}

	// Run the server in a goroutine so it doesn't block the test.
	// A goroutine is a lightweight thread — `go` keyword launches it.
	go server.Serve(listener)

	// t.Cleanup registers a function to run when the test finishes.
	// It's the modern alternative to defer — works correctly even when
	// called from helper functions like this one.
	t.Cleanup(func() {
		listener.Close()
		server.Close()
	})

	// listener.Addr() returns the actual address including the random port
	// the OS assigned — e.g. "127.0.0.1:54321"
	return fmt.Sprintf("http://%s", listener.Addr().String())
}

// plainClient returns an http.Client with no proxy and no keep-alives.
// Used in tests where we just want to verify probe() logic, not proxy routing.
func plainClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
		Timeout: 5 * time.Second,
	}
}

func TestProbe_Success(t *testing.T) {
	url := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	result := probe(plainClient(), url)

	if !result.Alive {
		t.Errorf("expected Alive=true, got false. err: %v", result.Err)
	}
	if result.LatencyMs < 0 {
		t.Errorf("LatencyMs should be >= 0, got %d", result.LatencyMs)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode: got %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestProbe_Non2xxMarkedDead(t *testing.T) {
	url := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))

	result := probe(plainClient(), url)

	if result.Alive {
		t.Error("expected Alive=false for 403 response")
	}
	if result.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode: got %d, want 403", result.StatusCode)
	}
}

func TestProbe_UnreachableServer(t *testing.T) {
	// Port 1 is reserved and always refuses connections — reliable dead target
	result := probe(plainClient(), "http://127.0.0.1:1")

	if result.Alive {
		t.Error("expected Alive=false for unreachable server")
	}
	if result.Err == nil {
		t.Error("expected non-nil Err for unreachable server")
	}
}

func TestCheck_InvalidProtocol(t *testing.T) {
	p := px.Proxy{Host: "1.2.3.4", Port: "8080", Protocol: px.Unknown}
	result := Check(p, DefaultTarget, 5*time.Second)

	if result.Alive {
		t.Error("expected Alive=false for unknown protocol")
	}
}

func TestCheck_RealProxy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	result := probe(plainClient(), DefaultTarget)

	if !result.Alive {
		t.Errorf("expected live httpbin.org to be reachable, got err: %v", result.Err)
	}
}
