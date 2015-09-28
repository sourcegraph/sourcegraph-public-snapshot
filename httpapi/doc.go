// Package httpapi contains the HTTP API, which implements a subset of
// the operations exposed by the gRPC API. Currently it is only
// accessed by JavaScript running on users's Web browsers on
// Sourcegraph.com.
//
// It is an unprivileged, untrusted API client of the gRPC API; it
// does not have any access beyond that granted by the HTTP request's
// authentication credentials. See docs/Security.md for more
// information.
package httpapi
