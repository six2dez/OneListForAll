package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Version         string             `json:"version"`
	Sources         []Source           `json:"sources"`
	Filters         Filters            `json:"filters"`
	CategoryFilters map[string]Filters `json:"category_filters"`
	Build           Build              `json:"build"`
	Dedupe          Dedupe             `json:"dedupe"`
	Release         Release            `json:"release"`
	TaxonomyFile    string             `json:"taxonomy_file"`
	Classification  Classification     `json:"classification"`
}

type Source struct {
	Name              string   `json:"name"`
	Repo              string   `json:"repo"`
	Branch            string   `json:"branch"`
	Paths             []string `json:"paths"`
	Tags              []string `json:"tags"`
	Priority          string   `json:"priority"`
	CollapseVersions  bool     `json:"collapse_versions"`
	VersionDirPattern string   `json:"version_dir_pattern"`
}

type Classification struct {
	ShortLineThreshold int              `json:"short_line_threshold"`
	ContentSampleLines int              `json:"content_sample_lines"`
	PathRules          []PathRule       `json:"path_rules"`
	ContentPatterns    []ContentPattern `json:"content_patterns"`
	ShortKeywords      []string         `json:"short_keywords"`
	LongKeywords       []string         `json:"long_keywords"`
}

type PathRule struct {
	Pattern            string `json:"pattern"`
	Category           string `json:"category"`
	CategoryFromDirname bool  `json:"category_from_dirname"`
}

type ContentPattern struct {
	Pattern  string `json:"pattern"`
	Category string `json:"category"`
}

type Filters struct {
	RegexDenylist  []string `json:"regex_denylist"`
	RegexAllowlist []string `json:"regex_allowlist"`
	MaxLineLen     int      `json:"max_line_len"`
	Trim           bool     `json:"trim"`
	Lowercase      bool     `json:"lowercase"`
	DropEmpty      bool     `json:"drop_empty"`
}

type Build struct {
	Profiles []Profile `json:"profiles"`
}

type Profile struct {
	Name         string   `json:"name"`
	IncludeGlobs []string `json:"include_globs"`
	ExcludeGlobs []string `json:"exclude_globs"`
	OutputFile   string   `json:"output_file"`
}

type Dedupe struct {
	Strategy   string `json:"strategy"`
	ChunkLines int    `json:"chunk_lines"`
	TempDir    string `json:"temp_dir"`
}

type Release struct {
	Split7z  Split7z  `json:"split_7z"`
	Outputs  []string `json:"outputs"`
	Checksum string   `json:"checksum_file"`
}

type Split7z struct {
	Enabled  bool `json:"enabled"`
	VolumeMB int  `json:"volume_mb"`
}

func Load(path string) (Config, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(buf, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config (JSON-in-YAML format): %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.Version == "" {
		return errors.New("config.version is required")
	}
	if len(c.Build.Profiles) == 0 {
		return errors.New("build.profiles cannot be empty")
	}
	if c.Dedupe.ChunkLines <= 0 {
		return errors.New("dedupe.chunk_lines must be > 0")
	}
	for _, p := range c.Build.Profiles {
		if p.Name == "" {
			return errors.New("build.profiles[].name is required")
		}
		if len(p.IncludeGlobs) == 0 {
			return fmt.Errorf("profile %q include_globs cannot be empty", p.Name)
		}
		if p.OutputFile == "" {
			return fmt.Errorf("profile %q output_file is required", p.Name)
		}
	}

	return nil
}

func (c Config) Profile(name string) (Profile, error) {
	for _, p := range c.Build.Profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return Profile{}, fmt.Errorf("profile %q not found", name)
}

// FiltersForCategory returns the global filters merged with any
// category-specific overrides. Category filters add extra deny/allow
// rules on top of the global ones.
func (c Config) FiltersForCategory(category string) Filters {
	merged := c.Filters
	catF, ok := c.CategoryFilters[category]
	if !ok {
		return merged
	}
	merged.RegexDenylist = append(merged.RegexDenylist, catF.RegexDenylist...)
	merged.RegexAllowlist = append(merged.RegexAllowlist, catF.RegexAllowlist...)
	if catF.MaxLineLen > 0 {
		merged.MaxLineLen = catF.MaxLineLen
	}
	return merged
}
