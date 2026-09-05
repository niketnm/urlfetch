package fetcher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectURLs_ArgsOnly(t *testing.T) {
	urls, err := CollectURLs("", []string{"https://a.example", "https://b.example"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(urls) != 2 {
		t.Fatalf("expected 2 urls, got %d", len(urls))
	}
}

func TestCollectURLs_FileAndArgsMerge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "urls.txt")

	content := "https://from-file.example\n# a comment\n\nhttps://another-from-file.example\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	urls, err := CollectURLs(path, []string{"https://from-args.example"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"https://from-args.example", "https://from-file.example", "https://another-from-file.example"}
	if len(urls) != len(want) {
		t.Fatalf("expected %d urls, got %d: %v", len(want), len(urls), urls)
	}
	for i, u := range want {
		if urls[i] != u {
			t.Errorf("url[%d] = %q, want %q", i, urls[i], u)
		}
	}
}

func TestCollectURLs_MissingFile(t *testing.T) {
	_, err := CollectURLs("/nonexistent/path/urls.txt", nil)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestFilenameFor(t *testing.T) {
	cases := map[string]string{
		"https://go.dev/":            "go.dev_index.html",
		"https://github.com/niketnm": "github.com_niketnm.html",
		"http://example.com":         "example.com.html",
	}
	for input, want := range cases {
		if got := filenameFor(input); got != want {
			t.Errorf("filenameFor(%q) = %q, want %q", input, got, want)
		}
	}
}
