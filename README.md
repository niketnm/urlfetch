# urlfetch

A small concurrent URL downloader written in Go. Fetches a list of URLs in
parallel, saves each response body to a local file, and prints a summary
report of successes and failures.

## Usage

```bash
go run main.go
```

Downloaded files are written to `downloads/` (git-ignored). Edit the `urls`
slice in `main.go` to change what gets fetched.

## What it demonstrates

- `context.WithTimeout` for per-request deadlines
- `defer` for guaranteed cleanup (`cancel()`, `resp.Body.Close()`, `out.Close()`)
- `sync.WaitGroup` for concurrent fetches
- Streaming response bodies to disk with `io.Copy`
- Structured error handling without panics
