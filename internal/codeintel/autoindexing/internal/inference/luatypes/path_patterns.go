package luatypes

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
)

// PathPattern is forms a tree of patterns to be matched against paths in an
// associated git repository.
type PathPattern struct {
	pattern  GlobAndPathspecPattern
	children []*PathPattern
	invert   bool
}

type GlobAndPathspecPattern struct {
	Glob      string
	Pathspecs []string
}

// NewPattern returns a new path pattern instance that includes a single pattern.
func NewPattern(glob string, pathspecs []string) *PathPattern {
	return &PathPattern{pattern: GlobAndPathspecPattern{
		Glob:      glob,
		Pathspecs: pathspecs,
	}}
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

// FlattenPatterns returns a concatenation of results from calling the function
// FlattenPattern on each of the inputs.
func FlattenPatterns(pathPatterns []*PathPattern, inverted bool) (patterns []GlobAndPathspecPattern) {
	for _, pathPattern := range pathPatterns {
		patterns = append(patterns, FlattenPattern(pathPattern, inverted)...)
	}

	return
}

// FlattenPattern returns the set of patterns matching the given inverted flag on this
// path pattern or any of its descendants.
func FlattenPattern(pathPattern *PathPattern, inverted bool) (patterns []GlobAndPathspecPattern) {
	if pathPattern.invert == inverted {
		if pathPattern.pattern.Glob != "" {
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
	return util.MapSliceOrSingleton(value, func(value lua.LValue) (*PathPattern, error) {
		return util.TypecheckUserData[*PathPattern](value, "*PathPattern")
	})
}
