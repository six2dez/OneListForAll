package categories

import (
	"path/filepath"
	"strings"

	"github.com/six2dez/OneListForAll/internal/config"
)

// BaseName extracts the base name from a filename, stripping .txt and _short/_long suffixes.
func BaseName(file string) string {
	base := strings.ToLower(filepath.Base(file))
	base = strings.TrimSuffix(base, ".txt")
	base = strings.TrimSuffix(base, "_short")
	base = strings.TrimSuffix(base, "_long")
	return base
}

// Normalize maps a name to its canonical form using the taxonomy.
// Returns the lowercase name if no taxonomy match is found.
func Normalize(name string, taxonomy *config.Taxonomy) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return ""
	}
	if taxonomy == nil {
		return n
	}
	if canonical, ok := taxonomy.Lookup(n); ok {
		return canonical
	}
	return n
}
