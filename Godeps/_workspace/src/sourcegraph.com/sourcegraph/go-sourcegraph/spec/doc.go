// Package spec contains regexps, parse functions, and stringification
// functions for user, repo, repo rev, etc., specifiers.
//
// If you're using patterns defined in this package in a mux route
// path definition, you should probably use the definitions in the
// sibling routevar package instead. If you use this package's regexps
// in a mux route, be sure to call routevar.NamedToNonCapturingGroups
// on the regexp.
//
// It is independent of URL routing and protobufs for simplicity.
package spec
