package update

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/six2dez/OneListForAll/internal/config"
)

type Options struct {
	ConfigPath   string
	SourceName   string
	DryRun       bool
}

type SourceResult struct {
	Name            string
	Commit          string
	DownloadedFiles int
}

func Run(opts Options) ([]SourceResult, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, err
	}

	selected := make([]config.Source, 0, len(cfg.Sources))
	for _, s := range cfg.Sources {
		if opts.SourceName == "" || s.Name == opts.SourceName {
			selected = append(selected, s)
		}
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no sources matched %q", opts.SourceName)
	}

	if err := os.MkdirAll("sources", 0o755); err != nil {
		return nil, fmt.Errorf("create sources dir: %w", err)
	}

	results := make([]SourceResult, 0, len(selected))
	for _, src := range selected {
		r, err := syncSource(src, opts.DryRun)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	if !opts.DryRun {
		if err := writeLock(results, cfg); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func syncSource(src config.Source, dryRun bool) (SourceResult, error) {
	repoDir := filepath.Join("sources", src.Name)
	repoURL := "https://github.com/" + src.Repo + ".git"

	if _, err := os.Stat(repoDir); errors.Is(err, os.ErrNotExist) {
		if dryRun {
			return SourceResult{Name: src.Name, Commit: "dry-run", DownloadedFiles: 0}, nil
		}
		if err := run("git", "clone", "--depth=1", "--branch", src.Branch, repoURL, repoDir); err != nil {
			return SourceResult{}, fmt.Errorf("clone %s: %w", src.Name, err)
		}
	} else {
		if dryRun {
			return SourceResult{Name: src.Name, Commit: "dry-run", DownloadedFiles: 0}, nil
		}
		if err := run("git", "-C", repoDir, "fetch", "--depth=1", "origin", src.Branch); err != nil {
			return SourceResult{}, fmt.Errorf("fetch %s: %w", src.Name, err)
		}
		if err := run("git", "-C", repoDir, "checkout", "-B", src.Branch, "FETCH_HEAD"); err != nil {
			return SourceResult{}, fmt.Errorf("checkout %s: %w", src.Name, err)
		}
	}

	commitBytes, err := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD").Output()
	if err != nil {
		return SourceResult{}, fmt.Errorf("rev-parse %s: %w", src.Name, err)
	}
	commit := strings.TrimSpace(string(commitBytes))

	// Count .txt files without copying them
	count := 0
	for _, p := range src.Paths {
		root := repoDir
		if p != "all" {
			root = filepath.Join(repoDir, p)
		}
		if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			if strings.HasSuffix(strings.ToLower(d.Name()), ".txt") {
				count++
			}
			return nil
		})
	}

	return SourceResult{Name: src.Name, Commit: commit, DownloadedFiles: count}, nil
}

func run(name string, args ...string) error {
	attempts := 1
	if shouldRetryCommand(name, args) {
		attempts = 3
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		var stderr bytes.Buffer
		cmd := exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

		if err := cmd.Run(); err == nil {
			return nil
		} else {
			lastErr = err
			if attempt == attempts || !isRetryableGitError(stderr.String()) {
				return err
			}
			wait := time.Duration(attempt*2) * time.Second
			fmt.Fprintf(os.Stderr, "warning: transient git error, retrying in %s (%d/%d)\n", wait, attempt, attempts)
			time.Sleep(wait)
		}
	}

	return lastErr
}

func shouldRetryCommand(name string, args []string) bool {
	if name != "git" {
		return false
	}
	for _, arg := range args {
		switch arg {
		case "clone", "fetch", "pull", "ls-remote":
			return true
		}
	}
	return false
}

func isRetryableGitError(stderr string) bool {
	s := strings.ToLower(stderr)
	patterns := []string{
		"ssl_error_syscall",
		"could not resolve host",
		"connection timed out",
		"connection reset by peer",
		"gnutls recv error",
		"rpc failed",
		"http 502",
		"http 503",
		"http 504",
		"the requested url returned error: 5",
		"unexpected disconnect while reading sideband packet",
	}
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

func writeLock(results []SourceResult, cfg config.Config) error {
	f, err := os.Create("sources_lock.yml")
	if err != nil {
		return fmt.Errorf("create lock file: %w", err)
	}
	defer f.Close()

	_, _ = fmt.Fprintf(f, "lock_date: %q\n", time.Now().UTC().Format(time.RFC3339))
	_, _ = fmt.Fprintln(f, "sources:")
	for _, r := range results {
		_, _ = fmt.Fprintf(f, "  %s:\n", r.Name)
		_, _ = fmt.Fprintf(f, "    repo: %q\n", sourceRepo(cfg, r.Name))
		_, _ = fmt.Fprintf(f, "    commit: %q\n", r.Commit)
		_, _ = fmt.Fprintf(f, "    downloaded_files: %d\n", r.DownloadedFiles)
		_, _ = fmt.Fprintf(f, "    last_updated: %q\n", time.Now().UTC().Format(time.RFC3339))
	}
	return nil
}

func sourceRepo(cfg config.Config, name string) string {
	for _, s := range cfg.Sources {
		if s.Name == name {
			return s.Repo
		}
	}
	return ""
}
