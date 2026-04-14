// parser.go handles converting raw text lines into structured Proxy values.
//
// Real-world proxy lists are messy. Providers export in different formats,
// users copy-paste from spreadsheets, and nobody agrees on a standard.
// This package absorbs that chaos so the rest of the codebase works with
// clean, typed data.
package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

// ParseResult holds the outcome of parsing an entire proxy list file.
// We return both the successes and failures so the caller can show the
// user a meaningful summary ("loaded 490 proxies, skipped 10 invalid lines").
type ParseResult struct {
	Proxies  []Proxy
	Failures []ParseFailure
}

// ParseFailure records a line we couldn't parse, and why.
// Returning structured errors (not just printing them) lets the caller
// decide how to surface them — as warnings, as a log file, etc.
type ParseFailure struct {
	Line   string
	Reason string
}

// ParseFile reads a file from disk and parses every non-empty line as a proxy.
// Returns both successfully parsed proxies and any lines that failed to parse.
func ParseFile(path string) (*ParseResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	return ParseReader(f)
}

// ParseReader parses proxies from any io.Reader.
// Accepts an interface so callers can pass files, stdin, or in-memory buffers.
func ParseReader(r io.Reader) (*ParseResult, error) {
	result := &ParseResult{}

	scanner := bufio.NewScanner(r)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments (lines starting with # or //)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		proxy, err := ParseLine(line)
		if err != nil {
			result.Failures = append(result.Failures, ParseFailure{
				Line:   line,
				Reason: err.Error(),
			})
			continue
		}

		result.Proxies = append(result.Proxies, proxy)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	return result, nil
}

// ParseLine converts a single raw string into a Proxy.
// Handles four formats:
//
//	protocol://user:pass@host:port
//	protocol://host:port
//	host:port:user:pass
//	host:port
func ParseLine(line string) (Proxy, error) {
	if strings.Contains(line, "://") {
		return parseURLFormat(line)
	}
	return parseColonFormat(line)
}

// parseURLFormat handles lines that contain "://".
func parseURLFormat(line string) (Proxy, error) {
	u, err := url.Parse(line)
	if err != nil {
		return Proxy{}, fmt.Errorf("invalid URL format: %w", err)
	}

	protocol := parseProtocol(u.Scheme)
	if protocol == Unknown {
		return Proxy{}, fmt.Errorf("unsupported protocol %q (want http, https, socks5, socks4)", u.Scheme)
	}

	host, port, err := splitHostPort(u.Host)
	if err != nil {
		return Proxy{}, fmt.Errorf("invalid host:port %q: %w", u.Host, err)
	}

	p := Proxy{
		Host:     host,
		Port:     port,
		Protocol: protocol,
	}

	if u.User != nil {
		p.Username = u.User.Username()
		p.Password, _ = u.User.Password()
	}

	return p, nil
}

// parseColonFormat handles bare host:port and host:port:user:pass lines.
func parseColonFormat(line string) (Proxy, error) {
	// n=4 keeps passwords that contain colons intact.
	parts := strings.SplitN(line, ":", 4)

	switch len(parts) {
	case 2:
		// host:port — no credentials, no explicit protocol
		host, port, err := splitHostPort(line)
		if err != nil {
			return Proxy{}, err
		}
		return Proxy{
			Host:     host,
			Port:     port,
			Protocol: HTTP, // default assumption for bare host:port
		}, nil

	case 4:
		// host:port:username:password
		host := parts[0]
		port := parts[1]
		username := parts[2]
		password := parts[3]

		if err := validateHost(host); err != nil {
			return Proxy{}, err
		}
		if err := validatePort(port); err != nil {
			return Proxy{}, err
		}

		return Proxy{
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
			Protocol: HTTP,
		}, nil

	default:
		return Proxy{}, fmt.Errorf("unrecognized format %q (expected host:port or host:port:user:pass)", line)
	}
}

// parseProtocol maps a raw scheme string to our Protocol type.
func parseProtocol(scheme string) Protocol {
	switch strings.ToLower(scheme) {
	case "http":
		return HTTP
	case "https":
		return HTTPS
	case "socks5":
		return SOCKS5
	case "socks4":
		return SOCKS4
	default:
		return Unknown
	}
}

// splitHostPort splits "host:port" into its parts, stripping IPv6 brackets.
func splitHostPort(hostport string) (host, port string, err error) {
	idx := strings.LastIndex(hostport, ":")
	if idx == -1 {
		return "", "", fmt.Errorf("missing port in %q", hostport)
	}
	host = hostport[:idx]
	port = hostport[idx+1:]

	// Strip brackets from IPv6 addresses: [::1] → ::1
	host = strings.Trim(host, "[]")

	if err := validateHost(host); err != nil {
		return "", "", err
	}
	if err := validatePort(port); err != nil {
		return "", "", err
	}

	return host, port, nil
}

// validateHost checks that the host is non-empty.
// We don't validate reachability — the connectivity check surfaces that naturally.
func validateHost(host string) error {
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("empty host")
	}
	return nil
}

// validatePort checks that the port is non-empty.
// Invalid ports fail at connection time with a clear error.
func validatePort(port string) error {
	if strings.TrimSpace(port) == "" {
		return fmt.Errorf("empty port")
	}
	return nil
}
