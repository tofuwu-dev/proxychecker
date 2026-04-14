package output

import (
	"bufio"
	"fmt"
	"os"

	"github.com/tofuwu-dev/proxychecker/internal/checker"
)

// Export writes alive proxies to a file in host:port format.
// Only alive proxies are written — dead and timeout proxies are skipped.
//
// bufio.Writer batches small writes into larger ones, reducing syscall overhead.
// Always call Flush() after all writes or the last chunk may be lost.
func Export(results []checker.Result, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create export file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, r := range results {
		if r.Status == checker.StatusAlive {
			fmt.Fprintf(w, "%s\n", r.Proxy.Address())
		}
	}

	return w.Flush()
}
