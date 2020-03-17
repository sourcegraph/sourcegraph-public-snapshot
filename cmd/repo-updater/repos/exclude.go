package repos

import (
	"regexp"
	"strings"
)

type excluder struct {
	exact    map[string]struct{}
	patterns []*regexp.Regexp

	err error
}

func (e *excluder) Exact(name string) {
	if e.exact == nil {
		e.exact = map[string]struct{}{}
	}
	if name == "" {
		return
	}
	e.exact[strings.ToLower(name)] = struct{}{}
}

func (e *excluder) Pattern(pattern string) {
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

func (e *excluder) Err() error {
	return e.err
}

func (e *excluder) Match(name string) bool {
	if _, ok := e.exact[strings.ToLower(name)]; ok {
		return true
	}

	for _, re := range e.patterns {
		if re.MatchString(name) {
			return true
		}
	}

	return false
}
