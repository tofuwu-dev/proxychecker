// Package progress provides a live progress bar for proxy checking.
package progress

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
	"github.com/tofuwu-dev/proxychecker/internal/checker"
)

// Bar wraps progressbar with proxy-check specific tracking.
// It counts alive, dead, and timeout results in real time so the user
// sees a running tally while checks are still in progress.
type Bar struct {
	bar     *progressbar.ProgressBar
	alive   int
	dead    int
	timeout int
	total   int
}

// New creates a Bar for a run of total proxy checks.
// We write to stderr so the bar doesn't pollute stdout —
// this keeps `--format json` pipeable without bar characters corrupting the JSON.
func New(total int) *Bar {
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription("checking"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		// Show elapsed time and estimated time remaining.
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintln(os.Stderr)
		}),
	)

	return &Bar{bar: bar, total: total}
}

// Add records one completed result and advances the bar.
// Call this once per result as it arrives from the checker channel.
func (b *Bar) Add(r checker.Result) {
	switch r.Status {
	case checker.StatusAlive:
		b.alive++
	case checker.StatusTimeout:
		b.timeout++
	default:
		b.dead++
	}

	// Update the description with the running tally.
	// The user sees "checking | 12 alive | 5 dead" updating in real time.
	b.bar.Describe(fmt.Sprintf(
		"checking [green]%d alive[reset] [red]%d dead[reset] [yellow]%d timeout[reset]",
		b.alive, b.dead, b.timeout,
	))

	b.bar.Add(1)
}

// Finish completes the bar. Call after all results have been processed.
func (b *Bar) Finish() {
	b.bar.Finish()
}
