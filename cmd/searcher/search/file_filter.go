package search

import (
	"regexp"
)

type fileFilter struct {
	includes []*regexp.Regexp
	exclude  *regexp.Regexp
}

func (f *fileFilter) Match(filename string) bool {
	if f.exclude != nil && f.exclude.MatchString(filename) {
		return false
	}

	for _, include := range f.includes {
		if !include.MatchString(filename) {
			return false
		}
	}

	return true
}

func newFileFilter(includePatterns []string, excludePattern string) (*fileFilter, error) {
	includes := make([]*regexp.Regexp, 0, len(includePatterns))
	for _, pat := range includePatterns {
		r, err := regexp.Compile(pat)
		if err != nil {
			return nil, err
		}

		includes = append(includes, r)
	}

	var exclude *regexp.Regexp
	var err error
	if excludePattern != "" {
		exclude, err = regexp.Compile(excludePattern)
		if err != nil {
			return nil, err
		}
	}

	return &fileFilter{
		includes: includes,
		exclude:  exclude,
	}, nil
}
