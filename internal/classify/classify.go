package classify

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/six2dez/OneListForAll/internal/config"
)

// Classifier assigns source files to taxonomy categories using multiple signals.
type Classifier struct {
	taxonomy    *config.Taxonomy
	cfg         config.Classification
	sources     []config.Source
	pathRules   []compiledPathRule
	contentREs  []compiledContentPattern
}

type compiledPathRule struct {
	pattern             string
	category            string
	categoryFromDirname bool
}

type compiledContentPattern struct {
	re       *regexp.Regexp
	category string
}

// FileClassification is the result of classifying a single source file.
type FileClassification struct {
	SourceName      string   `json:"source_name"`
	SourcePath      string   `json:"source_path"`
	AbsPath         string   `json:"abs_path"`
	Categories      []string `json:"categories"`
	LineCount       int64    `json:"line_count"`
	IsShortEligible bool     `json:"is_short_eligible"`
}

// NewClassifier creates a classifier from taxonomy and config.
func NewClassifier(taxonomy *config.Taxonomy, cfg config.Classification, sources []config.Source) (*Classifier, error) {
	c := &Classifier{
		taxonomy: taxonomy,
		cfg:      cfg,
		sources:  sources,
	}

	// Compile path rules
	for _, r := range cfg.PathRules {
		c.pathRules = append(c.pathRules, compiledPathRule{
			pattern:             r.Pattern,
			category:            r.Category,
			categoryFromDirname: r.CategoryFromDirname,
		})
	}

	// Compile content patterns
	for _, p := range cfg.ContentPatterns {
		re, err := regexp.Compile("(?i)" + p.Pattern)
		if err != nil {
			return nil, err
		}
		c.contentREs = append(c.contentREs, compiledContentPattern{re: re, category: p.Category})
	}

	return c, nil
}

// ClassifyFile determines which categories a source file belongs to.
func (c *Classifier) ClassifyFile(sourceName, sourceBasePath, relPath string) (FileClassification, error) {
	absPath := filepath.Join(sourceBasePath, relPath)

	lineCount, err := CountLines(absPath)
	if err != nil {
		return FileClassification{}, err
	}

	cats := make(map[string]struct{})

	// Signal 1: Explicit path rules
	c.applyPathRules(relPath, cats)

	// Signal 2: Directory structure keywords
	c.applyPathKeywords(relPath, cats)

	// Signal 3: Filename keywords
	c.applyFilenameKeywords(relPath, cats)

	// Signal 4: Content sampling (only if no categories found yet)
	if len(cats) == 0 && c.cfg.ContentSampleLines > 0 {
		lines, err := SampleLines(absPath, c.cfg.ContentSampleLines)
		if err == nil {
			c.applyContentPatterns(lines, cats)
		}
	}

	// Signal 5: Source-level tags (last resort)
	if len(cats) == 0 {
		c.applySourceTags(sourceName, cats)
	}

	// Determine short eligibility
	isShort := c.isShortEligible(relPath, lineCount)

	// Collect and sort categories
	catList := make([]string, 0, len(cats))
	for cat := range cats {
		catList = append(catList, cat)
	}
	sort.Strings(catList)

	return FileClassification{
		SourceName:      sourceName,
		SourcePath:      relPath,
		AbsPath:         absPath,
		Categories:      catList,
		LineCount:       lineCount,
		IsShortEligible: isShort,
	}, nil
}

// applyPathRules checks explicit path-to-category mappings from config.
func (c *Classifier) applyPathRules(relPath string, cats map[string]struct{}) {
	for _, rule := range c.pathRules {
		matched, _ := filepath.Match(rule.pattern, relPath)
		if !matched {
			// Also try matching just against the path with forward slashes normalized
			normalized := filepath.ToSlash(relPath)
			matched, _ = filepath.Match(rule.pattern, normalized)
		}
		if !matched {
			continue
		}
		if rule.categoryFromDirname {
			// Extract category from the parent directory name
			dir := filepath.Dir(relPath)
			dirName := strings.ToLower(filepath.Base(dir))
			if canonical, ok := c.taxonomy.Lookup(dirName); ok {
				cats[canonical] = struct{}{}
			}
		} else if rule.category != "" {
			if canonical, ok := c.taxonomy.Lookup(rule.category); ok {
				cats[canonical] = struct{}{}
			}
		}
	}
}

