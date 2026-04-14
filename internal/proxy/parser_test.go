// parser_test.go — tests for the proxy parser.
//
// In Go, test files live alongside the code they test (same package/folder)
// and are named with a _test.go suffix. The `go test` command knows to
// compile and run them but exclude them from the final binary.
//
// We use `package proxy` (not `package proxy_test`) — this gives us access
// to unexported helpers like parseColonFormat and parseURLFormat directly,
// which is useful for unit testing internals without making them public.
package proxy

import (
	"strings"
	"testing"
)

// TestParseLine uses Go's "table-driven test" pattern — the idiomatic way
// to test many inputs without repeating boilerplate.
//
// The idea: define a slice of test cases, each with a name, input, and
// expected output. Loop over them and call t.Run() for each. This gives you:
//   - Clean failure messages: "FAIL TestParseLine/socks5_with_credentials"
//   - Easy to add new cases without writing new test functions
//   - All cases visible in one place
func TestParseLine(t *testing.T) {
	// Each test case is an anonymous struct. Go lets you define struct types
	// inline like this — no need to declare a named type just for tests.
	tests := []struct {
		name      string
		input     string
		wantProxy Proxy // what we expect on success
		wantErr   bool  // true if we expect an error
	}{
		// --- Bare host:port ---
		{
			name:  "bare host:port defaults to http",
			input: "192.168.1.1:8080",
			wantProxy: Proxy{
				Host:     "192.168.1.1",
				Port:     "8080",
				Protocol: HTTP,
			},
		},

		// --- URL format without credentials ---
		{
			name:  "http url no credentials",
			input: "http://198.51.100.10:8080",
			wantProxy: Proxy{
				Host:     "198.51.100.10",
				Port:     "8080",
				Protocol: HTTP,
			},
		},
		{
			name:  "socks5 url no credentials",
			input: "socks5://198.51.100.30:1080",
			wantProxy: Proxy{
				Host:     "198.51.100.30",
				Port:     "1080",
				Protocol: SOCKS5,
			},
		},
		{
			name:  "https url no credentials",
			input: "https://198.51.100.20:443",
			wantProxy: Proxy{
				Host:     "198.51.100.20",
				Port:     "443",
				Protocol: HTTPS,
			},
		},

		// --- URL format with credentials ---
		{
			name:  "http url with credentials",
			input: "http://user:password@198.51.100.50:8080",
			wantProxy: Proxy{
				Host:     "198.51.100.50",
				Port:     "8080",
				Username: "user",
				Password: "password",
				Protocol: HTTP,
			},
		},
		{
			name:  "socks5 url with credentials",
			input: "socks5://proxyuser:secret123@198.51.100.60:1080",
			wantProxy: Proxy{
				Host:     "198.51.100.60",
				Port:     "1080",
				Username: "proxyuser",
				Password: "secret123",
				Protocol: SOCKS5,
			},
		},

		// --- Colon-delimited with credentials ---
		{
			name:  "host:port:user:pass format",
			input: "198.51.100.80:8080:myuser:mypass",
			wantProxy: Proxy{
				Host:     "198.51.100.80",
				Port:     "8080",
				Username: "myuser",
				Password: "mypass",
				Protocol: HTTP,
			},
		},

		// --- Error cases ---
		{
			name:    "unsupported protocol",
			input:   "ftp://198.51.100.90:21",
			wantErr: true,
		},
		{
			name:    "completely invalid input",
			input:   "notaproxy",
			wantErr: true,
		},
		{
			name:    "missing host",
			input:   ":8080",
			wantErr: true,
		},
	}

	// t.Run creates a sub-test with its own name. If one fails, others still run.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// The underscore `_` discards the second return value (error) when
			// we know we won't need it, or when we're checking it separately.
			got, err := ParseLine(tc.input)

			// Check error expectation first
			if tc.wantErr {
				if err == nil {
					// t.Errorf marks the test as failed but continues running.
					// t.Fatalf would stop the test immediately.
					t.Errorf("ParseLine(%q) expected error, got nil", tc.input)
				}
				return // no need to check the proxy value if we expected an error
			}

			if err != nil {
				t.Fatalf("ParseLine(%q) unexpected error: %v", tc.input, err)
			}

			// Compare each field explicitly — Go structs are comparable with ==
			// but custom error messages are more helpful when a test fails.
			if got.Host != tc.wantProxy.Host {
				t.Errorf("Host: got %q, want %q", got.Host, tc.wantProxy.Host)
			}
			if got.Port != tc.wantProxy.Port {
				t.Errorf("Port: got %q, want %q", got.Port, tc.wantProxy.Port)
			}
			if got.Protocol != tc.wantProxy.Protocol {
				t.Errorf("Protocol: got %q, want %q", got.Protocol, tc.wantProxy.Protocol)
			}
			if got.Username != tc.wantProxy.Username {
				t.Errorf("Username: got %q, want %q", got.Username, tc.wantProxy.Username)
			}
			if got.Password != tc.wantProxy.Password {
				t.Errorf("Password: got %q, want %q", got.Password, tc.wantProxy.Password)
			}
		})
	}
}

// TestParseReader tests the full file parsing pipeline using an in-memory
// string instead of a real file — this makes the test fast and self-contained.
//
// strings.NewReader turns a string into an io.Reader, which is exactly what
// ParseReader accepts. This is the power of interface-based design — you can
// swap a real file for a string in tests with zero code changes to the parser.
func TestParseReader(t *testing.T) {
	input := `
# comment line — should be skipped
192.168.1.1:8080
http://10.0.0.1:3128
socks5://user:pass@10.0.0.2:1080

// another comment style
notaproxy
`
	result, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseReader unexpected error: %v", err)
	}

	// We expect 3 valid proxies and 1 failure ("notaproxy")
	wantProxies := 3
	wantFailures := 1

	if len(result.Proxies) != wantProxies {
		t.Errorf("Proxies: got %d, want %d", len(result.Proxies), wantProxies)
		for i, p := range result.Proxies {
			t.Logf("  proxy[%d]: %s", i, p)
		}
	}

	if len(result.Failures) != wantFailures {
		t.Errorf("Failures: got %d, want %d", len(result.Failures), wantFailures)
		for i, f := range result.Failures {
			t.Logf("  failure[%d]: %q → %s", i, f.Line, f.Reason)
		}
	}
}

// TestProxyURL tests the URL() method on Proxy.
// Small focused tests like this catch regressions when you change
// the URL formatting logic later.
func TestProxyURL(t *testing.T) {
	tests := []struct {
		name  string
		proxy Proxy
		want  string
	}{
		{
			name:  "no credentials",
			proxy: Proxy{Host: "1.2.3.4", Port: "8080", Protocol: HTTP},
			want:  "http://1.2.3.4:8080",
		},
		{
			name:  "with credentials",
			proxy: Proxy{Host: "1.2.3.4", Port: "1080", Username: "u", Password: "p", Protocol: SOCKS5},
			want:  "socks5://u:p@1.2.3.4:1080",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.proxy.URL()
			if got != tc.want {
				t.Errorf("URL(): got %q, want %q", got, tc.want)
			}
		})
	}
}
