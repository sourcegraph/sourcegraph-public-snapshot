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

/*
ParsePattern is used internally by Goji to parse route patterns. It is exposed
publicly to make it easier to write thin wrappers around the built-in Pattern
implementations.

Although its parameter has type interface{}, ParsePattern only accepts arguments
of three types:
	- web.Pattern, which is passed through
	- string, which is interpreted as a Sinatra-like URL pattern. In
	  particular, the following syntax is recognized:
		- a path segment starting with with a colon will match any
		  string placed at that position. e.g., "/:name" will match
		  "/carl", binding "name" to "carl".
		- a pattern ending with "/*" will match any route with that
		  prefix. For instance, the pattern "/u/:name/*" will match
		  "/u/carl/" and "/u/carl/projects/123", but not "/u/carl"
		  (because there is no trailing slash). In addition to any names
		  bound in the pattern, the special key "*" is bound to the
		  unmatched tail of the match, but including the leading "/". So
		  for the two matching examples above, "*" would be bound to "/"
		  and "/projects/123" respectively.
	- regexp.Regexp, which is assumed to be a Perl-style regular expression
	  that is anchored on the left (i.e., the beginning of the string). If
	  your regular expression is not anchored on the left, a
	  hopefully-identical left-anchored regular expression will be created
	  and used instead. Named capturing groups will bind URLParams of the
	  same name; unnamed capturing groups will be bound to the variables
	  "$1", "$2", etc.

ParsePattern fatally exits (using log.Fatalf) if it is passed a value of an
unexpected type. It is the caller's responsibility to ensure that ParsePattern
is called in a type-safe manner.
*/
func ParsePattern(raw interface{}) Pattern {
	switch v := raw.(type) {
	case Pattern:
		return v
	case *regexp.Regexp:
		return parseRegexpPattern(v)
	case string:
		return parseStringPattern(v)
	default:
		log.Fatalf("Unknown pattern type %T. Expected a web.Pattern, "+
			"regexp.Regexp, or a string.", v)
	}
	panic("log.Fatalf does not return")
}
