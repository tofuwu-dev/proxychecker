# Contributing to proxychecker

Thank you for considering a contribution. This document explains how to get started.

## Ways to Contribute

- **Bug reports** — open an issue describing the problem, steps to reproduce, and your OS/Go version
- **Feature requests** — open an issue describing the use case before writing any code
- **Code contributions** — see the workflow below
- **Documentation** — fixing typos, clarifying examples, improving the README

## Development Setup

Requirements: Go 1.22+, WSL or Linux/macOS recommended (Windows has known network test issues).

```bash
git clone https://github.com/tofuwu-dev/proxychecker
cd proxychecker
go mod tidy
make test
```

## Project Structure

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

## Pull Request Workflow

1. Fork the repository and create a branch from `main`
2. Branch naming: `feat/your-feature` or `fix/your-bug`
3. Write tests for any new behaviour
4. Make sure all checks pass before opening a PR:

```bash
make test   # all unit tests
make vet    # static analysis
gofmt -l .  # formatting check (should print nothing)
```

5. Open a pull request with a clear description of what changed and why

## Code Style

- Follow standard Go conventions — run `gofmt` before committing
- Exported symbols must have godoc comments starting with the symbol name
- Internal helpers should have a single-line comment explaining *why*, not *what*
- Keep functions focused — if a function needs more than one paragraph to explain, consider splitting it
- Table-driven tests are preferred for testing multiple input cases

## Commit Messages

Use the conventional commit format:

```
feat: add SOCKS4 proxy support
fix: handle IPv6 addresses in parser
docs: improve README install instructions
test: add edge cases for empty proxy list
refactor: simplify anonymity header detection
chore: update dependencies
```

## Adding a New Check

The check system is designed to be extended. To add a new check (e.g. SSL certificate validation):

1. Create `internal/checks/yourcheck.go` with a `CheckYourThing(p px.Proxy, timeout time.Duration) YourResult` function
2. Add `YourResult` fields to `internal/checker/result.go`
3. Add a `CheckYourThing bool` flag to `WorkerConfig` in `internal/checker/worker.go`
4. Call it in `checkWithRetry` after the alive check, same pattern as anonymity and geo
5. Wire the flag in `cmd/proxychecker/main.go`
6. Display it in `internal/output/table.go`

## Reporting Security Issues

Do not open a public issue for security vulnerabilities. Email the maintainer directly instead.

## License

By contributing, you agree that your contributions will be licensed under the same MIT license as this project.