package inference

import (
	"sort"

	"github.com/becheran/wildmatch-go"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference/luatypes"
)

// filterPathsByPatterns returns a slice containing all of the input paths that match
// any of the given path patterns. Both patterns and inverted patterns are considered
// when a path is matched.
func filterPathsByPatterns(paths []string, patterns []*luatypes.PathPattern) ([]string, error) {
	return filterPaths(
		paths,
		compileWildcards(flattenPatterns(patterns, false)),
		compileWildcards(flattenPatterns(patterns, true)),
	), nil
}

// flattenPatterns converts a tree of patterns into a flat list of compiled glob patterns.
func flattenPatterns(patterns []*luatypes.PathPattern, inverted bool) []string {
	return normalizePatterns(luatypes.FlattenPatterns(patterns, inverted))
}

// compileWildcards converts a list of wildcard strings into objects that can match inputs.
func compileWildcards(patterns []string) []*wildmatch.WildMatch {
	compiledPatterns := make([]*wildmatch.WildMatch, 0, len(patterns))
	for _, pattern := range patterns {
		compiledPatterns = append(compiledPatterns, wildmatch.NewWildMatch(pattern))
	}

	return compiledPatterns
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
func filterPaths(paths []string, patterns, invertedPatterns []*wildmatch.WildMatch) []string {
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

func filterPath(path string, pattern, invertedPattern []*wildmatch.WildMatch) bool {
	for _, p := range pattern {
		if p.IsMatch(path) {
			// Matched an inclusion pattern; ensure we don't match an exclusion pattern
			return len(invertedPattern) == 0 || !filterPath(path, invertedPattern, nil)
		}
	}

	// We didn't match any inclusion pattern
	return false
}
