package search

import (
	"strings"

	"github.com/grafana/regexp"
)

type pathMatcher struct {
	Include []*regexp.Regexp
	Exclude *regexp.Regexp
}

func (pm *pathMatcher) MatchPath(path string) bool {
	for _, re := range pm.Include {
		if !re.MatchString(path) {
			return false
		}
	}
	return pm.Exclude == nil || !pm.Exclude.MatchString(path)
}

func (pm *pathMatcher) String() string {
	var parts []string
	for _, re := range pm.Include {
		parts = append(parts, re.String())
	}
	if pm.Exclude != nil {
		parts = append(parts, "!"+pm.Exclude.String())
	}
	return strings.Join(parts, " ")
}

// compilePathPatterns returns a pathMatcher that matches a path iff:
//
// * all of the includePatterns match the path; AND
// * the excludePattern does NOT match the path.
func compilePathPatterns(includePatterns []string, excludePattern string, caseSensitive bool) (*pathMatcher, error) {
	// set err once if non-nil. This simplifies our many calls to compile.
	var err error
	compile := func(p string) *regexp.Regexp {
		if !caseSensitive {
			// Respect the CaseSensitive option. However, if the pattern already contains
			// (?i:...), then don't clear that 'i' flag (because we assume that behavior
			// is desirable in more cases).
			p = "(?i:" + p + ")"
		}
		re, innerErr := regexp.Compile(p)
		if innerErr != nil {
			err = innerErr
		}
		return re
	}

	var include []*regexp.Regexp
	for _, p := range includePatterns {
		include = append(include, compile(p))
	}

	var exclude *regexp.Regexp
	if excludePattern != "" {
		exclude = compile(excludePattern)
	}

	return &pathMatcher{
		Include: include,
		Exclude: exclude,
	}, err
}
