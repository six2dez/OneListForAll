# OneListForAll

Wordlists for web fuzzing: curated `micro`, categorized `short`/`long`, and combined final lists.

## What this repo generates

- `onelistforallmicro.txt`: curated list (maintained manually).
- `dict/<category>_short.txt`: per-category curated wordlist (small/quality sources).
- `dict/<category>_long.txt`: per-category comprehensive wordlist (all sources).
- `onelistforall.txt`: micro + all `*_short.txt`, deduplicated.
- `onelistforall_big.txt`: everything combined, deduplicated.

## How it works

1. **Sync** (`update`): Clones ~36 wordlist repos into `sources/`.
2. **Classify** (`classify`): Walks every `.txt` file in `sources/`, classifies each into categories using path structure, filename keywords, and content sampling. Produces `classification_index.json`.
3. **Build** (`build --all-categories`): For each category, merges classified source files into `dict/{cat}_short.txt` and `dict/{cat}_long.txt` with filtering and deduplication.
4. **Assemble** (`assemble`): Combines micro + all shorts into `onelistforall.txt`, and everything into `onelistforall_big.txt`.
5. **Package** (`package`): Creates 7z archives with checksums.
6. **Publish** (`publish`): Auto-commits `dict/` changes and pushes to remote.

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

# Full pipeline (all steps in one command)
go run ./cmd/olfa pipeline
go run ./cmd/olfa pipeline --dry-run
go run ./cmd/olfa pipeline --skip-publish       # skip git commit+push
go run ./cmd/olfa pipeline --skip-update        # skip git fetch
go run ./cmd/olfa pipeline --commit-msg "my custom message"
```

## Recommended workflow

```bash
# Option A: single command
go run ./cmd/olfa pipeline

# Option B: step by step
go run ./cmd/olfa check
go run ./cmd/olfa update
go run ./cmd/olfa classify
go run ./cmd/olfa build --all-categories
go run ./cmd/olfa assemble
go run ./cmd/olfa validate-categories
go run ./cmd/olfa package
git add dict/ && git commit -m "chore: update dict/ wordlists" && git push
```

## Disk space and large files

The repo includes ~418 category wordlists in `dict/` (~930 MB). However, 6 files exceed GitHub's 100 MB file size limit and are **not included** in the repository:

| File | Size |
|------|------|
| `dict/subdomains_long.txt` | 493 MB |
| `dict/passwords_long.txt` | 351 MB |
| `dict/passwords_short.txt` | 296 MB |
| `dict/fuzz_general_long.txt` | 178 MB |
| `dict/directories_long.txt` | 153 MB |
| `dict/dns_long.txt` | 112 MB |

To generate them locally, run:

```bash
go run ./cmd/olfa pipeline
```

Running the full pipeline (syncing sources + building all categories) requires **~15 GB** of disk space.

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
