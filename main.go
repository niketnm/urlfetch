package main

import (
	"context"
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

// urls to fetch — edit this list, or later swap it for reading from a file/flag.
var urls = []string{
	"https://www.lipsum.com/",
	"https://github.com/niketnm",
	"https://go.dev/",
	"https://this-domain-does-not-exist-12345.com/", // intentionally broken, to show error handling
}

const (
	outputDir     = "downloads"
	perRequestTTL = 8 * time.Second
)

func main() {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Println("failed to create output dir:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	results := make([]result, len(urls))

	start := time.Now()

	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			results[i] = fetchAndSave(u)
		}(i, u)
	}

	wg.Wait()

	printSummary(results, time.Since(start))
}

// fetchAndSave downloads one URL with a timeout and writes the body to a file.
func fetchAndSave(rawURL string) result {
	ctx, cancel := context.WithTimeout(context.Background(), perRequestTTL)
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

	path := filepath.Join(outputDir, filenameFor(rawURL))
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
