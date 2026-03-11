package classify

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/six2dez/OneListForAll/internal/config"
)

// Options controls the classification run.
type Options struct {
	ConfigPath  string
	TaxonomyPath string
	SourcesDir  string
	OutputIndex string
	DryRun      bool
}

// Run walks all source directories, classifies each .txt file,
// and writes the classification index.
func Run(opts Options) (ClassificationIndex, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return ClassificationIndex{}, err
	}

	taxonomy, err := config.LoadTaxonomy(opts.TaxonomyPath)
	if err != nil {
		return ClassificationIndex{}, err
	}

	classifier, err := NewClassifier(taxonomy, cfg.Classification, cfg.Sources)
	if err != nil {
		return ClassificationIndex{}, err
	}

	entries := make([]FileClassification, 0, 4096)

	for _, src := range cfg.Sources {
		srcDir := filepath.Join(opts.SourcesDir, src.Name)
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			continue
		}

		for _, p := range src.Paths {
			root := srcDir
			if p != "all" {
				root = filepath.Join(srcDir, p)
			}
			if _, err := os.Stat(root); os.IsNotExist(err) {
				continue
			}

			err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if !strings.HasSuffix(strings.ToLower(d.Name()), ".txt") {
					return nil
				}

				relPath, err := filepath.Rel(srcDir, path)
				if err != nil {
					return err
				}

				// Handle version collapsing: mark version files as not short eligible
				if src.CollapseVersions && IsVersionPath(relPath, src.VersionDirPattern) {
					fc, err := classifier.ClassifyFile(src.Name, srcDir, relPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "warning: classify %s/%s: %v\n", src.Name, relPath, err)
						return nil
					}
					fc.IsShortEligible = false // version files are never short
					entries = append(entries, fc)
					return nil
				}

				fc, err := classifier.ClassifyFile(src.Name, srcDir, relPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: classify %s/%s: %v\n", src.Name, relPath, err)
					return nil
				}
				entries = append(entries, fc)
				return nil
			})
			if err != nil {
				return ClassificationIndex{}, fmt.Errorf("walk %s/%s: %w", src.Name, p, err)
			}
		}
	}

	idx := ClassificationIndex{
		Stats:   computeStats(entries),
		Entries: entries,
	}

	if !opts.DryRun && opts.OutputIndex != "" {
		if err := WriteIndex(opts.OutputIndex, entries); err != nil {
			return ClassificationIndex{}, err
		}
	}

	return idx, nil
}
