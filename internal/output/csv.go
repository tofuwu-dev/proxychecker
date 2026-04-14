package output

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/tofuwu-dev/proxychecker/internal/checker"
)

// csvFormatter writes results as CSV — one row per proxy.
// CSV is useful for loading results into spreadsheets or other scripts.
type csvFormatter struct {
	w      *csv.Writer
	header bool
}

func newCSVFormatter(w io.Writer) *csvFormatter {
	return &csvFormatter{w: csv.NewWriter(w)}
}

func (f *csvFormatter) Write(r checker.Result) error {
	if !f.header {
		f.w.Write([]string{"host", "port", "protocol", "status", "latency_ms", "error"})
		f.header = true
	}

	f.w.Write([]string{
		r.Proxy.Host,
		r.Proxy.Port,
		string(r.Proxy.Protocol),
		string(r.Status),
		fmt.Sprintf("%d", r.LatencyMs),
		r.Error,
	})

	return nil
}

// Flush flushes the CSV writer's internal buffer to the underlying io.Writer.
// encoding/csv buffers writes internally for performance — without Flush,
// the last rows may never reach the output.
func (f *csvFormatter) Flush() error {
	f.w.Flush()
	return f.w.Error()
}
