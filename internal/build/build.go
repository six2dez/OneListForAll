package build

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/six2dez/OneListForAll/internal/classify"
	"github.com/six2dez/OneListForAll/internal/config"
	"github.com/six2dez/OneListForAll/internal/dedupe"
	"github.com/six2dez/OneListForAll/internal/filter"
)

type Options struct {
	ConfigPath    string
	TaxonomyPath  string
	IndexPath     string
	Category      string
	Variant       string
	AllCategories bool
	OutputDir     string
	DryRun        bool
	Clean         bool
}

type Result struct {
	Mode         string
	Category     string
	Variant      string
	OutputFile   string
	InputFiles   int
	LinesRead    int64
	LinesWritten int64
}

func Run(opts Options) ([]Result, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, err
	}

	taxonomy, err := config.LoadTaxonomy(opts.TaxonomyPath)
	if err != nil {
		return nil, err
	}

	idx, err := classify.ReadIndex(opts.IndexPath)
	if err != nil {
		return nil, fmt.Errorf("read classification index: %w (run 'olfa classify' first)", err)
	}

	// Clean existing outputs before building
	if opts.Clean && opts.AllCategories {
		if err := cleanOutputDir(opts.OutputDir); err != nil {
			return nil, fmt.Errorf("clean output dir: %w", err)
		}
	}

	switch {
	case opts.AllCategories:
		return runAllCategories(cfg, taxonomy, idx, opts)
	case opts.Category != "":
		return runCategory(cfg, taxonomy, idx, opts)
	default:
		return nil, fmt.Errorf("specify --category or --all-categories")
	}
}

func runCategory(cfg config.Config, taxonomy *config.Taxonomy, idx classify.ClassificationIndex, opts Options) ([]Result, error) {
	canonical, ok := taxonomy.Lookup(opts.Category)
	if !ok {
		return nil, fmt.Errorf("category %q not found in taxonomy", opts.Category)
	}
	if opts.Variant == "" {
		return nil, fmt.Errorf("variant is required with --category (short|long)")
	}
	if opts.Variant != "short" && opts.Variant != "long" {
		return nil, fmt.Errorf("invalid variant %q", opts.Variant)
	}

	res, err := buildCategoryVariant(cfg, idx, canonical, opts.Variant, opts.OutputDir, opts.DryRun)
	if err != nil {
		return nil, err
	}
	return []Result{res}, nil
}

func runAllCategories(cfg config.Config, taxonomy *config.Taxonomy, idx classify.ClassificationIndex, opts Options) ([]Result, error) {
	variants := []string{"short", "long"}
	if opts.Variant != "" {
		if opts.Variant != "short" && opts.Variant != "long" {
			return nil, fmt.Errorf("invalid variant %q", opts.Variant)
		}
		variants = []string{opts.Variant}
	}

	totalSteps := len(taxonomy.Categories) * len(variants)
	step := 0

	results := make([]Result, 0, totalSteps)
	for _, cat := range taxonomy.Categories {
		for _, variant := range variants {
			step++
			files := sourceFilesForCategory(idx, cat.Name, variant)
			if len(files) == 0 {
				fmt.Fprintf(os.Stderr, "[%d/%d] %s_%s: skip (no files)\n", step, totalSteps, cat.Name, variant)
				continue
			}
			fmt.Fprintf(os.Stderr, "[%d/%d] %s_%s: %d files ...\n", step, totalSteps, cat.Name, variant, len(files))

			res, err := buildCategoryVariant(cfg, idx, cat.Name, variant, opts.OutputDir, opts.DryRun)
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(os.Stderr, "[%d/%d] %s_%s: done (%d lines written)\n", step, totalSteps, cat.Name, variant, res.LinesWritten)
			results = append(results, res)
		}
	}
	return results, nil
}

// sourceFilesForCategory collects source file paths from the classification index.
func sourceFilesForCategory(idx classify.ClassificationIndex, category string, variant string) []string {
	seen := make(map[string]struct{})
	files := make([]string, 0, 64)
	for _, entry := range idx.Entries {
		for _, cat := range entry.Categories {
			if cat != category {
				continue
			}
			// Long includes everything; short only includes short-eligible files
			if variant == "short" && !entry.IsShortEligible {
				continue
			}
			if _, ok := seen[entry.AbsPath]; ok {
				continue
			}
			seen[entry.AbsPath] = struct{}{}
			files = append(files, entry.AbsPath)
			break
		}
	}
	sort.Strings(files)
	return files
}

func buildCategoryVariant(cfg config.Config, idx classify.ClassificationIndex, category, variant, outputDir string, dryRun bool) (Result, error) {
	files := sourceFilesForCategory(idx, category, variant)

	outName := fmt.Sprintf("%s_%s.txt", strings.ToLower(category), variant)
	out := filepath.Join(outputDir, outName)

	// Use category-specific filters merged with global filters
	catFilters := cfg.FiltersForCategory(category)
	res, err := buildFromFilesWithFilters(cfg, catFilters, files, out, dryRun)
	if err != nil {
		return Result{}, err
	}
	res.Mode = "category"
	res.Category = category
	res.Variant = variant
	return res, nil
}

// cleanOutputDir removes all *_short.txt and *_long.txt files from the output directory.
func cleanOutputDir(dir string) error {
	patterns := []string{"*_short.txt", "*_long.txt"}
	removed := 0
	for _, pat := range patterns {
		matches, _ := filepath.Glob(filepath.Join(dir, pat))
		for _, m := range matches {
			if err := os.Remove(m); err != nil && !os.IsNotExist(err) {
				return err
			}
			removed++
		}
	}
	if removed > 0 {
		fmt.Fprintf(os.Stderr, "clean: removed %d files from %s\n", removed, dir)
	}
	return nil
}

func buildFromFilesWithFilters(cfg config.Config, filters config.Filters, files []string, output string, dryRun bool) (Result, error) {
	if dryRun {
		return Result{OutputFile: output, InputFiles: len(files)}, nil
	}

	if len(files) == 0 {
		return Result{OutputFile: output, InputFiles: 0}, nil
	}

	engine, err := filter.New(filters)
	if err != nil {
		return Result{}, err
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return Result{}, fmt.Errorf("create output dir: %w", err)
	}

	cw := dedupe.NewChunkWriter(cfg.Dedupe.ChunkLines, cfg.Dedupe.TempDir)
	var linesRead int64

	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			// Source file may have been removed; skip
			fmt.Fprintf(os.Stderr, "warning: skip %s: %v\n", path, err)
			continue
		}
		s := bufio.NewScanner(f)
		s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for s.Scan() {
			linesRead++
			if line, keep := engine.Process(s.Text()); keep {
				if err := cw.Add(line); err != nil {
					_ = f.Close()
					return Result{}, err
				}
			}
		}
		if err := s.Err(); err != nil {
			_ = f.Close()
			return Result{}, fmt.Errorf("scan input %q: %w", path, err)
		}
		_ = f.Close()
	}

	chunks, err := cw.Close()
	if err != nil {
		return Result{}, err
	}
	written, err := dedupe.MergeToOutput(chunks, output)
	if err != nil {
		return Result{}, err
	}

	// Remove empty output files (no lines survived filtering)
	if written == 0 {
		_ = os.Remove(output)
	}

	return Result{
		OutputFile:   output,
		InputFiles:   len(files),
		LinesRead:    linesRead,
		LinesWritten: written,
	}, nil
}
