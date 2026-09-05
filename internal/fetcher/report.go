package fetcher

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// PrintSummary writes a human-readable report of results to w.
func PrintSummary(w io.Writer, results []Result, total time.Duration) {
	sep := strings.Repeat("-", 60)

	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "Fetched %d URLs in %v\n", len(results), total.Round(time.Millisecond))
	fmt.Fprintln(w, sep)

	okCount := 0
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "FAIL  %-45s %v\n", r.URL, r.Err)
			continue
		}
		okCount++
		fmt.Fprintf(w, "OK    %-45s %6d bytes  %v -> %s\n",
			r.URL, r.Bytes, r.Duration.Round(time.Millisecond), r.FilePath)
	}

	fmt.Fprintln(w, sep)
	fmt.Fprintf(w, "%d/%d succeeded\n", okCount, len(results))
}

// AnyFailed reports whether any result has an error.
func AnyFailed(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}
