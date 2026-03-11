package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/six2dez/OneListForAll/internal/build"
	"github.com/six2dez/OneListForAll/internal/categories"
	"github.com/six2dez/OneListForAll/internal/check"
	"github.com/six2dez/OneListForAll/internal/classify"
	"github.com/six2dez/OneListForAll/internal/config"
	"github.com/six2dez/OneListForAll/internal/release"
	"github.com/six2dez/OneListForAll/internal/stats"
	"github.com/six2dez/OneListForAll/internal/update"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		runBuild(os.Args[2:])
	case "update":
		runUpdate(os.Args[2:])
	case "classify":
		runClassify(os.Args[2:])
	case "assemble":
		runAssemble(os.Args[2:])
	case "stats":
		runStats(os.Args[2:])
	case "check":
		runCheck(os.Args[2:])
	case "list":
		runList(os.Args[2:])
	case "package":
		runPackage(os.Args[2:])
	case "pipeline":
		runPipeline(os.Args[2:])
	case "validate-categories":
		runValidateCategories(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yml", "config file path")
	taxonomyPath := fs.String("taxonomy", "configs/taxonomy.json", "taxonomy file path")
	indexPath := fs.String("index", "classification_index.json", "classification index path")
	category := fs.String("category", "", "category name")
	variant := fs.String("variant", "", "category variant: short|long")
	allCategories := fs.Bool("all-categories", false, "build every category (short+long)")
	outputDir := fs.String("output-dir", "dict", "output directory for category builds")
	dryRun := fs.Bool("dry-run", false, "show what would run")
	clean := fs.Bool("clean", false, "remove existing outputs before building")
	_ = fs.Parse(args)

	if *allCategories && *category != "" {
		die(errors.New("use either --category or --all-categories, not both"))
	}

	if *category == "" && !*allCategories {
		die(errors.New("specify --category or --all-categories"))
	}

	results, err := build.Run(build.Options{
		ConfigPath:    *configPath,
		TaxonomyPath:  *taxonomyPath,
		IndexPath:     *indexPath,
		Category:      *category,
		Variant:       strings.ToLower(strings.TrimSpace(*variant)),
		AllCategories: *allCategories,
		OutputDir:     *outputDir,
		DryRun:        *dryRun,
		Clean:         *clean,
	})
	if err != nil {
		die(err)
	}

	for _, res := range results {
		fmt.Printf("mode=%s category=%s variant=%s input_files=%d output=%s lines_read=%d lines_written=%d\n",
			res.Mode, res.Category, res.Variant, res.InputFiles, res.OutputFile, res.LinesRead, res.LinesWritten)
	}
}

func runUpdate(args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yml", "config file path")
	source := fs.String("source", "", "source name")
	dryRun := fs.Bool("dry-run", false, "show what would run")
	_ = fs.Parse(args)

	res, err := update.Run(update.Options{ConfigPath: *configPath, SourceName: *source, DryRun: *dryRun})
	if err != nil {
		die(err)
	}
	for _, r := range res {
		fmt.Printf("source=%s commit=%s txt_files=%d\n", r.Name, r.Commit, r.DownloadedFiles)
	}
}

