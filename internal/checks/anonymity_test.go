package checks

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
)

// startHeadersServer spins up a local server that mimics httpbin.org/headers.
// It echoes back whatever headers it receives as JSON — exactly what our
// anonymity checker expects to parse.
func startHeadersServer(t *testing.T, extraHeaders map[string]string) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Build the headers map from what we received + any injected extras.
			// This lets each test simulate what a proxy would forward.
			headers := make(map[string]string)
			for k, v := range r.Header {
				headers[k] = v[0]
			}
			for k, v := range extraHeaders {
				headers[k] = v
			}

			resp := httpbinHeadersResponse{Headers: headers}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}),
	}

	go server.Serve(listener)
	t.Cleanup(func() {
		listener.Close()
		server.Close()
	})

	return fmt.Sprintf("http://%s", listener.Addr().String())
}

func TestClassifyAnonymity_Elite(t *testing.T) {
	// No proxy-identifying headers at all → elite
	headers := map[string]string{
		"Host":       "example.com",
		"User-Agent": "Go-http-client/1.1",
	}

	level := classifyAnonymity(headers)

	if level != AnonymityElite {
		t.Errorf("got %q, want %q", level, AnonymityElite)
	}
}

func TestClassifyAnonymity_Anonymous(t *testing.T) {
	// Via header present but no IP leak → anonymous
	headers := map[string]string{
		"Host": "example.com",
		"Via":  "1.1 proxy-server",
	}

	level := classifyAnonymity(headers)

	if level != AnonymityAnonymous {
		t.Errorf("got %q, want %q", level, AnonymityAnonymous)
	}
}

func TestClassifyAnonymity_Transparent(t *testing.T) {
	// X-Forwarded-For present → transparent
	headers := map[string]string{
		"Host":            "example.com",
		"X-Forwarded-For": "203.0.113.42",
		"Via":             "1.1 proxy-server",
	}

	level := classifyAnonymity(headers)

	if level != AnonymityTransparent {
		t.Errorf("got %q, want %q", level, AnonymityTransparent)
	}
}

func TestClassifyAnonymity_CaseInsensitive(t *testing.T) {
	// Headers can arrive in any case — our check must be case-insensitive
	headers := map[string]string{
		"x-forwarded-for": "203.0.113.42", // lowercase
	}

	level := classifyAnonymity(headers)

	if level != AnonymityTransparent {
		t.Errorf("case-insensitive check failed: got %q, want %q",
			level, AnonymityTransparent)
	}
}

func TestCheckAnonymity_InvalidProtocol(t *testing.T) {
	p := px.Proxy{Host: "1.2.3.4", Port: "8080", Protocol: px.Unknown}
	result := CheckAnonymity(p, 5*time.Second)

	if result.Level != AnonymityUnknown {
		t.Errorf("expected unknown level for invalid protocol, got %q", result.Level)
	}
	if result.Err == nil {
		t.Error("expected non-nil error for invalid protocol")
	}
}

func TestCheckAnonymity_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test with a real proxy if you have one — otherwise this just verifies
	// the full pipeline compiles and runs without panicking using DefaultClient.
	result := CheckAnonymity(
		px.Proxy{Host: "1.2.3.4", Port: "8080", Protocol: px.HTTP},
		5*time.Second,
	)

	// We don't assert a specific level — we just confirm the function
	// returns a structured result without panicking.
	t.Logf("anonymity level: %s, err: %v", result.Level, result.Err)
}
