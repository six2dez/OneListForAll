package classify

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ClassificationIndex is the full index of classified source files.
type ClassificationIndex struct {
	GeneratedAt string               `json:"generated_at"`
	Stats       IndexStats           `json:"stats"`
	Entries     []FileClassification `json:"entries"`
}

// IndexStats contains summary statistics about the classification.
type IndexStats struct {
	TotalFiles         int            `json:"total_files"`
	CategorizedFiles   int            `json:"categorized_files"`
	UncategorizedFiles int            `json:"uncategorized_files"`
	CategoryCounts     map[string]int `json:"category_counts"`
}

// WriteIndex writes the classification index to a JSON file.
func WriteIndex(path string, entries []FileClassification) error {
	stats := computeStats(entries)
	idx := ClassificationIndex{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Stats:       stats,
		Entries:     entries,
	}
	buf, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal classification index: %w", err)
	}
	return os.WriteFile(path, append(buf, '\n'), 0o644)
}

// ReadIndex reads the classification index from a JSON file.
func ReadIndex(path string) (ClassificationIndex, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return ClassificationIndex{}, fmt.Errorf("read classification index: %w", err)
	}
	var idx ClassificationIndex
	if err := json.Unmarshal(buf, &idx); err != nil {
		return ClassificationIndex{}, fmt.Errorf("parse classification index: %w", err)
	}
	return idx, nil
}

func computeStats(entries []FileClassification) IndexStats {
	stats := IndexStats{
		TotalFiles:     len(entries),
		CategoryCounts: make(map[string]int),
	}
	for _, e := range entries {
		if len(e.Categories) == 0 {
			stats.UncategorizedFiles++
		} else {
			stats.CategorizedFiles++
		}
		for _, cat := range e.Categories {
			stats.CategoryCounts[cat]++
		}
	}
	return stats
}
