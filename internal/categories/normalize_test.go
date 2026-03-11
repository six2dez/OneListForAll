package categories

import (
	"testing"

	"github.com/six2dez/OneListForAll/internal/config"
)

func TestNormalizeWithTaxonomy(t *testing.T) {
	taxonomy, _ := config.LoadTaxonomyFromBytes([]byte(`{
		"categories": [
			{"name": "api", "aliases": ["api_objetcs", "api_objects"]}
		]
	}`))
	got := Normalize("api_objetcs", taxonomy)
	if got != "api" {
		t.Fatalf("expected api, got %s", got)
	}
}

func TestNormalizeNilTaxonomy(t *testing.T) {
	got := Normalize("something", nil)
	if got != "something" {
		t.Fatalf("expected something, got %s", got)
	}
}

func TestBaseNameRemovesVariantSuffix(t *testing.T) {
	got := BaseName("dict/wordpress-random_short.txt")
	if got != "wordpress-random" {
		t.Fatalf("unexpected basename: %s", got)
	}
}
