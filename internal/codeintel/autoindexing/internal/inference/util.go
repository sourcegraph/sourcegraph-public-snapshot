package inference

import (
	"sort"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference/luatypes"
)

// filterPathsByPatterns returns a slice containing all of the input paths that match
// any of the given path patterns. Both patterns and inverted patterns are considered
// when a path is matched.
func filterPathsByPatterns(paths []string, patterns []*luatypes.PathPattern) ([]string, error) {
	pattern, err := flattenPatterns(patterns, false)
	if err != nil {
		return nil, err
	}

	invertedPattern, err := flattenPatterns(patterns, true)
	if err != nil {
		return nil, err
	}

	return filterPaths(paths, pattern, invertedPattern), nil
}

// flattenPatterns returns a single regular expression composed of an alternation of
// all patterns reachable from the given path pattern.
func flattenPatterns(patterns []*luatypes.PathPattern, inverted bool) (*regexp.Regexp, error) {
	return regexp.Compile(luatypes.CombinePatterns(normalizePattterns(luatypes.FlattenPatterns(patterns, inverted))))
}

// normalizePattterns sorts the given slice and removes duplicate elements. This function
// modifies the given slice in place but also returns it to enable method chaining.
func normalizePattterns(patterns []string) []string {
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
func filterPaths(paths []string, pattern, invertedPattern *regexp.Regexp) []string {
	if pattern.String() == "" {
		return nil
	}

	filtered := make([]string, 0, len(paths))
	for _, path := range paths {
		if pattern.MatchString(path) && (invertedPattern.String() == "" || !invertedPattern.MatchString(path)) {
			filtered = append(filtered, path)
		}
	}

	return filtered
}
