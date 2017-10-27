/*
Package web provides a fast and flexible middleware stack and mux.

This package attempts to solve three problems that net/http does not. First, it
allows you to specify flexible patterns, including routes with named parameters
and regular expressions. Second, it allows you to write reconfigurable
middleware stacks. And finally, it allows you to attach additional context to
requests, in a manner that can be manipulated by both compliant middleware and
handlers.
*/
package web

import (
	"net/http"
)

/*
C is a request-local context object which is threaded through all compliant
middleware layers and given to the final request handler.
*/
type C struct {
	// URLParams is a map of variables extracted from the URL (typically
	// from the path portion) during routing. See the documentation for the
	// URL Pattern you are using (or the documentation for PatternType for
	// the case of standard pattern types) for more information about how
	// variables are extracted and named.
	URLParams map[string]string
	// Env is a free-form environment for storing request-local data. Keys
	// may be arbitrary types that support equality, however package-private
	// types with type-safe accessors provide a convenient way for packages
	// to mediate access to their request-local data.
	Env map[interface{}]interface{}
}

// Handler is similar to net/http's http.Handler, but also accepts a Goji
// context object.
type Handler interface {
	ServeHTTPC(C, http.ResponseWriter, *http.Request)
}

// HandlerFunc is similar to net/http's http.HandlerFunc, but supports a context
// object. Implements both http.Handler and Handler.
type HandlerFunc func(C, http.ResponseWriter, *http.Request)

// ServeHTTP implements http.Handler, allowing HandlerFunc's to be used with
// net/http and other compliant routers. When used in this way, the underlying
// function will be passed an empty context.
func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(C{}, w, r)
}

// ServeHTTPC implements Handler.
func (h HandlerFunc) ServeHTTPC(c C, w http.ResponseWriter, r *http.Request) {
	h(c, w, r)
}

/*
PatternType is the type denoting Patterns and types that Goji internally
converts to Pattern (via the ParsePattern function). In order to provide an
expressive API, this type is an alias for interface{} that is named for the
purposes of documentation, however only the following concrete types are
accepted:
	- types that implement Pattern
	- string, which is interpreted as a Sinatra-like URL pattern. In
	  particular, the following syntax is recognized:
		- a path segment starting with a colon will match any
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
	  Unlike http.ServeMux's patterns, string patterns support neither the
	  "rooted subtree" behavior nor Host-specific routes. Users who require
	  either of these features are encouraged to compose package http's mux
	  with the mux provided by this package.
	- regexp.Regexp, which is assumed to be a Perl-style regular expression
	  that is anchored on the left (i.e., the beginning of the string). If
	  your regular expression is not anchored on the left, a
	  hopefully-identical left-anchored regular expression will be created
	  and used instead.

	  Capturing groups will be converted into bound URL parameters in
	  URLParams. If the capturing group is named, that name will be used;
	  otherwise the special identifiers "$1", "$2", etc. will be used.
*/
type PatternType interface{}

/*
HandlerType is the type of Handlers and types that Goji internally converts to
Handler. In order to provide an expressive API, this type is an alias for
interface{} that is named for the purposes of documentation, however only the
following concrete types are accepted:
	- types that implement http.Handler
	- types that implement Handler
	- func(http.ResponseWriter, *http.Request)
	- func(web.C, http.ResponseWriter, *http.Request)
*/
type HandlerType interface{}

/*
MiddlewareType is the type of Goji middleware. In order to provide an expressive
API, this type is an alias for interface{} that is named for the purposes of
documentation, however only the following concrete types are accepted:
	- func(http.Handler) http.Handler
	- func(*web.C, http.Handler) http.Handler
*/
type MiddlewareType interface{}
