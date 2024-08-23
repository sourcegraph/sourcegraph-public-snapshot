package web

import (
	"log"
	"net/http"
	"regexp"
)

// A Pattern determines whether or not a given request matches some criteria.
// They are often used in routes, which are essentially (pattern, methodSet,
// handler) tuples. If the method and pattern match, the given handler is used.
//
// Built-in implementations of this interface are used to implement regular
// expression and string matching.
type Pattern interface {
	// In practice, most real-world routes have a string prefix that can be
	// used to quickly determine if a pattern is an eligible match. The
	// router uses the result of this function to optimize away calls to the
	// full Match function, which is likely much more expensive to compute.
	// If your Pattern does not support prefixes, this function should
	// return the empty string.
	Prefix() string
	// Returns true if the request satisfies the pattern. This function is
	// free to examine both the request and the context to make this
	// decision. Match should not modify either argument, and since it will
	// potentially be called several times over the course of matching a
	// request, it should be reasonably efficient.
	Match(r *http.Request, c *C) bool
	// Run the pattern on the request and context, modifying the context as
	// necessary to bind URL parameters or other parsed state.
	Run(r *http.Request, c *C)
}

const unknownPattern = `Unknown pattern type %T. See http://godoc.org/github.com/zenazn/goji/web#PatternType for a list of acceptable types.`

/*
ParsePattern is used internally by Goji to parse route patterns. It is exposed
publicly to make it easier to write thin wrappers around the built-in Pattern
implementations.

ParsePattern fatally exits (using log.Fatalf) if it is passed a value of an
unexpected type (see the documentation for PatternType for a list of which types
are accepted). It is the caller's responsibility to ensure that ParsePattern is
called in a type-safe manner.
*/
func ParsePattern(raw PatternType) Pattern {
	switch v := raw.(type) {
	case Pattern:
		return v
	case *regexp.Regexp:
		return parseRegexpPattern(v)
	case string:
		return parseStringPattern(v)
	default:
		log.Fatalf(unknownPattern, v)
		panic("log.Fatalf does not return")
	}
}
