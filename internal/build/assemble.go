package build

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/six2dez/OneListForAll/internal/config"
	"github.com/six2dez/OneListForAll/internal/dedupe"
	"github.com/six2dez/OneListForAll/internal/filter"
)

// AssembleOptions controls the final list assembly.
type AssembleOptions struct {
	ConfigPath    string
	CategoriesDir string // where {cat}_short.txt and {cat}_long.txt live
	MicroFile     string // path to onelistforall_micro.txt
	OutputDir     string
	DryRun        bool
	Clean         bool
}

// AssembleResult reports what was produced.
type AssembleResult struct {
	OutputFile   string
	InputFiles   int
	LinesWritten int64
}

// Assemble builds the final combined wordlists from category outputs.
//
// - onelistforall.txt = micro + all *_short.txt
// - onelistforall_big.txt = everything (*_short.txt + *_long.txt)
func Assemble(opts AssembleOptions) ([]AssembleResult, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, err
	}

	// Clean existing outputs before assembling
	if opts.Clean {
		for _, name := range []string{"onelistforall.txt", "onelistforall_big.txt"} {
			p := filepath.Join(opts.OutputDir, name)
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("clean %s: %w", p, err)
			}
		}
	}

	// Collect category files
	shortFiles, _ := filepath.Glob(filepath.Join(opts.CategoriesDir, "*_short.txt"))
	longFiles, _ := filepath.Glob(filepath.Join(opts.CategoriesDir, "*_long.txt"))
	sort.Strings(shortFiles)
	sort.Strings(longFiles)

	var results []AssembleResult

	// Build onelistforall.txt = micro + all shorts
	{
		inputFiles := make([]string, 0, len(shortFiles)+1)
		if opts.MicroFile != "" {
			if _, err := os.Stat(opts.MicroFile); err == nil {
				inputFiles = append(inputFiles, opts.MicroFile)
			}
		}
		inputFiles = append(inputFiles, shortFiles...)

		out := filepath.Join(opts.OutputDir, "onelistforall.txt")
		res, err := assembleFiles(cfg, inputFiles, out, opts.DryRun)
		if err != nil {
			return nil, fmt.Errorf("assemble onelistforall.txt: %w", err)
		}
		results = append(results, res)
	}

	// Build onelistforall_big.txt = everything
	{
		allFiles := make([]string, 0, len(shortFiles)+len(longFiles)+1)
		if opts.MicroFile != "" {
			if _, err := os.Stat(opts.MicroFile); err == nil {
				allFiles = append(allFiles, opts.MicroFile)
			}
		}
		allFiles = append(allFiles, shortFiles...)
		allFiles = append(allFiles, longFiles...)

		out := filepath.Join(opts.OutputDir, "onelistforall_big.txt")
		res, err := assembleFiles(cfg, allFiles, out, opts.DryRun)
		if err != nil {
			return nil, fmt.Errorf("assemble onelistforall_big.txt: %w", err)
		}
		results = append(results, res)
	}

	return results, nil
}

func assembleFiles(cfg config.Config, files []string, output string, dryRun bool) (AssembleResult, error) {
	if dryRun {
		return AssembleResult{OutputFile: output, InputFiles: len(files)}, nil
	}

	if len(files) == 0 {
		return AssembleResult{OutputFile: output}, nil
	}

	// For assembly, use relaxed filter (files are already filtered by category build)
	engine, err := filter.New(config.Filters{
		Trim:      true,
		DropEmpty: true,
	})
	if err != nil {
		return AssembleResult{}, err
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return AssembleResult{}, fmt.Errorf("create output dir: %w", err)
	}

	cw := dedupe.NewChunkWriter(cfg.Dedupe.ChunkLines, cfg.Dedupe.TempDir)
	var linesRead int64

	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
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
					return AssembleResult{}, err
				}
			}
		}
		if err := s.Err(); err != nil {
			_ = f.Close()
			return AssembleResult{}, fmt.Errorf("scan %q: %w", path, err)
		}
		_ = f.Close()
	}

	chunks, err := cw.Close()
	if err != nil {
		return AssembleResult{}, err
	}
	written, err := dedupe.MergeToOutput(chunks, output)
	if err != nil {
		return AssembleResult{}, err
	}

	return AssembleResult{
		OutputFile:   output,
		InputFiles:   len(files),
		LinesWritten: written,
	}, nil
}
