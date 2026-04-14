package output

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/tofuwu-dev/proxychecker/internal/checker"
)

// tableFormatter writes a colored row per result to the terminal.
// It prints a header once, then one line per proxy as results arrive.
type tableFormatter struct {
	w             io.Writer
	headerPrinted bool
}

func newTableFormatter(w io.Writer) *tableFormatter {
	return &tableFormatter{w: w}
}

// Write prints one result row. The first call also prints the header.
// Colors give instant visual feedback — green for alive, red for dead,
// yellow for timeout — so users can scan hundreds of results at a glance.
func (f *tableFormatter) Write(r checker.Result) error {
	if !f.headerPrinted {
		fmt.Fprintf(f.w, "\n%-25s %-8s %-10s %-6s %s\n",
			"PROXY", "PROTOCOL", "STATUS", "MS", "ERROR")
		fmt.Fprintf(f.w, "%-25s %-8s %-10s %-6s %s\n",
			"─────────────────────────", "────────", "──────────", "──────", "─────")
		f.headerPrinted = true
	}

	addr := fmt.Sprintf("%-25s", r.Proxy.Address())
	proto := fmt.Sprintf("%-8s", r.Proxy.Protocol)
	latency := fmt.Sprintf("%-6d", r.LatencyMs)

	switch r.Status {
	case checker.StatusAlive:
		anon := ""
		if r.Anonymity != "" {
			anon = fmt.Sprintf("[%s]", r.Anonymity)
		}
		status := color.GreenString("%-10s", "alive")
		fmt.Fprintf(f.w, "%s %s %s %s %s\n", addr, proto, status, latency, anon)

	case checker.StatusTimeout:
		status := color.YellowString("%-10s", "timeout")
		fmt.Fprintf(f.w, "%s %s %s %s\n", addr, proto, status, latency)

	default:
		status := color.RedString("%-10s", "dead")
		errMsg := r.Error
		if len(errMsg) > 50 {
			errMsg = errMsg[:50] + "..."
		}
		fmt.Fprintf(f.w, "%s %s %s %s %s\n", addr, proto, status, latency, errMsg)
	}

	return nil
}

// Flush is a no-op for the table formatter — we write each row immediately
// as it arrives. It exists to satisfy the Formatter interface.
func (f *tableFormatter) Flush() error {
	fmt.Fprintln(f.w)
	return nil
}
