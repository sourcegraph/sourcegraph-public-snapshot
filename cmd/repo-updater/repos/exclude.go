package repos

import (
	"regexp"
	"strings"
)

// excludeFunc takes a string and returns true if it should be excluded. In
// the case of repo sourcing it will take a repository name or ID as input.
type excludeFunc func(string) bool

// excludeBuilder builds an excludeFunc.
type excludeBuilder struct {
	exact    map[string]struct{}
	patterns []*regexp.Regexp

	err error
}

// Exact will case-insensitively exclude the string name.
func (e *excludeBuilder) Exact(name string) {
	if e.exact == nil {
		e.exact = map[string]struct{}{}
	}
	if name == "" {
		return
	}
	e.exact[strings.ToLower(name)] = struct{}{}
}

// Pattern will exclude strings matching the regex pattern.
func (e *excludeBuilder) Pattern(pattern string) {
	if pattern == "" {
		return
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		e.err = err
		return
	}
	e.patterns = append(e.patterns, re)
}

// Build will return an excludeFunc based on the previous calls to Exact and
// Pattern.
func (e *excludeBuilder) Build() (excludeFunc, error) {
	return func(name string) bool {
		if _, ok := e.exact[strings.ToLower(name)]; ok {
			return true
		}

		for _, re := range e.patterns {
			if re.MatchString(name) {
				return true
			}
		}

		return false
	}, e.err
}
