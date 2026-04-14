package checks

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
	"github.com/tofuwu-dev/proxychecker/pkg/httputil"
)

// geoEndpoint returns the ip-api.com URL for a given IP address.
// We use http (not https) because ip-api.com's free tier doesn't support
// HTTPS. This is fine — geo data isn't sensitive.
const geoEndpoint = "http://ip-api.com/json/%s?fields=status,country,countryCode,city,isp,org,query"

// GeoResult holds location and network information for a proxy's outbound IP.
type GeoResult struct {
	IP          string
	Country     string
	CountryCode string
	City        string
	ISP         string
	Org         string
	Err         error
}

// ipAPIResponse matches the JSON shape returned by ip-api.com.
type ipAPIResponse struct {
	Status      string `json:"status"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
	Query       string `json:"query"` // the IP that was looked up
}

// CheckGeo looks up the geographic location of a proxy's outbound IP.
// Routes the lookup through the proxy itself so we get the IP it presents
// to the outside world, not our real IP.
func CheckGeo(p px.Proxy, timeout time.Duration) GeoResult {
	ip, err := resolveHost(p.Host)
	if err != nil {
		return GeoResult{Err: fmt.Errorf("resolve host: %w", err)}
	}

	client, err := httputil.NewClient(p, timeout)
	if err != nil {
		return GeoResult{Err: fmt.Errorf("build client: %w", err)}
	}

	return fetchGeo(client, ip)
}

// resolveHost returns the first IP address for a hostname.
func resolveHost(host string) (string, error) {
	addrs, err := net.LookupHost(host)
	if err != nil {
		return "", fmt.Errorf("lookup %q: %w", host, err)
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("no addresses for %q", host)
	}
	return addrs[0], nil
}

// fetchGeo calls ip-api.com and parses the response into a GeoResult.
func fetchGeo(client *http.Client, ip string) GeoResult {
	url := fmt.Sprintf(geoEndpoint, ip)

	resp, err := client.Get(url)
	if err != nil {
		return GeoResult{Err: fmt.Errorf("geo request failed: %w", err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if err != nil {
		return GeoResult{Err: fmt.Errorf("read geo response: %w", err)}
	}

	var apiResp ipAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return GeoResult{Err: fmt.Errorf("decode geo response: %w", err)}
	}

	if apiResp.Status != "success" {
		return GeoResult{Err: fmt.Errorf("geo lookup failed for IP %s", ip)}
	}

	return GeoResult{
		IP:          apiResp.Query,
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		City:        apiResp.City,
		ISP:         apiResp.ISP,
		Org:         apiResp.Org,
	}
}
