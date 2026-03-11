package stats

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type Entry struct {
	Name  string
	Lines int64
	Bytes int64
}

type Result struct {
	Entries []Entry
}

func Run(outputDir string, files []string) (Result, error) {
	entries := make([]Entry, 0, len(files))
	for _, f := range files {
		path := filepath.Join(outputDir, f)
		st, err := os.Stat(path)
		if err != nil {
			continue
		}
		lines, err := countLines(path)
		if err != nil {
			return Result{}, err
		}
		entries = append(entries, Entry{Name: f, Lines: lines, Bytes: st.Size()})
	}
	return Result{Entries: entries}, nil
}

func countLines(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var n int64
	for s.Scan() {
		n++
	}
	if err := s.Err(); err != nil {
		return 0, fmt.Errorf("scan %q: %w", path, err)
	}
	return n, nil
}
