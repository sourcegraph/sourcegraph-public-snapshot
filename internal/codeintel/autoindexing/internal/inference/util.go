package inference

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/luatypes"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/paths"
)

// filterPathsByPatterns returns a slice containing all of the input paths that match
// any of the given path patterns. Both patterns and inverted patterns are considered
// when a path is matched.
func filterPathsByPatterns(paths []string, rawPatterns []*luatypes.PathPattern) ([]string, error) {
	patterns, _, err := flattenPatterns(rawPatterns, false)
	if err != nil {
		return nil, err
	}
	invertedPatterns, _, err := flattenPatterns(rawPatterns, true)
	if err != nil {
		return nil, err
	}

	return filterPaths(paths, patterns, invertedPatterns), nil
}

// flattenPatterns converts a tree of patterns into a flat list of compiled glob and pathspec patterns.
func flattenPatterns(patterns []*luatypes.PathPattern, inverted bool) ([]*paths.GlobPattern, []gitdomain.Pathspec, error) {
	var globPatterns []string
	var pathspecPatterns []string
	for _, pattern := range luatypes.FlattenPatterns(patterns, inverted) {
		globPatterns = append(globPatterns, pattern.Glob)
		pathspecPatterns = append(pathspecPatterns, pattern.Pathspecs...)
	}

	globs, err := compileWildcards(normalizePatterns(globPatterns))
	if err != nil {
		return nil, nil, err
	}

	var pathspecs []gitdomain.Pathspec
	for _, pathspec := range normalizePatterns(pathspecPatterns) {
		pathspecs = append(pathspecs, gitdomain.Pathspec(pathspec))
	}

	return globs, pathspecs, nil
}

// compileWildcards converts a list of wildcard strings into objects that can match inputs.
func compileWildcards(patterns []string) ([]*paths.GlobPattern, error) {
	compiledPatterns := make([]*paths.GlobPattern, 0, len(patterns))
	for _, rawPattern := range patterns {
		compiledPattern, err := paths.Compile(rawPattern)
		if err != nil {
			return nil, err
		}

		compiledPatterns = append(compiledPatterns, compiledPattern)
	}

	return compiledPatterns, nil
}

// normalizePatterns sorts the given slice and removes duplicate elements. This function
// modifies the given slice in place but also returns it to enable method chaining.
func normalizePatterns(patterns []string) []string {
	sort.Strings(patterns)

	filtered := patterns[:0]
	for _, pattern := range patterns {
		if n := len(filtered); n == 0 || filtered[n-1] != pattern {
			filtered = append(filtered, pattern)
		}
	}

	return filtered
}

// filterPaths returns a slice containing all of the input paths that match the given
// pattern but not the given inverted pattern. If the given inverted pattern is empty
// then it is not considered for filtering. The input slice is NOT modified in-place.
func filterPaths(paths []string, patterns, invertedPatterns []*paths.GlobPattern) []string {
	if len(patterns) == 0 {
		return nil
	}

	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		if filterPath(path, patterns, invertedPatterns) {
			filtered = append(filtered, path)
		}
	}

	return filtered
}

func filterPath(path string, pattern, invertedPattern []*paths.GlobPattern) bool {
	if path[0] != '/' {
		path = "/" + path
	}

	for _, p := range pattern {
		if p.Match(path) {
			// Matched an inclusion pattern; ensure we don't match an exclusion pattern
			return len(invertedPattern) == 0 || !filterPath(path, invertedPattern, nil)
		}
	}

	// We didn't match any inclusion pattern
	return false
}
