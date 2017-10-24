// Package pathmatch provides helpers for matching paths against globs
// and regular expressions.
package pathmatch

import (
	"regexp"
	"strings"

	"github.com/gobwas/glob"
)

// PathMatcher reports whether the path was matched.
type PathMatcher func(path string) bool

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
		return p.MatchString, nil
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
		return func(path string) bool {
			return p.Match(strings.ToLower(path))
		}, nil
	}
	return p.Match, nil
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

	return func(path string) bool {
		for _, match := range matchers {
			if !match(path) {
				return false
			}
		}
		return true
	}, nil
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
		return All, nil
	}
	if include == nil {
		// Just negate the exclude func.
		return func(path string) bool {
			return !exclude(path)
		}, nil
	}
	if exclude == nil {
		return include, nil
	}

	return func(path string) bool {
		return include(path) && !exclude(path)
	}, nil
}

// All is a PathMatcher that matches all paths.
func All(path string) bool { return true }

// None is a PathMatcher that matches no paths.
func None(path string) bool { return false }
