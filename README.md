# polymarket

Scans Polymarket's public Gamma API for multi-outcome event groups (e.g. "who
will win X") where the group's Yes-leg ask prices sum to less than $1. Exactly
one leg resolves Yes, so buying every Yes leg guarantees a $1 payout — a sum
below $1 is a same-platform arbitrage edge, before fees and slippage.

Read-only: it only reports numbers for you to review. It does not place
orders, hold credentials, or identify traders.

## Usage

```bash
go run ./cmd --limit 300 --min-edge 0.02
```

Flags:

| Flag | Default | Meaning |
|---|---|---|
| `--limit` | `200` | number of events to fetch |
| `--min-edge` | `0.02` | minimum gross edge (e.g. `0.02` = 2¢ on the dollar) to report |
| `--csv` | `output/results.csv` | path to write results as CSV (empty to skip) |

Results print to stdout as a sorted table and are also written to
`output/results.csv` (directory created automatically). `output/` is
gitignored — it holds generated data, not source.

## Development

```bash
go vet ./...
go test ./...
```

## Changelog

- 2026-09-21: Restructured into `cmd/` (`cmd/main.go`, `cmd/main_test.go`); run via `go run ./cmd ...` instead of `go run .`.
- 2026-09-21: Added `output/` folder — results now default to `output/results.csv`, auto-created on run.
- 2026-09-21: Initial port from the `travis` repo scaffold into this standalone module (`go.mod` as `github.com/champ-isaac/polymarket`).
