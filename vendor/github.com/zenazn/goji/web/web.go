/*
Package web provides a fast and flexible middleware stack and mux.

This package attempts to solve three problems that net/http does not. First, it
allows you to specify flexible patterns, including routes with named parameters
and regular expressions. Second, it allows you to write reconfigurable
middleware stacks. And finally, it allows you to attach additional context to
requests, in a manner that can be manipulated by both compliant middleware and
handlers.

A usage example:

	m := web.New()

Use your favorite HTTP verbs and the interfaces you know and love from net/http:

	var legacyFooHttpHandler http.Handler // From elsewhere
	m.Get("/foo", legacyFooHttpHandler)
	m.Post("/bar", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})

Bind parameters using either named captures or regular expressions:

	m.Get("/hello/:name", func(c web.C, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s!", c.URLParams["name"])
	})
	pattern := regexp.MustCompile(`^/ip/(?P<ip>(?:\d{1,3}\.){3}\d{1,3})$`)
	m.Get(pattern, func(c web.C, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Info for IP address %s:", c.URLParams["ip"])
	})

Middleware are functions that wrap http.Handlers, just like you'd use with raw
net/http. Middleware functions can optionally take a context parameter, which
will be threaded throughout the middleware stack and to the final handler, even
if not all of the intervening middleware or handlers support contexts.
Middleware are encouraged to use the Env parameter to pass request-scoped data
to other middleware and to the final handler:

	func LoggerMiddleware(h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			log.Println("Before request")
			h.ServeHTTP(w, r)
			log.Println("After request")
		}
		return http.HandlerFunc(handler)
	}
	func AuthMiddleware(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if cookie, err := r.Cookie("user"); err == nil {
				c.Env["user"] = cookie.Value
			}
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	}

	// This makes the AuthMiddleware above a little cleaner
	m.Use(middleware.EnvInit)
	m.Use(AuthMiddleware)
	m.Use(LoggerMiddleware)
	m.Get("/baz", func(c web.C, w http.ResponseWriter, r *http.Request) {
		if user, ok := c.Env["user"].(string); ok {
			w.Write([]byte("Hello " + user))
		} else {
			w.Write([]byte("Hello Stranger!"))
		}
	})
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
	// URL Pattern you are using (or the documentation for ParsePattern for
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
// object. Implements both http.Handler and web.Handler.
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
