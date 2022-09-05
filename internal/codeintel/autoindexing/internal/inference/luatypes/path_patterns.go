package luatypes

import (
	"strings"

	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

// PathPattern is forms a tree of patterns to be matched against paths in an
// associated git repository.
type PathPattern struct {
	pattern  string
	children []*PathPattern
	invert   bool
}

// NewPattern returns a new path pattern instance that includes a single pattern.
func NewPattern(pattern string) *PathPattern {
	return &PathPattern{pattern: pattern}
}

// NewCombinedPattern returns a new path pattern instance that includes the given
// set of patterns.
//
// Specifically: any path matching none of the given patterns is removed
func NewCombinedPattern(children []*PathPattern) *PathPattern {
	return &PathPattern{children: children}
}

// NewExcludePattern returns a new path pattern instance that excludes the given
// set of patterns.
//
// Specifically: any path matching one of the given patterns is removed
func NewExcludePattern(children []*PathPattern) *PathPattern {
	return &PathPattern{children: children, invert: true}
}

// CombinePatterns joins the given patterns together with a regex OR operator.
func CombinePatterns(patterns []string) string {
	return strings.Join(patterns, "|")
}

// FlattenPatterns returns a concatenation of results from calling the function
// FlattenPattern on each of the inputs.
func FlattenPatterns(pathPatterns []*PathPattern, inverted bool) (patterns []string) {
	for _, pathPattern := range pathPatterns {
		patterns = append(patterns, FlattenPattern(pathPattern, inverted)...)
	}

	return
}

// FlattenPattern returns the set of patterns matching the given inverted flag on this
// path pattern or any of its descendants.
func FlattenPattern(pathPattern *PathPattern, inverted bool) (patterns []string) {
	if pathPattern.invert == inverted {
		if pathPattern.pattern != "" {
			patterns = append(patterns, pathPattern.pattern)
		}

		for _, child := range pathPattern.children {
			patterns = append(patterns, FlattenPattern(child, inverted)...)
		}
	}

	return
}

// PathPatternsFromUserData decodes a single path pattern or slice of path patterns from
// the given Lua value.
func PathPatternsFromUserData(value lua.LValue) (patterns []*PathPattern, err error) {
	err = util.UnwrapSliceOrSingleton(value, func(value lua.LValue) error {
		return util.UnwrapLuaUserData(value, func(value any) error {
			if pathPattern, ok := value.(*PathPattern); ok {
				patterns = append(patterns, pathPattern)
				return nil
			}

			return util.NewTypeError("*PathPattern", value)
		})
	})

	return
}
