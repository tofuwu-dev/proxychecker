// Package httputil provides HTTP client construction for proxied requests.
package httputil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"

	// px alias avoids collision with golang.org/x/net/proxy.
	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
)

// NewClient builds an *http.Client that routes all traffic through p.
// The timeout applies to the entire request lifecycle — dial, TLS handshake,
// and response body read combined.
//
// Returns an error if the proxy configuration is invalid or unsupported.
func NewClient(p px.Proxy, timeout time.Duration) (*http.Client, error) {
	switch p.Protocol {
	case px.HTTP, px.HTTPS:
		return newHTTPProxyClient(p, timeout)
	case px.SOCKS5, px.SOCKS4:
		return newSOCKS5Client(p, timeout)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", p.Protocol)
	}
}

// newHTTPProxyClient configures an http.Client for HTTP/HTTPS proxies.
func newHTTPProxyClient(p px.Proxy, timeout time.Duration) (*http.Client, error) {
	proxyURL, err := url.Parse(p.URL())
	if err != nil {
		return nil, fmt.Errorf("parse proxy url: %w", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}

// newSOCKS5Client configures an http.Client for SOCKS5 proxies.
// Transport doesn't speak SOCKS5 natively, so we plug in a custom dialer.
func newSOCKS5Client(p px.Proxy, timeout time.Duration) (*http.Client, error) {
	var auth *proxy.Auth
	if p.Username != "" {
		auth = &proxy.Auth{
			User:     p.Username,
			Password: p.Password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", p.Address(), auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("create socks5 dialer: %w", err)
	}

	transport := &http.Transport{
		DialContext: dialContext(dialer),
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}

// dialContext adapts proxy.Dialer to the DialContext signature http.Transport expects.
// Checks context cancellation before dialing so an already-expired context returns immediately.
func dialContext(d proxy.Dialer) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return d.Dial(network, addr)
	}
}
