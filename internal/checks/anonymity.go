package checks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
	"github.com/tofuwu-dev/proxychecker/pkg/httputil"
)

// AnonymityLevel represents how much information a proxy leaks to the target.
type AnonymityLevel string

const (
	// AnonymityElite means the target cannot detect proxy usage at all.
	AnonymityElite AnonymityLevel = "elite"

	// AnonymityAnonymous means the target knows a proxy is in use but
	// cannot see the client's real IP address.
	AnonymityAnonymous AnonymityLevel = "anonymous"

	// AnonymityTransparent means the target can see the client's real IP
	// via forwarding headers added by the proxy.
	AnonymityTransparent AnonymityLevel = "transparent"

	// AnonymityUnknown means the check could not be completed.
	AnonymityUnknown AnonymityLevel = "unknown"
)

// headersEndpoint echoes back all request headers the server received.
// This is what lets us see exactly what the proxy is forwarding.
const headersEndpoint = "https://httpbin.org/headers"

// AnonymityResult holds the outcome of an anonymity check.
type AnonymityResult struct {
	Level   AnonymityLevel
	Headers map[string]string // headers the target actually received
	Err     error
}

// httpbinHeadersResponse matches the JSON shape of httpbin.org/headers.
type httpbinHeadersResponse struct {
	Headers map[string]string `json:"headers"`
}

// CheckAnonymity sends a request through the proxy and inspects which
// headers the target server received to classify the anonymity level.
func CheckAnonymity(p px.Proxy, timeout time.Duration) AnonymityResult {
	client, err := httputil.NewClient(p, timeout)
	if err != nil {
		return AnonymityResult{
			Level: AnonymityUnknown,
			Err:   fmt.Errorf("build client: %w", err),
		}
	}

	headers, err := fetchHeaders(client)
	if err != nil {
		return AnonymityResult{
			Level: AnonymityUnknown,
			Err:   fmt.Errorf("fetch headers: %w", err),
		}
	}

	level := classifyAnonymity(headers)

	return AnonymityResult{
		Level:   level,
		Headers: headers,
	}
}

// fetchHeaders calls httpbin.org/headers through the proxy and returns
// the headers map from the response JSON.
func fetchHeaders(client *http.Client) (map[string]string, error) {
	resp, err := client.Get(headersEndpoint)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read at most 32KB — httpbin responses are small, but we defend
	// against a misbehaving proxy that streams garbage indefinitely.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result httpbinHeadersResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Headers, nil
}

// classifyAnonymity inspects the headers the target received and returns
// the appropriate anonymity level.
//
// Detection logic:
//  1. If any header contains or forwards the real client IP → transparent
//  2. If proxy-identifying headers are present but no real IP → anonymous
//  3. If none of those headers are present → elite
func classifyAnonymity(headers map[string]string) AnonymityLevel {
	// Headers that reveal the client's real IP address.
	// A transparent proxy adds these so the target knows who is really connecting.
	ipLeakHeaders := []string{
		"X-Forwarded-For",
		"X-Real-Ip",
		"X-Client-Ip",
		"Client-Ip",
		"True-Client-Ip",
		"X-Originating-Ip",
	}

	// Headers that reveal proxy usage without leaking the real IP.
	// An anonymous proxy strips IP headers but still announces itself.
	proxyHeaders := []string{
		"Via",
		"Proxy-Connection",
		"X-Proxy-Id",
		"Forwarded",
	}

	// HTTP headers are case-insensitive; normalize before map lookups.
	normalized := make(map[string]string, len(headers))
	for k, v := range headers {
		normalized[strings.ToLower(k)] = v
	}

	for _, h := range ipLeakHeaders {
		if _, found := normalized[strings.ToLower(h)]; found {
			return AnonymityTransparent
		}
	}

	for _, h := range proxyHeaders {
		if _, found := normalized[strings.ToLower(h)]; found {
			return AnonymityAnonymous
		}
	}

	return AnonymityElite
}
