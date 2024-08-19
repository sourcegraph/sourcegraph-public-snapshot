// Package grpcui provides a gRPC web UI in the form of HTTP handlers that can
// be added to a web server.
//
// This package provides multiple functions which, all combined, provide a fully
// functional web UI. Users of this package can use these pieces to embed a UI
// into any existing web application. The web form sources can be embedded in an
// existing HTML page, and the HTTP handlers wired up to make the form fully
// functional.
//
// For users that don't need as much control over layout and style of the web
// page, instead consider using standalone.Handler, which is an all-in-one
// handler that includes its own HTML and CSS as well as all other dependencies.
package grpcui
