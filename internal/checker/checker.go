package checker

import (
	"sync"

	px "github.com/tofuwu-dev/proxychecker/internal/proxy"
)

// Config holds the top-level configuration for a checker run.
type Config struct {
	Concurrency int
	Worker      WorkerConfig
}

// Run checks all proxies concurrently and returns results via a channel.
//
// The caller receives results as they arrive — they don't have to wait for
// all checks to finish before processing the first result. This is the
// "pipeline" pattern: produce → transform → consume, all running concurrently.
//
// Returning a channel (instead of []Result) means:
//   - Output can start printing before all checks are done
//   - Memory stays flat — we never hold all results in memory at once
//   - The progress bar can update in real time
//
// The returned channel is closed when all workers are done, signalling
// to the receiver that iteration is complete.
func Run(proxies []px.Proxy, cfg Config) <-chan Result {
	jobs := make(chan px.Proxy, len(proxies))
	results := make(chan Result, len(proxies))

	for _, p := range proxies {
		jobs <- p
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(cfg.Concurrency)

	for i := 0; i < cfg.Concurrency; i++ {
		go func() {
			defer wg.Done()
			runWorker(jobs, results, cfg.Worker)
		}()
	}

	// Close results once all workers finish, signalling the receiver that
	// iteration is complete.
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