func runClassify(args []string) {
	fs := flag.NewFlagSet("classify", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yml", "config file path")
	taxonomyPath := fs.String("taxonomy", "configs/taxonomy.json", "taxonomy file path")
	sourcesDir := fs.String("sources-dir", "sources", "sources directory")
	output := fs.String("output", "classification_index.json", "output index path")
	dryRun := fs.Bool("dry-run", false, "show what would run")
	format := fs.String("format", "summary", "output format: summary|json")
	_ = fs.Parse(args)

	idx, err := classify.Run(classify.Options{
		ConfigPath:   *configPath,
		TaxonomyPath: *taxonomyPath,
		SourcesDir:   *sourcesDir,
		OutputIndex:  *output,
		DryRun:       *dryRun,
	})
	if err != nil {
		die(err)
	}

	if *format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(idx)
		return
	}

	fmt.Printf("Classification complete:\n")
	fmt.Printf("  total_files:         %d\n", idx.Stats.TotalFiles)
	fmt.Printf("  categorized_files:   %d\n", idx.Stats.CategorizedFiles)
	fmt.Printf("  uncategorized_files: %d\n", idx.Stats.UncategorizedFiles)
	fmt.Printf("  categories_found:    %d\n", len(idx.Stats.CategoryCounts))
	if len(idx.Stats.CategoryCounts) > 0 {
		fmt.Println("  top categories:")
		type kv struct {
			k string
			v int
		}
		sorted := make([]kv, 0, len(idx.Stats.CategoryCounts))
		for k, v := range idx.Stats.CategoryCounts {
			sorted = append(sorted, kv{k, v})
		}
		// Sort by count desc
		for i := range sorted {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].v > sorted[i].v {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		limit := 20
		if len(sorted) < limit {
			limit = len(sorted)
		}
		for _, kv := range sorted[:limit] {
			fmt.Printf("    %-30s %d files\n", kv.k, kv.v)
		}
	}
}

func runAssemble(args []string) {
	fs := flag.NewFlagSet("assemble", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yml", "config file path")
	categoriesDir := fs.String("categories-dir", "dict", "directory with category short/long files")
	microFile := fs.String("micro", "onelistforallmicro.txt", "path to micro wordlist")
	outputDir := fs.String("output-dir", "dist", "output directory for assembled lists")
	dryRun := fs.Bool("dry-run", false, "show what would run")
	cleanAssemble := fs.Bool("clean", false, "remove existing outputs before assembling")
	_ = fs.Parse(args)

	results, err := build.Assemble(build.AssembleOptions{
		ConfigPath:    *configPath,
		CategoriesDir: *categoriesDir,
		MicroFile:     *microFile,
		OutputDir:     *outputDir,
		DryRun:        *dryRun,
		Clean:         *cleanAssemble,
	})
	if err != nil {
		die(err)
	}

	for _, res := range results {
		fmt.Printf("output=%s input_files=%d lines_written=%d\n", res.OutputFile, res.InputFiles, res.LinesWritten)
	}
}

func runStats(args []string) {
	fs := flag.NewFlagSet("stats", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yml", "config file path")
	outputDir := fs.String("output-dir", "dist", "output directory")
	format := fs.String("format", "table", "table|json")
	_ = fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		die(err)
	}
	res, err := stats.Run(*outputDir, cfg.Release.Outputs)
	if err != nil {
		die(err)
	}

	if *format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return
	}

	fmt.Println("OneListForAll Statistics")
	for _, e := range res.Entries {
		fmt.Printf("- %s: lines=%d size=%d bytes\n", e.Name, e.Lines, e.Bytes)
	}
}

func runCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yml", "config file path")
	_ = fs.Parse(args)

	res, err := check.Run(*configPath)
	if err != nil {
		die(err)
	}
	fmt.Printf("config_ok=%t git_ok=%t sevenzip_ok=%t\n", res.ConfigOK, res.GitOK, res.SevenZip)
}

func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yml", "config file path")
	taxonomyPath := fs.String("taxonomy", "configs/taxonomy.json", "taxonomy file path")
	format := fs.String("format", "table", "table|json")
	listCategories := fs.Bool("categories", false, "list taxonomy categories")
	_ = fs.Parse(args)

	if *listCategories {
		taxonomy, err := config.LoadTaxonomy(*taxonomyPath)
		if err != nil {
			die(err)
		}
		if *format == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(taxonomy.Categories)
			return
		}
		for _, c := range taxonomy.Categories {
			fmt.Printf("- %s aliases=%d\n", c.Name, len(c.Aliases))
		}
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		die(err)
	}

	if *format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(cfg.Sources)
		return
	}
	for _, s := range cfg.Sources {
		fmt.Printf("- %s (%s) branch=%s priority=%s\n", s.Name, s.Repo, s.Branch, s.Priority)
	}
}

func runPackage(args []string) {
	fs := flag.NewFlagSet("package", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yml", "config file path")
	outputDir := fs.String("output-dir", "dist", "output directory")
	_ = fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		die(err)
	}
	if err := release.Package7z(*outputDir, cfg.Release); err != nil {
		die(err)
	}
	if err := release.WriteChecksums(*outputDir, cfg.Release.Outputs, cfg.Release.Checksum); err != nil {
		die(err)
	}
	fmt.Println("package complete")
}

