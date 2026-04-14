package httputil

import (
	"testing"
	"time"

	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
)

// TestNewClient_ReturnsClient checks that NewClient produces a non-nil client
// for each supported protocol without erroring.
//
// We're not making real network calls here — that belongs in integration tests.
// Unit tests only verify that the construction logic is correct.
func TestNewClient_ReturnsClient(t *testing.T) {
	tests := []struct {
		name    string
		proxy   px.Proxy
		wantErr bool
	}{
		{
			name:  "http proxy no auth",
			proxy: px.Proxy{Host: "1.2.3.4", Port: "8080", Protocol: px.HTTP},
		},
		{
			name:  "https proxy with auth",
			proxy: px.Proxy{Host: "1.2.3.4", Port: "8080", Username: "u", Password: "p", Protocol: px.HTTPS},
		},
		{
			name:  "socks5 no auth",
			proxy: px.Proxy{Host: "1.2.3.4", Port: "1080", Protocol: px.SOCKS5},
		},
		{
			name:  "socks5 with auth",
			proxy: px.Proxy{Host: "1.2.3.4", Port: "1080", Username: "u", Password: "p", Protocol: px.SOCKS5},
		},
		{
			name:    "unknown protocol errors",
			proxy:   px.Proxy{Host: "1.2.3.4", Port: "8080", Protocol: px.Unknown},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.proxy, 10*time.Second)

			if tc.wantErr {
				if err == nil {
					t.Errorf("NewClient() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewClient() unexpected error: %v", err)
			}
			if client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

// TestNewClient_TimeoutIsSet verifies the timeout is applied to the client.
// This matters because a zero timeout means "wait forever" — a silent footgun.
func TestNewClient_TimeoutIsSet(t *testing.T) {
	p := px.Proxy{Host: "1.2.3.4", Port: "8080", Protocol: px.HTTP}
	want := 5 * time.Second

	client, err := NewClient(p, want)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Timeout != want {
		t.Errorf("Timeout: got %v, want %v", client.Timeout, want)
	}
}
