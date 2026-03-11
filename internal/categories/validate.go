package categories

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/six2dez/OneListForAll/internal/config"
)

type ValidationResult struct {
	Category       string
	ShortExists    bool
	LongExists     bool
	ShortLines     int64
	LongLines      int64
	Passed         bool
	FailureReasons []string
}

// ValidateFromTaxonomy checks that dict/{cat}_short.txt and dict/{cat}_long.txt
// exist for each category in the taxonomy.
func ValidateFromTaxonomy(taxonomy *config.Taxonomy, categoriesDir string) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0, len(taxonomy.Categories))
	for _, cat := range taxonomy.Categories {
		shortFile := filepath.Join(categoriesDir, cat.Name+"_short.txt")
		longFile := filepath.Join(categoriesDir, cat.Name+"_long.txt")

		r := ValidationResult{
			Category: cat.Name,
			Passed:   true,
		}

		if _, err := os.Stat(shortFile); err == nil {
			r.ShortExists = true
			n, err := countLines(shortFile)
			if err != nil {
				return nil, err
			}
			r.ShortLines = n
		}

		if _, err := os.Stat(longFile); err == nil {
			r.LongExists = true
			n, err := countLines(longFile)
			if err != nil {
				return nil, err
			}
			r.LongLines = n
		}

		if !r.ShortExists && !r.LongExists {
			r.Passed = false
			r.FailureReasons = append(r.FailureReasons, "no output files found")
		}

		results = append(results, r)
	}
	return results, nil
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
