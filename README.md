# urlfetch

A small concurrent URL downloader written in Go. Fetches a list of URLs in
parallel, saves each response body to a local file, and prints a summary
report of successes and failures.

## Project layout

```
urlfetch/
├── cmd/
│   └── urlfetch/
│       └── main.go       # CLI entrypoint: flag parsing, wiring, exit codes
├── internal/
│   └── fetcher/
│       ├── fetcher.go       # core Fetch() logic + Result type
│       ├── urls.go          # CollectURLs(): merges args + file input
│       ├── report.go        # PrintSummary() / AnyFailed()
│       └── fetcher_test.go  # unit tests
├── go.mod
├── .gitignore
└── README.md
```

`internal/` is a Go compiler convention: packages under it can only be
imported by code inside this module, signaling that `fetcher` is an
implementation detail, not a public API. `cmd/urlfetch/` follows the standard
layout for a project that may one day host multiple binaries under `cmd/`.

## Usage

Pass URLs directly as arguments:

```bash
go run ./cmd/urlfetch https://go.dev/ https://github.com/niketnm https://www.lipsum.com/
```

Or put them in a file, one per line (`#` lines are treated as comments):

```bash
go run ./cmd/urlfetch -file urls.txt
```

You can mix both — args and `-file` are combined.

### Flags

| Flag           | Default      | Description                                  |
|----------------|--------------|-----------------------------------------------|
| `-out`         | `downloads`  | Directory to save downloaded files into       |
| `-timeout`     | `8s`         | Per-request timeout (e.g. `5s`, `500ms`)      |
| `-file`        | (none)       | Path to a text file with one URL per line     |
| `-concurrency` | `0`          | Max concurrent fetches (`0` = unlimited)      |

Example with everything:

```bash
go run ./cmd/urlfetch -out results -timeout 5s -concurrency 3 -file urls.txt
```

See all flags: `go run ./cmd/urlfetch -h`

Downloaded files are written to the output dir (git-ignored by default).
The program exits with a non-zero status if any URL failed — useful for
scripting or CI.

### Building a binary

```bash
go build -o urlfetch ./cmd/urlfetch
./urlfetch https://go.dev/
```

### Running tests

```bash
go test ./...
```

## What it demonstrates

- `context.WithTimeout` for per-request deadlines
- `defer` for guaranteed cleanup (`cancel()`, `resp.Body.Close()`, `out.Close()`)
- `sync.WaitGroup` for concurrent fetches, plus a semaphore for optional concurrency limiting
- Streaming response bodies to disk with `io.Copy`
- Structured error handling without panics
- The `flag` package for CLI arguments
- A standard Go project layout (`cmd/` + `internal/`) with separated, testable logic

