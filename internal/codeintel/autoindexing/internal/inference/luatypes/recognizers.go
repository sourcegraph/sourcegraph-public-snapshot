package luatypes

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/sourcegraph/sourcegraph/internal/luasandbox/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Recognizer is a Go struct that is constructed and manipulated by Lua
// scripts via UserData values. This struct can take one of two mutually
// exclusive forms:
//
//	(1) An applicable recognizer with patterns and a generate function.
//	(2) A fallback recognizer, which consists of a list of children.
//	    Execution of a fallback recognizer will invoke its children,
//	    in order and recursively, until the non-empty value is yielded.
type Recognizer struct {
	patterns           []*PathPattern
	patternsForContent []*PathPattern
	generate           *lua.LFunction
	hints              *lua.LFunction
	fallback           []*Recognizer
}

// Patterns returns the set of filepath patterns in which this recognizer is
// interested. If the given forContent flag is true, then the patterns will
// consist of the files for which we want full content. By default we only
// need to expand the path patterns to concrete repo-relative file paths.
func (r *Recognizer) Patterns(forContent bool) []*PathPattern {
	if forContent {
		return r.patternsForContent
	}

	return r.patterns
}

// Generator returns the registered Lua generate callback and its suspended environment.
func (r *Recognizer) Generator() *lua.LFunction {
	return r.generate
}

// Hinter returns the registered Lua hints callback and its suspended environment.
func (r *Recognizer) Hinter() *lua.LFunction {
	return r.hints
}

// NewFallback returns a new fallback recognizer.
func NewFallback(fallback []*Recognizer) *Recognizer {
	return &Recognizer{fallback: fallback}
}

// FlattenRecognizerPatterns returns a concatenation of results from calling the function
// FlattenRecognizerPattern on each of the inputs.
func FlattenRecognizerPatterns(recognizers []*Recognizer, forContent bool) (flatten []*PathPattern) {
	for _, recognizer := range recognizers {
		flatten = append(flatten, FlattenRecognizerPattern(recognizer, forContent)...)
	}

	return
}

// FlattenRecognizerPattern flattens all patterns reachable from the given recognizer.
func FlattenRecognizerPattern(recognizer *Recognizer, forContent bool) (patterns []*PathPattern) {
	patterns = append(patterns, recognizer.Patterns(forContent)...)

	for _, recognizer := range recognizer.fallback {
		patterns = append(patterns, FlattenRecognizerPattern(recognizer, forContent)...)
	}

	return
}

// LinearizeGenerators returns a concatenation of results from calling the function
// LinearizeRecognizer on each of the inputs.
func LinearizeGenerators(recognizers []*Recognizer) (linearized []*Recognizer) {
	for _, recognizer := range recognizers {
		linearized = append(linearized, LinearizeGenerator(recognizer)...)
	}

	return
}

// LinearizeGenerator returns the depth-first ordering of recognizers whose generate functions
// should be invoked in order of fallback. If this is not a fallback recognizer, it should invoke
// only itself. All recognizers returned by this function should have an associated non-nil
// generate function.
func LinearizeGenerator(recognizer *Recognizer) (recognizers []*Recognizer) {
	if recognizer.generate != nil {
		recognizers = append(recognizers, recognizer)
	}

	for _, recognizer := range recognizer.fallback {
		recognizers = append(recognizers, LinearizeGenerator(recognizer)...)
	}

	return
}

// LinearizeHinters returns a concatenation of results from calling the function
// LinearizeHinter on each of the inputs.
func LinearizeHinters(recognizers []*Recognizer) (linearized []*Recognizer) {
	for _, recognizer := range recognizers {
		linearized = append(linearized, LinearizeHinter(recognizer)...)
	}

	return
}

// LinearizeHinter returns the depth-first ordering of recognizers whose hints functions
// should be invoked in order of fallback. If this is not a fallback recognizer, it should invoke
// only itself. All recognizers returned by this function should have an associated non-nil
// hints function.
func LinearizeHinter(recognizer *Recognizer) (recognizers []*Recognizer) {
	if recognizer.hints != nil {
		recognizers = append(recognizers, recognizer)
	}

	for _, recognizer := range recognizer.fallback {
		recognizers = append(recognizers, LinearizeHinter(recognizer)...)
	}

	return
}

// NamedRecognizersFromUserDataMap decodes a keyed map of recognizers from the given Lua value.
// If allowFalseAsNil is true, then a `false` value for a recognizer will be interpreted as a
// nil recognizer value in Go. This is to allow the user to disable the built-in recognizers.
func NamedRecognizersFromUserDataMap(value lua.LValue, allowFalseAsNil bool) (recognizers map[string]*Recognizer, err error) {
	recognizers = map[string]*Recognizer{}

	err = util.ForEach(value, func(key, value lua.LValue) error {
		name := key.String()

		if value.Type() == lua.LTBool && !lua.LVAsBool(value) {
			if allowFalseAsNil {
				recognizers[name] = nil
				return nil
			}
		}

		return util.UnwrapLuaUserData(value, func(value any) error {
			recognizer, ok := value.(*Recognizer)
			if !ok {
				return util.NewTypeError("*Recognizer", value)
			}

			recognizers[name] = recognizer
			return nil
		})
	})

	return
}

// RecognizersFromUserData decodes a single recognize or slice of recognizers from the
// given Lua value.
func RecognizersFromUserData(value lua.LValue) (recognizers []*Recognizer, err error) {
	err = util.UnwrapSliceOrSingleton(value, func(value lua.LValue) error {
		return util.UnwrapLuaUserData(value, func(value any) error {
			if recognizer, ok := value.(*Recognizer); ok {
				recognizers = append(recognizers, recognizer)
				return nil
			}

			return util.NewTypeError("*Recognizer", value)
		})
	})

	return
}

// RecognizerFromTable decodes a single Lua table value into a recognizer instance.
func RecognizerFromTable(table *lua.LTable) (*Recognizer, error) {
	recognizer := &Recognizer{}

	if err := util.DecodeTable(table, map[string]func(lua.LValue) error{
		"patterns":             setPathPatterns(&recognizer.patterns),
		"patterns_for_content": setPathPatterns(&recognizer.patternsForContent),
		"generate":             util.SetLuaFunction(&recognizer.generate),
		"hints":                util.SetLuaFunction(&recognizer.hints),
	}); err != nil {
		return nil, err
	}

	if recognizer.generate == nil && recognizer.hints == nil {
		return nil, errors.Newf("no generate or hints function supplied - at least one is required")
	}
	if recognizer.patterns == nil && recognizer.patternsForContent == nil {
		return nil, errors.Newf("no patterns supplied")
	}

	return recognizer, nil
}

// setPathPatterns returns a decoder function that updates the given path patterns
// slice value on invocation. For use in util.DecodeTable.
func setPathPatterns(ptr *[]*PathPattern) func(lua.LValue) error {
	return func(value lua.LValue) (err error) {
		*ptr, err = PathPatternsFromUserData(value)
		return
	}
}
