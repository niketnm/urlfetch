package fetcher

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CollectURLs merges URLs passed as positional args with URLs read from a
// file (one per line; blank lines and lines starting with # are skipped).
// filePath may be empty, in which case only args are used.
func CollectURLs(filePath string, args []string) ([]string, error) {
	urls := append([]string{}, args...)

	if filePath == "" {
		return urls, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading url file: %w", err)
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
		return nil, fmt.Errorf("scanning url file: %w", err)
	}

	return urls, nil
}
