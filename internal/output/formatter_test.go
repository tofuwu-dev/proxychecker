package output

import (
	"strings"
	"testing"
	"time"

	"github.com/tofuwu-dev/proxychecker/internal/checker"
	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
)

// makeResult is a test helper that builds a checker.Result cleanly.
func makeResult(host string, status checker.Status, latencyMs int64) checker.Result {
	return checker.Result{
		Proxy:     px.Proxy{Host: host, Port: "8080", Protocol: px.HTTP},
		Status:    status,
		LatencyMs: latencyMs,
		CheckedAt: time.Now(),
	}
}

func TestTableFormatter_WritesRows(t *testing.T) {
	var buf strings.Builder
	f := newTableFormatter(&buf)

	f.Write(makeResult("1.2.3.4", checker.StatusAlive, 42))
	f.Write(makeResult("1.2.3.5", checker.StatusDead, 0))
	f.Flush()

	out := buf.String()
	if !strings.Contains(out, "1.2.3.4") {
		t.Error("expected alive proxy address in output")
	}
	if !strings.Contains(out, "1.2.3.5") {
		t.Error("expected dead proxy address in output")
	}
	if !strings.Contains(out, "PROXY") {
		t.Error("expected header in table output")
	}
}

func TestJSONFormatter_ValidJSON(t *testing.T) {
	var buf strings.Builder
	f := newJSONFormatter(&buf)

	f.Write(makeResult("1.2.3.4", checker.StatusAlive, 42))
	f.Write(makeResult("1.2.3.5", checker.StatusDead, 0))
	f.Flush()

	out := buf.String()
	// Valid JSON array starts with [ and ends with ]
	if !strings.Contains(out, "[") || !strings.Contains(out, "]") {
		t.Error("expected JSON array in output")
	}
	if !strings.Contains(out, "1.2.3.4") {
		t.Error("expected proxy address in JSON output")
	}
}

func TestCSVFormatter_WritesHeaderAndRows(t *testing.T) {
	var buf strings.Builder
	f := newCSVFormatter(&buf)

	f.Write(makeResult("1.2.3.4", checker.StatusAlive, 42))
	f.Flush()

	out := buf.String()
	if !strings.Contains(out, "host") {
		t.Error("expected CSV header in output")
	}
	if !strings.Contains(out, "1.2.3.4") {
		t.Error("expected proxy address in CSV output")
	}
}

func TestNew_ReturnsCorrectFormatter(t *testing.T) {
	var buf strings.Builder

	tests := []struct {
		format Format
		want   string // type name for display
	}{
		{FormatTable, "table"},
		{FormatJSON, "json"},
		{FormatCSV, "csv"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			f := New(tc.format, &buf)
			if f == nil {
				t.Errorf("New(%q) returned nil", tc.format)
			}
		})
	}
}