// applyPathKeywords extracts keywords from directory path segments.
func (c *Classifier) applyPathKeywords(relPath string, cats map[string]struct{}) {
	// Split path into segments
	segments := strings.FieldsFunc(filepath.ToSlash(relPath), func(r rune) bool {
		return r == '/'
	})

	for _, seg := range segments {
		// Normalize segment: lowercase, split on separators
		keywords := splitKeywords(seg)
		for _, kw := range keywords {
			if canonical, ok := c.taxonomy.Lookup(kw); ok {
				cats[canonical] = struct{}{}
			}
		}
	}
}

// applyFilenameKeywords matches the filename against taxonomy names/aliases.
func (c *Classifier) applyFilenameKeywords(relPath string, cats map[string]struct{}) {
	base := filepath.Base(relPath)
	base = strings.TrimSuffix(strings.ToLower(base), ".txt")

	// Try exact match first
	if canonical, ok := c.taxonomy.Lookup(base); ok {
		cats[canonical] = struct{}{}
		return
	}

	// Strip common suffixes and try again
	for _, suffix := range []string{"_short", "_long", "_small", "_big", "_common", "_full"} {
		stripped := strings.TrimSuffix(base, suffix)
		if stripped != base {
			if canonical, ok := c.taxonomy.Lookup(stripped); ok {
				cats[canonical] = struct{}{}
			}
		}
	}

	// Split on separators and try individual parts
	keywords := splitKeywords(base)
	for _, kw := range keywords {
		if len(kw) < 2 {
			continue
		}
		if canonical, ok := c.taxonomy.Lookup(kw); ok {
			cats[canonical] = struct{}{}
		}
	}
}

// applyContentPatterns samples file content and matches regex patterns.
func (c *Classifier) applyContentPatterns(lines []string, cats map[string]struct{}) {
	// Count pattern matches per category
	counts := make(map[string]int)
	for _, line := range lines {
		for _, cp := range c.contentREs {
			if cp.re.MatchString(line) {
				if canonical, ok := c.taxonomy.Lookup(cp.category); ok {
					counts[canonical]++
				}
			}
		}
	}

	// A category needs at least 10% of sampled lines matching to be assigned
	threshold := len(lines) / 10
	if threshold < 3 {
		threshold = 3
	}
	for cat, count := range counts {
		if count >= threshold {
			cats[cat] = struct{}{}
		}
	}
}

// applySourceTags uses source-level tags as a weak classification signal.
func (c *Classifier) applySourceTags(sourceName string, cats map[string]struct{}) {
	for _, src := range c.sources {
		if src.Name != sourceName {
			continue
		}
		for _, tag := range src.Tags {
			if canonical, ok := c.taxonomy.Lookup(tag); ok {
				cats[canonical] = struct{}{}
			}
		}
		break
	}
}

// isShortEligible determines if a file qualifies for the short variant.
func (c *Classifier) isShortEligible(relPath string, lineCount int64) bool {
	lower := strings.ToLower(relPath)

	// Anti-keywords: never short
	for _, kw := range c.cfg.LongKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return false
		}
	}

	// Short keywords: always short regardless of size
	for _, kw := range c.cfg.ShortKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}

	// Fall back to line count threshold
	threshold := int64(c.cfg.ShortLineThreshold)
	if threshold <= 0 {
		threshold = 5000
	}
	return lineCount <= threshold
}

// IsVersionPath checks if a path matches a version directory pattern.
func IsVersionPath(relPath, versionPattern string) bool {
	if versionPattern == "" {
		return false
	}
	matched, _ := filepath.Match(versionPattern, relPath)
	if matched {
		return true
	}
	// Check if any parent directory matches
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	for i := range parts {
		partial := strings.Join(parts[:i+1], "/")
		if m, _ := filepath.Match(versionPattern, partial); m {
			return true
		}
	}
	return false
}

// splitKeywords splits a string on common separators.
func splitKeywords(s string) []string {
	s = strings.ToLower(s)
	s = strings.TrimSuffix(s, ".txt")
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == ' '
	})
}
