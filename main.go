package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// result holds the outcome of fetching one URL.
type result struct {
	URL      string
	FilePath string
	Bytes    int
	Duration time.Duration
	Err      error
}

func main() {
	// Flags
	outDir := flag.String("out", "downloads", "directory to save downloaded files into")
	timeout := flag.Duration("timeout", 8*time.Second, "per-request timeout, e.g. 5s, 500ms")
	urlFile := flag.String("file", "", "path to a text file with one URL per line (optional)")
	concurrency := flag.Int("concurrency", 0, "max concurrent fetches (0 = unlimited)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [flags] <url1> <url2> ...\n  %s [flags] -file urls.txt\n\nFlags:\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	urls, err := collectURLs(*urlFile, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "error: no URLs provided (pass them as args or use -file)")
		flag.Usage()
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "failed to create output dir:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	results := make([]result, len(urls))

	// Optional concurrency limiter: a buffered channel used as a semaphore.
	var sem chan struct{}
	if *concurrency > 0 {
		sem = make(chan struct{}, *concurrency)
	}

	start := time.Now()

	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			if sem != nil {
				sem <- struct{}{}        // acquire slot
				defer func() { <-sem }() // release slot
			}
			results[i] = fetchAndSave(u, *outDir, *timeout)
		}(i, u)
	}

	wg.Wait()

	printSummary(results, time.Since(start))

	// Exit with non-zero status if anything failed — useful for scripting/CI.
	for _, r := range results {
		if r.Err != nil {
			os.Exit(1)
		}
	}
}

// collectURLs merges URLs passed as positional args with URLs read from a
// file (one per line, blank lines and lines starting with # are skipped).
func collectURLs(filePath string, args []string) ([]string, error) {
	urls := append([]string{}, args...)

	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading -file: %w", err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			urls = append(urls, line)
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("scanning -file: %w", err)
		}
	}

	return urls, nil
}

// fetchAndSave downloads one URL with a timeout and writes the body to a file.
func fetchAndSave(rawURL string, outDir string, timeout time.Duration) result {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return result{URL: rawURL, Err: err}
	}

	client := &http.Client{}
	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return result{URL: rawURL, Err: err, Duration: time.Since(start)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result{URL: rawURL, Err: fmt.Errorf("bad status: %s", resp.Status), Duration: time.Since(start)}
	}

	path := filepath.Join(outDir, filenameFor(rawURL))
	out, err := os.Create(path)
	if err != nil {
		return result{URL: rawURL, Err: err, Duration: time.Since(start)}
	}
	defer out.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return result{URL: rawURL, Err: err, Duration: time.Since(start)}
	}

	return result{
		URL:      rawURL,
		FilePath: path,
		Bytes:    int(n),
		Duration: time.Since(start),
	}
}

// filenameFor turns a URL into a safe-ish local filename.
func filenameFor(rawURL string) string {
	name := strings.NewReplacer(
		"https://", "",
		"http://", "",
		"/", "_",
		":", "_",
	).Replace(rawURL)
	if name == "" || strings.HasSuffix(name, "_") {
		name += "index"
	}
	return name + ".html"
}

// printSummary prints a simple report of what succeeded and failed.
func printSummary(results []result, total time.Duration) {
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Fetched %d URLs in %v\n", len(results), total.Round(time.Millisecond))
	fmt.Println(strings.Repeat("-", 60))

	okCount := 0
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("FAIL  %-45s %v\n", r.URL, r.Err)
			continue
		}
		okCount++
		fmt.Printf("OK    %-45s %6d bytes  %v -> %s\n",
			r.URL, r.Bytes, r.Duration.Round(time.Millisecond), r.FilePath)
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%d/%d succeeded\n", okCount, len(results))
}
