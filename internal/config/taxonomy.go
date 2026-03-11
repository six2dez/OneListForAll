package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Taxonomy struct {
	Categories []TaxCategory `json:"categories"`

	// lookup maps built on load: alias/name → canonical name
	lookup map[string]string
}

type TaxCategory struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

func LoadTaxonomy(path string) (*Taxonomy, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read taxonomy: %w", err)
	}
	return LoadTaxonomyFromBytes(buf)
}

// LoadTaxonomyFromBytes parses taxonomy JSON from raw bytes.
func LoadTaxonomyFromBytes(buf []byte) (*Taxonomy, error) {
	var t Taxonomy
	if err := json.Unmarshal(buf, &t); err != nil {
		return nil, fmt.Errorf("parse taxonomy: %w", err)
	}

	t.lookup = make(map[string]string, len(t.Categories)*4)
	for _, cat := range t.Categories {
		name := strings.ToLower(strings.TrimSpace(cat.Name))
		if name == "" {
			return nil, fmt.Errorf("taxonomy: empty category name")
		}
		if prev, exists := t.lookup[name]; exists {
			return nil, fmt.Errorf("taxonomy: duplicate name %q (conflicts with %q)", name, prev)
		}
		t.lookup[name] = name
		for _, alias := range cat.Aliases {
			a := strings.ToLower(strings.TrimSpace(alias))
			if a == "" {
				continue
			}
			if prev, exists := t.lookup[a]; exists {
				return nil, fmt.Errorf("taxonomy: duplicate alias %q in category %q (already mapped to %q)", a, name, prev)
			}
			t.lookup[a] = name
		}
	}

	return &t, nil
}

// Lookup returns the canonical category name for a given name or alias.
// Returns ("", false) if not found.
func (t *Taxonomy) Lookup(name string) (string, bool) {
	canonical, ok := t.lookup[strings.ToLower(strings.TrimSpace(name))]
	return canonical, ok
}

// AllCanonicalNames returns all canonical category names.
func (t *Taxonomy) AllCanonicalNames() []string {
	names := make([]string, len(t.Categories))
	for i, cat := range t.Categories {
		names[i] = cat.Name
	}
	return names
}

// AllLookupKeys returns all names and aliases that can be looked up.
func (t *Taxonomy) AllLookupKeys() []string {
	keys := make([]string, 0, len(t.lookup))
	for k := range t.lookup {
		keys = append(keys, k)
	}
	return keys
}

// CategoryByName returns the TaxCategory for a canonical name.
func (t *Taxonomy) CategoryByName(name string) (TaxCategory, bool) {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, cat := range t.Categories {
		if strings.ToLower(cat.Name) == n {
			return cat, true
		}
	}
	return TaxCategory{}, false
}
