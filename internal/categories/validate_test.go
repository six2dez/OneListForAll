package categories

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/six2dez/OneListForAll/internal/config"
)

func TestValidateFromTaxonomy(t *testing.T) {
	dict := t.TempDir()
	if err := os.WriteFile(filepath.Join(dict, "wordpress_short.txt"), []byte("a\nb\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dict, "wordpress_long.txt"), []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	taxonomy, _ := config.LoadTaxonomyFromBytes([]byte(`{
		"categories": [
			{"name": "wordpress", "aliases": ["wp-content"]},
			{"name": "nginx", "aliases": []}
		]
	}`))

	res, err := ValidateFromTaxonomy(taxonomy, dict)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
	if !res[0].Passed {
		t.Fatalf("expected wordpress to pass, got %+v", res[0])
	}
	if res[0].ShortLines != 2 {
		t.Fatalf("expected 2 short lines, got %d", res[0].ShortLines)
	}
	if res[1].Passed {
		t.Fatalf("expected nginx to fail (no files), got %+v", res[1])
	}
}
