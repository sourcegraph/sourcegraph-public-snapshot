// Package pathmatch provides helpers for matching paths against globs
// and regular expressions.
package pathmatch

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/grafana/regexp"
)

// PathMatcher reports whether the path was matched.
type PathMatcher interface {
	MatchPath(path string) bool

	// String returns the source text used to compile the PatchMatcher.
	String() string
}

type pathMatcherFunc struct {
	matcher func(path string) bool
	pattern string
}

func (f *pathMatcherFunc) MatchPath(path string) bool { return f.matcher(path) }

func (f *pathMatcherFunc) String() string {
	return f.pattern
}

// regexpMatcher is a PathMatcher backed by a regexp.
type regexpMatcher regexp.Regexp

func (m *regexpMatcher) MatchPath(path string) bool {
	return (*regexp.Regexp)(m).MatchString(path)
}

func (m *regexpMatcher) String() string {
	return fmt.Sprintf("re:%s", (*regexp.Regexp)(m).String())
}

// CompileOptions specifies options about the patterns to compile.
type CompileOptions struct {
	RegExp        bool // whether the patterns are regular expressions (false means globs)
	CaseSensitive bool // whether the patterns are case sensitive
}

// CompilePattern compiles pattern into a PathMatcher func.
func CompilePattern(pattern string, options CompileOptions) (PathMatcher, error) {
	if options.RegExp {
		// Respect the CaseSensitive option. However, if the pattern already contains
		// (?i:...), then don't clear that 'i' flag (because we assume that behavior
		// is desirable in more cases).
		if !options.CaseSensitive {
			pattern = "(?i:" + pattern + ")"
		}
		p, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		return (*regexpMatcher)(p), nil
	}

	if !options.CaseSensitive {
		pattern = strings.ToLower(pattern)
	}
	p, err := glob.Compile(pattern)
	if err != nil {
		return nil, err
	}
	if !options.CaseSensitive {
		// Use a match func that lowercases the input because globbing has no
		// first-class concept of case-insensitivity (as regexps do, with the 'i' flag).
		return &pathMatcherFunc{
			matcher: func(path string) bool {
				return p.Match(strings.ToLower(path))
			},
			pattern: "iglob:" + pattern,
		}, nil
	}
	return &pathMatcherFunc{
		matcher: p.Match,
		pattern: "glob:" + pattern,
	}, nil
}

// pathMatcherAnd is a PathMatcher that matches a path iff all of the
// underlying matchers match the path.
type pathMatcherAnd []PathMatcher

func (pm pathMatcherAnd) MatchPath(path string) bool {
	for _, m := range pm {
		if !m.MatchPath(path) {
			return false
		}
	}
	return true
}

func (pm pathMatcherAnd) String() string {
	var b bytes.Buffer
	b.WriteString("li:")
	for i, m := range pm {
		b.WriteString(m.String())
		if i != len(pm)-1 {
			b.WriteString(", ")
		}
	}
	return b.String()
}

// CompilePatterns compiles the patterns into a PathMatcher func that matches
// a path iff all patterns match the path.
func CompilePatterns(patterns []string, options CompileOptions) (PathMatcher, error) {
	matchers := make([]PathMatcher, len(patterns))
	for i, pattern := range patterns {
		matcher, err := CompilePattern(pattern, options)
		if err != nil {
			return nil, err
		}
		matchers[i] = matcher
	}

	if len(matchers) == 1 {
		return matchers[0], nil
	}

	return pathMatcherAnd(matchers), nil
}

// pathMatcherIncludeExclude is a PathMatcher that matches a path iff it matches
// the include matcher AND it does not match the exclude matcher.
type pathMatcherIncludeExclude struct {
	include PathMatcher
	exclude PathMatcher
}

func (pm pathMatcherIncludeExclude) MatchPath(path string) bool {
	include := pm.include == nil || pm.include.MatchPath(path)
	if !include {
		return false
	}

	exclude := pm.exclude != nil && pm.exclude.MatchPath(path)
	return !exclude
}

func (pm pathMatcherIncludeExclude) String() string {
	if pm.include != nil && pm.exclude != nil {
		return fmt.Sprintf("%s !%s", pm.include.String(), pm.exclude.String())
	}
	if pm.include != nil {
		return pm.include.String()
	}
	return "!" + pm.exclude.String()
}

// CompilePathPatterns returns a PathMatcher func that matches a path iff:
//
// * all of the includePatterns match the path; AND
// * the excludePattern does NOT match the path.
//
// This is the most common behavior for include/exclude paths in a search interface.
func CompilePathPatterns(includePatterns []string, excludePattern string, options CompileOptions) (PathMatcher, error) {
	var include PathMatcher
	if len(includePatterns) > 0 {
		var err error
		include, err = CompilePatterns(includePatterns, options)
		if err != nil {
			return nil, err
		}
	}

	var exclude PathMatcher
	if excludePattern != "" {
		var err error
		exclude, err = CompilePattern(excludePattern, options)
		if err != nil {
			return nil, err
		}
	}

	if include == nil && exclude == nil {
		return &pathMatcherFunc{
			matcher: func(path string) bool { return true },
			pattern: "noop",
		}, nil
	}
	if exclude == nil {
		return include, nil
	}
	return pathMatcherIncludeExclude{
		include: include,
		exclude: exclude,
	}, nil
}
