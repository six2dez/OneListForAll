# OneListForAll

Wordlists for web fuzzing: curated `micro`, categorized `short`/`long`, and combined final lists.

## What this repo generates

- `onelistforallmicro.txt`: curated list (maintained manually).
- `dict/<category>_short.txt`: per-category curated wordlist (small/quality sources).
- `dict/<category>_long.txt`: per-category comprehensive wordlist (all sources).
- `dist/onelistforall.txt`: micro + all `*_short.txt`, deduplicated.
- `dist/onelistforall_big.txt`: everything combined, deduplicated.

## How it works

1. **Sync** (`update`): Clones ~36 wordlist repos into `sources/`.
2. **Classify** (`classify`): Walks every `.txt` file in `sources/`, classifies each into categories using path structure, filename keywords, and content sampling. Produces `classification_index.json`.
3. **Build** (`build --all-categories`): For each category, merges classified source files into `dict/{cat}_short.txt` and `dict/{cat}_long.txt` with filtering and deduplication.
4. **Assemble** (`assemble`): Combines micro + all shorts into `onelistforall.txt`, and everything into `onelistforall_big.txt`.

## Requirements

- Go 1.22+
- `git` (for `update`)
- `7z` (only for `package`)

## CLI (`olfa`)

```bash
# Check dependencies
go run ./cmd/olfa check

# List sources and categories
go run ./cmd/olfa list
go run ./cmd/olfa list --categories

# Sync source repos
go run ./cmd/olfa update
go run ./cmd/olfa update --source SecLists

# Classify source files
go run ./cmd/olfa classify
go run ./cmd/olfa classify --format json

# Build all category wordlists
go run ./cmd/olfa build --all-categories
go run ./cmd/olfa build --category wordpress --variant short

# Assemble final combined lists
go run ./cmd/olfa assemble

# Validate and package
go run ./cmd/olfa validate-categories
go run ./cmd/olfa stats --output-dir dist
go run ./cmd/olfa package
```

## Recommended workflow

```bash
go run ./cmd/olfa check
go run ./cmd/olfa update
go run ./cmd/olfa classify
go run ./cmd/olfa build --all-categories
go run ./cmd/olfa assemble
go run ./cmd/olfa validate-categories
go run ./cmd/olfa package
```

## Configuration

- `configs/pipeline.yml`: sources, filters, classification rules, dedup settings, release config.
- `configs/taxonomy.json`: category taxonomy (236 categories with aliases, including attack vectors, CMS, frameworks, cloud, and more).

## Category taxonomy

Categories are defined in `configs/taxonomy.json`. Each has a canonical name and aliases. Source files are matched to categories by:

1. Explicit path rules (e.g., `Discovery/DNS/*` → subdomains)
2. Directory path keywords matched against taxonomy
3. Filename keywords matched against taxonomy
4. Content sampling with regex patterns (fallback)
5. Source-level tags (last resort)

Short/long split is determined by line count threshold (default 5000) and filename keywords.
