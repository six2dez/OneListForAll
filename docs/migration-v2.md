# Migration to v2 Go Platform

## Before
- Build script: `./olfa.sh`
- External dedupe tool required: `duplicut`
- No first-class category pipeline

## Now
- CLI: `go run ./cmd/olfa <subcommand>`
- Dedupe is internal to Go pipeline (chunk sort + merge)
- Config-driven behavior in `configs/pipeline.yml` (JSON-compatible format)
- Category generation by technology/use-case (`short` and `long` variants)

## Command Mapping
- Old build: `./olfa.sh`
- New build all: `go run ./cmd/olfa build --config configs/pipeline.yml --profile all --output-dir dist`
- New build short: `go run ./cmd/olfa build --config configs/pipeline.yml --profile short --output-dir dist`
- New build one category: `go run ./cmd/olfa build --config configs/pipeline.yml --category wordpress --variant short --categories-dir dist/categories`
- New build all categories: `go run ./cmd/olfa build --config configs/pipeline.yml --all-categories --categories-dir dist/categories`
- New validation gate: `go run ./cmd/olfa validate-categories --config configs/pipeline.yml`
- New package: `go run ./cmd/olfa package --config configs/pipeline.yml --output-dir dist`

## New Operational Files
- `configs/pipeline.yml`: sources/filter/build/release/category configuration
- `sources_lock.yml`: pinned source state after updates
- `dict/index.json`: imported source index with detected categories
- `dist/SHA256SUMS`: output checksums
- `dist/categories/*.txt`: generated category lists
