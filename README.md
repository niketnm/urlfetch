# urlfetch

A small concurrent URL downloader written in Go. Fetches a list of URLs in
parallel, saves each response body to a local file, and prints a summary
report of successes and failures.

## Usage

Pass URLs directly as arguments:

```bash
go run main.go https://go.dev/ https://github.com/niketnm https://www.lipsum.com/
```

Or put them in a file, one per line (`#` lines are treated as comments):

```bash
go run main.go -file urls.txt
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
go run main.go -out results -timeout 5s -concurrency 3 -file urls.txt
```

See all flags: `go run main.go -h`

Downloaded files are written to the output dir (git-ignored by default).
The program exits with a non-zero status if any URL failed — useful for
scripting or CI.

### Building a binary

```bash
go build -o urlfetch .
./urlfetch https://go.dev/
```

## What it demonstrates

- `context.WithTimeout` for per-request deadlines
- `defer` for guaranteed cleanup (`cancel()`, `resp.Body.Close()`, `out.Close()`)
- `sync.WaitGroup` for concurrent fetches
- Streaming response bodies to disk with `io.Copy`
- Structured error handling without panics