// Package proxy defines the Proxy data model and input parser.
package proxy

import "fmt"

// Protocol represents the transport protocol a proxy uses.
// Defined as a named type (not plain string) for compiler-enforced type safety.
type Protocol string

const (
	HTTP    Protocol = "http"
	HTTPS   Protocol = "https"
	SOCKS5  Protocol = "socks5"
	SOCKS4  Protocol = "socks4"
	Unknown Protocol = "unknown"
)

// Proxy holds all connection details for a single proxy entry.
// Username and Password are optional — only set for authenticated proxies.
type Proxy struct {
	Host     string   `json:"host"`
	Port     string   `json:"port"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	Protocol Protocol `json:"protocol"`
}

// Address returns "host:port".
func (p Proxy) Address() string {
	return fmt.Sprintf("%s:%s", p.Host, p.Port)
}

// URL returns the full proxy URL including protocol and optional credentials.
// Examples:
//
//	http://1.2.3.4:8080
//	socks5://user:pass@1.2.3.4:1080
func (p Proxy) URL() string {
	if p.Username != "" && p.Password != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%s",
			p.Protocol, p.Username, p.Password, p.Host, p.Port)
	}
	return fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)
}

// String implements fmt.Stringer.
func (p Proxy) String() string {
	return p.URL()
}
