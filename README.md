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
go run ./cmd/olfa stats
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

Everything is controlled by two files:

- **`configs/pipeline.yml`** — sources, filters, classification rules, dedup settings, release config.
- **`configs/taxonomy.json`** — category taxonomy (236 categories with aliases).

### Adding a new source repo

Edit `configs/pipeline.yml` and add an entry to the `sources` array:

```json
{
  "name": "my-wordlists",
  "repo": "username/repo-name",
  "branch": "main",
  "paths": ["all"],
  "tags": ["directories", "api"],
  "priority": "medium"
}
```

| Field | Description |
|-------|-------------|
| `name` | Unique identifier for the source |
| `repo` | GitHub `owner/repo` (cloned via `https://github.com/...`) |
| `branch` | Branch to track |
| `paths` | Directories to scan inside the repo (`["all"]` = everything) |
| `tags` | Fallback categories when auto-classification can't determine the category |
| `priority` | `high`, `medium`, or `low` — high-priority sources go into `*_short.txt`, all sources go into `*_long.txt` |

After adding a source, run the pipeline to pull and classify it:

```bash
go run ./cmd/olfa pipeline
```

Or sync just the new source:

```bash
go run ./cmd/olfa update --source my-wordlists
```

### Filters

Global filters in `pipeline.yml` control what lines are kept or dropped:

- `regex_denylist` — lines matching any pattern are removed (e.g., URLs, UUIDs, image extensions)
- `max_line_len` — lines longer than this are dropped (default: 100 chars)
- `trim` — strip leading/trailing whitespace
- `drop_empty` — remove blank lines

Per-category filters can be set in `category_filters` (e.g., subdomains only allow valid hostname characters).

### Short vs Long split

Each category produces two wordlists:

- **`*_short.txt`** — only from `high` priority sources, or files with fewer than 5000 lines, or filenames containing keywords like `common`, `short`, `top`, `default`.
- **`*_long.txt`** — all sources combined.

Thresholds and keywords are configurable in `classification.short_line_threshold`, `classification.short_keywords`, and `classification.long_keywords`.

## Category taxonomy

Categories are defined in `configs/taxonomy.json`. Each has a canonical name and aliases. Source files are matched to categories by:

1. Explicit path rules (e.g., `Discovery/DNS/*` → subdomains)
2. Directory path keywords matched against taxonomy
3. Filename keywords matched against taxonomy
4. Content sampling with regex patterns (fallback)
5. Source-level tags (last resort)

To list all available categories:

```bash
go run ./cmd/olfa list --categories
go run ./cmd/olfa list --categories --format json
```