func runValidateCategories(args []string) {
	fs := flag.NewFlagSet("validate-categories", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yml", "config file path")
	taxonomyPath := fs.String("taxonomy", "configs/taxonomy.json", "taxonomy file path")
	categoriesDir := fs.String("categories-dir", "dict", "directory with built category files")
	_ = fs.Parse(args)

	taxonomy, err := config.LoadTaxonomy(*taxonomyPath)
	if err != nil {
		die(err)
	}

	results, err := categories.ValidateFromTaxonomy(taxonomy, *categoriesDir)
	if err != nil {
		die(err)
	}

	_ = configPath // kept for future use

	var warned, passed int
	for _, r := range results {
		status := "PASS"
		if !r.Passed {
			status = "WARN"
			warned++
		} else {
			passed++
		}
		fmt.Printf("[%s] category=%s short_exists=%t long_exists=%t short_lines=%d long_lines=%d\n",
			status, r.Category, r.ShortExists, r.LongExists, r.ShortLines, r.LongLines)
	}
	fmt.Printf("\nSummary: %d passed, %d empty (no source files matched)\n", passed, warned)
}

func runPipeline(args []string) {
	fs := flag.NewFlagSet("pipeline", flag.ExitOnError)
	configPath := fs.String("config", "configs/pipeline.yml", "config file path")
	taxonomyPath := fs.String("taxonomy", "configs/taxonomy.json", "taxonomy file path")
	skipUpdate := fs.Bool("skip-update", false, "skip the update step")
	skipClassify := fs.Bool("skip-classify", false, "skip the classify step")
	skipBuild := fs.Bool("skip-build", false, "skip the build step")
	skipAssemble := fs.Bool("skip-assemble", false, "skip the assemble step")
	skipPackage := fs.Bool("skip-package", false, "skip the package step")
	skipPublish := fs.Bool("skip-publish", false, "skip the git commit+push step")
	commitMsg := fs.String("commit-msg", "chore: update dict/ wordlists", "commit message for publish step")
	dryRun := fs.Bool("dry-run", false, "show what would run")
	clean := fs.Bool("clean", true, "remove existing outputs before building")
	_ = fs.Parse(args)

	type step struct {
		name string
		skip *bool
		fn   func() error
	}

	sourcesDir := "sources"
	indexPath := "classification_index.json"
	outputDir := "dict"
	assembleDir := "dist"
	microFile := "onelistforallmicro.txt"

	steps := []step{
		{
			name: "update",
			skip: skipUpdate,
			fn: func() error {
				_, err := update.Run(update.Options{
					ConfigPath: *configPath,
					DryRun:     *dryRun,
				})
				return err
			},
		},
		{
			name: "classify",
			skip: skipClassify,
			fn: func() error {
				_, err := classify.Run(classify.Options{
					ConfigPath:   *configPath,
					TaxonomyPath: *taxonomyPath,
					SourcesDir:   sourcesDir,
					OutputIndex:  indexPath,
					DryRun:       *dryRun,
				})
				return err
			},
		},
		{
			name: "build",
			skip: skipBuild,
			fn: func() error {
				_, err := build.Run(build.Options{
					ConfigPath:    *configPath,
					TaxonomyPath:  *taxonomyPath,
					IndexPath:     indexPath,
					AllCategories: true,
					OutputDir:     outputDir,
					DryRun:        *dryRun,
					Clean:         *clean,
				})
				return err
			},
		},
		{
			name: "assemble",
			skip: skipAssemble,
			fn: func() error {
				_, err := build.Assemble(build.AssembleOptions{
					ConfigPath:    *configPath,
					CategoriesDir: outputDir,
					MicroFile:     microFile,
					OutputDir:     assembleDir,
					DryRun:        *dryRun,
					Clean:         *clean,
				})
				return err
			},
		},
		{
			name: "package",
			skip: skipPackage,
			fn: func() error {
				cfg, err := config.Load(*configPath)
				if err != nil {
					return err
				}
				if err := release.Package7z(assembleDir, cfg.Release); err != nil {
					return err
				}
				return release.WriteChecksums(assembleDir, cfg.Release.Outputs, cfg.Release.Checksum)
			},
		},
		{
			name: "publish",
			skip: skipPublish,
			fn: func() error {
				if *dryRun {
					fmt.Println("  [dry-run] git add dict/")
					fmt.Println("  [dry-run] git diff --cached --quiet")
					fmt.Println("  [dry-run] git commit -m " + *commitMsg)
					fmt.Println("  [dry-run] git push")
					return nil
				}
				if err := gitCmd("add", "dict/"); err != nil {
					return fmt.Errorf("git add: %w", err)
				}
				if err := gitCmd("diff", "--cached", "--quiet"); err == nil {
					fmt.Println("  nothing to commit")
					return nil
				}
				if err := gitCmd("commit", "-m", *commitMsg); err != nil {
					return fmt.Errorf("git commit: %w", err)
				}
				if err := gitCmd("push"); err != nil {
					return fmt.Errorf("git push: %w", err)
				}
				return nil
			},
		},
	}

	total := len(steps)
	for i, s := range steps {
		n := i + 1
		if *s.skip {
			fmt.Printf("[%d/%d] %s [skip]\n", n, total, s.name)
			continue
		}
		fmt.Printf("[%d/%d] %s ...\n", n, total, s.name)
		if err := s.fn(); err != nil {
			die(fmt.Errorf("step %s failed: %w", s.name, err))
		}
		fmt.Printf("[%d/%d] %s done\n", n, total, s.name)
	}

	fmt.Println("pipeline complete")
}

func usage() {
	fmt.Println(`olfa - OneListForAll wordlist builder

Commands:
  pipeline             Run full workflow (update→classify→build→assemble→package→publish)
  update               Sync source repositories (git clone/fetch)
  classify             Classify source files into categories
  build                Build category wordlists (short + long)
  assemble             Combine category outputs into final lists
  stats                Show output file statistics
  check                Validate dependencies and config
  list                 List configured sources or categories
  package              Create 7z archives with checksums
  validate-categories  Validate category output coverage

Workflow:
  olfa pipeline                     # run all steps
  olfa pipeline --skip-update       # skip git fetch
  olfa pipeline --skip-package      # skip 7z packaging
  olfa pipeline --skip-publish      # skip git commit+push

  Or run steps individually:
  1. olfa update
  2. olfa classify
  3. olfa build --all-categories
  4. olfa assemble
  5. olfa package
  6. git add dict/ && git commit && git push`)
}

func gitCmd(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

// Ensure filepath is used (for future use)
var _ = filepath.Join
