// Package output handles formatting and writing checker results.
package output

import (
	"io"

	"github.com/tofuwu-dev/proxychecker/internal/checker"
)

// Format represents the output format selected by the user.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

// Formatter is the interface every output format implements.
// Write receives results as they stream in from the checker — one at a time —
// and Flush is called once at the end to finalize output (e.g. closing a JSON
// array or flushing a buffered writer).
//
// Accepting results one at a time (not []Result) means output can start
// printing before all checks are done — important for large proxy lists.
type Formatter interface {
	Write(r checker.Result) error
	Flush() error
}

// New returns the Formatter for the requested format.
// w is where output goes — pass os.Stdout for terminal, or a file for export.
func New(format Format, w io.Writer) Formatter {
	switch format {
	case FormatJSON:
		return newJSONFormatter(w)
	case FormatCSV:
		return newCSVFormatter(w)
	default:
		return newTableFormatter(w)
	}
}
