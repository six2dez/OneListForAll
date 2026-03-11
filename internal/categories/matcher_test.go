package categories

import (
	"testing"

	"github.com/six2dez/OneListForAll/internal/config"
)

func TestMatchCategoryByName(t *testing.T) {
	taxonomy := buildTestTaxonomy()
	if !MatchCategory("dict/wordpress_long.txt", taxonomy, "wordpress") {
		t.Fatalf("expected wordpress file to match wordpress category")
	}
	if MatchCategory("dict/nginx_short.txt", taxonomy, "wordpress") {
		t.Fatalf("nginx file should not match wordpress category")
	}
}

func TestMatchCategoryByAlias(t *testing.T) {
	taxonomy := buildTestTaxonomy()
	if !MatchCategory("dict/wp-content_short.txt", taxonomy, "wordpress") {
		t.Fatalf("expected wp-content alias to match wordpress category")
	}
}

func buildTestTaxonomy() *config.Taxonomy {
	t, _ := config.LoadTaxonomyFromBytes([]byte(`{
		"categories": [
			{"name": "wordpress", "aliases": ["wp-content", "wp-plugins"]},
			{"name": "nginx", "aliases": []}
		]
	}`))
	return t
}
