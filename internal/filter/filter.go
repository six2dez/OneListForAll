package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/six2dez/OneListForAll/internal/config"
)

type Engine struct {
	maxLen    int
	trim      bool
	lowercase bool
	dropEmpty bool
	deny      []*regexp.Regexp
	allow     []*regexp.Regexp
}

func New(cfg config.Filters) (*Engine, error) {
	eng := &Engine{
		maxLen:    cfg.MaxLineLen,
		trim:      cfg.Trim,
		lowercase: cfg.Lowercase,
		dropEmpty: cfg.DropEmpty,
	}
	for _, expr := range cfg.RegexDenylist {
		r, err := regexp.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("invalid deny regex %q: %w", expr, err)
		}
		eng.deny = append(eng.deny, r)
	}
	for _, expr := range cfg.RegexAllowlist {
		r, err := regexp.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("invalid allow regex %q: %w", expr, err)
		}
		eng.allow = append(eng.allow, r)
	}
	return eng, nil
}

func (e *Engine) Process(line string) (string, bool) {
	if e.trim {
		line = strings.TrimSpace(line)
	}
	if e.lowercase {
		line = strings.ToLower(line)
	}
	if e.dropEmpty && line == "" {
		return "", false
	}
	if e.maxLen > 0 && len(line) > e.maxLen {
		return "", false
	}
	for _, r := range e.deny {
		if r.MatchString(line) {
			return "", false
		}
	}
	// If allowlist is set, line must match at least one pattern
	if len(e.allow) > 0 {
		matched := false
		for _, r := range e.allow {
			if r.MatchString(line) {
				matched = true
				break
			}
		}
		if !matched {
			return "", false
		}
	}
	return line, true
}
