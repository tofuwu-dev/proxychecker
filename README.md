# proxychecker

A fast, concurrent proxy health checker written in Go. Tests HTTP, HTTPS, SOCKS4, and SOCKS5 proxies for connectivity, latency, anonymity level, and geo-location.

## Features

- **Concurrent checking** — configurable worker pool, checks hundreds of proxies in seconds
- **Multiple protocols** — HTTP, HTTPS, SOCKS4, SOCKS5 with optional authentication
- **Anonymity detection** — classifies proxies as elite, anonymous, or transparent
- **Geo-location** — resolves country, city, and ISP for each proxy's outbound IP
- **Flexible input** — accepts `host:port`, `protocol://host:port`, `host:port:user:pass`
- **Multiple output formats** — colored table, JSON, CSV
- **Export** — writes working proxies to a file for immediate reuse
- **Single binary** — no runtime dependencies, download and run

## Installation

### Download binary

Grab the latest release for your platform from the [releases page](https://github.com/tofuwu-dev/proxychecker/releases).

```bash
# Linux / WSL
curl -L https://github.com/tofuwu-dev/proxychecker/releases/latest/download/proxychecker-linux-amd64 -o proxychecker
chmod +x proxychecker

# macOS (Apple Silicon)
curl -L https://github.com/tofuwu-dev/proxychecker/releases/latest/download/proxychecker-darwin-arm64 -o proxychecker
chmod +x proxychecker

# macOS (Intel)
curl -L https://github.com/tofuwu-dev/proxychecker/releases/latest/download/proxychecker-darwin-amd64 -o proxychecker
chmod +x proxychecker

# Windows (download manually from releases page and run the .exe)
```

### Build from source

```bash
git clone https://github.com/tofuwu-dev/proxychecker.git
cd proxychecker
make build
```

Requires Go 1.22+.

## Usage

```bash
proxychecker check [file] [flags]
```

### Basic check

```bash
proxychecker check proxies.txt
```

### All flags

| Flag | Default | Description |
|---|---|---|
| `--concurrency, -c` | `50` | Number of concurrent workers |
| `--timeout, -t` | `10s` | Timeout per proxy |
| `--retries, -r` | `1` | Retries on failure |
| `--target` | `httpbin.org/ip` | Custom URL to test against |
| `--anonymity` | `false` | Check anonymity level |
| `--geo` | `false` | Check geo-location |
| `--format, -f` | `table` | Output format: `table`, `json`, `csv` |
| `--export` | — | Write working proxies to file |
| `--silent` | `false` | Suppress output (use with `--export`) |

### Examples

```bash
# Fast check with custom concurrency
proxychecker check proxies.txt --concurrency 100 --timeout 5s

# Full check with anonymity and geo
proxychecker check proxies.txt --anonymity --geo

# Export only working proxies
proxychecker check proxies.txt --export working_proxies.txt

# JSON output — pipe to jq
proxychecker check proxies.txt --format json | jq '.[] | select(.status=="alive")'

# Test against a specific target
proxychecker check proxies.txt --target https://example.com

# Silent mode for scripting
proxychecker check proxies.txt --silent --export working_proxies.txt
```

## Input formats

All of these are valid input lines:

```
# Bare host:port (defaults to HTTP)
192.168.1.1:8080

# Full URL
http://192.168.1.1:8080
socks5://192.168.1.1:1080

# With credentials
http://user:pass@192.168.1.1:8080
192.168.1.1:8080:user:pass

# Comments and blank lines are skipped
```

## Output

### Table (default)

```
PROXY                     PROTOCOL STATUS     MS     ERROR
───────────────────────── ──────── ────────── ────── ─────
1.2.3.4:8080              http     alive      342    [elite] New York, US (Acme ISP)
1.2.3.5:8080              http     dead       45     connection refused
1.2.3.6:8080              http     timeout    10001
```

### JSON

```bash
proxychecker check proxies.txt --format json
```

```json
[
  {
    "proxy": { "host": "1.2.3.4", "port": "8080", "protocol": "http" },
    "status": "alive",
    "latency_ms": 342,
    "anonymity": "elite",
    "geo": {
      "IP": "1.2.3.4",
      "Country": "United States",
      "CountryCode": "US",
      "City": "New York",
      "ISP": "Acme ISP"
    },
    "checked_at": "2026-04-14T06:41:19Z"
  }
]
```

## Development

```bash
# Run unit tests
make test

# Run with race detector
go test ./... -race -short

# Check test coverage
make coverage

# Build binaries for all platforms
make release
```

## Architecture

```
cmd/proxychecker/     — CLI entrypoint (Cobra)
internal/
  proxy/              — proxy struct + input parser
  checks/             — connectivity, anonymity, geo checks
  checker/            — concurrent worker pool
  output/             — table, JSON, CSV formatters
  progress/           — progress bar
pkg/
  httputil/           — HTTP client builder (HTTP + SOCKS5)
testdata/             — sample proxy lists for testing
```

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a PR.

## License

MIT — see [LICENSE](LICENSE) for details.