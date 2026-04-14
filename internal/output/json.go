package output

import (
	"encoding/json"
	"io"

	"github.com/tofuwu-dev/proxychecker/internal/checker"
)

// jsonFormatter collects all results then writes a single JSON array.
// We buffer everything because streaming partial JSON is invalid —
// a JSON array must be complete to be parseable by other tools.
type jsonFormatter struct {
	w       io.Writer
	results []checker.Result
}

func newJSONFormatter(w io.Writer) *jsonFormatter {
	return &jsonFormatter{w: w}
}

// Write buffers the result — actual writing happens in Flush.
func (f *jsonFormatter) Write(r checker.Result) error {
	f.results = append(f.results, r)
	return nil
}

// Flush encodes all buffered results as a JSON array.
// json.MarshalIndent produces human-readable output with 2-space indentation.
// The `json:"..."` tags on Result fields control the key names.
func (f *jsonFormatter) Flush() error {
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(f.results)
}
