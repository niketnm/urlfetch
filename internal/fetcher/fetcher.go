// Package fetcher implements concurrent HTTP downloading of URLs to local files.
package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Result holds the outcome of fetching one URL.
type Result struct {
	URL      string
	FilePath string
	Bytes    int
	Duration time.Duration
	Err      error
}

// Fetch downloads a single URL with the given timeout and writes the response
// body to a file inside outDir. The filename is derived from the URL.
func Fetch(rawURL string, outDir string, timeout time.Duration) Result {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return Result{URL: rawURL, Err: err}
	}

	client := &http.Client{}
	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: rawURL, Err: err, Duration: time.Since(start)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{URL: rawURL, Err: fmt.Errorf("bad status: %s", resp.Status), Duration: time.Since(start)}
	}

	path := filepath.Join(outDir, filenameFor(rawURL))
	out, err := os.Create(path)
	if err != nil {
		return Result{URL: rawURL, Err: err, Duration: time.Since(start)}
	}
	defer out.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return Result{URL: rawURL, Err: err, Duration: time.Since(start)}
	}

	return Result{
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
