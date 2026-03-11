package categories

import (
	"strings"

	"github.com/six2dez/OneListForAll/internal/config"
)

// MatchCategory checks if a file belongs to a taxonomy category by matching
// the file's base name against the category's canonical name and aliases.
func MatchCategory(file string, taxonomy *config.Taxonomy, categoryName string) bool {
	cat, ok := taxonomy.CategoryByName(categoryName)
	if !ok {
		return false
	}

	base := BaseName(file)

	// Check canonical name
	if base == strings.ToLower(cat.Name) {
		return true
	}

	// Check aliases
	for _, alias := range cat.Aliases {
		if base == strings.ToLower(alias) {
			return true
		}
	}

	return false
}
