// Command urlfetch concurrently downloads a list of URLs to local files.
package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"urlfetch/internal/fetcher"
)

func main() {
	outDir := flag.String("out", "downloads", "directory to save downloaded files into")
	timeout := flag.Duration("timeout", 8*time.Second, "per-request timeout, e.g. 5s, 500ms")
	urlFile := flag.String("file", "", "path to a text file with one URL per line (optional)")
	concurrency := flag.Int("concurrency", 0, "max concurrent fetches (0 = unlimited)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [flags] <url1> <url2> ...\n  %s [flags] -file urls.txt\n\nFlags:\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(*outDir, *timeout, *urlFile, *concurrency, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run contains the actual program logic, kept separate from main() so it's
// testable and so os.Exit isn't scattered throughout.
func run(outDir string, timeout time.Duration, urlFile string, concurrency int, args []string) error {
	urls, err := fetcher.CollectURLs(urlFile, args)
	if err != nil {
		return err
	}
	if len(urls) == 0 {
		flag.Usage()
		return fmt.Errorf("no URLs provided (pass them as args or use -file)")
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	start := time.Now()
	results := fetchAll(urls, outDir, timeout, concurrency)
	elapsed := time.Since(start)

	fetcher.PrintSummary(os.Stdout, results, elapsed)

	if fetcher.AnyFailed(results) {
		return fmt.Errorf("one or more URLs failed")
	}
	return nil
}

// fetchAll runs fetcher.Fetch for every URL concurrently, optionally capped
// by a semaphore, and returns results in the same order as urls.
func fetchAll(urls []string, outDir string, timeout time.Duration, concurrency int) []fetcher.Result {
	var wg sync.WaitGroup
	results := make([]fetcher.Result, len(urls))

	var sem chan struct{}
	if concurrency > 0 {
		sem = make(chan struct{}, concurrency)
	}

	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			if sem != nil {
				sem <- struct{}{}
				defer func() { <-sem }()
			}
			results[i] = fetcher.Fetch(u, outDir, timeout)
		}(i, u)
	}

	wg.Wait()
	return results
}

