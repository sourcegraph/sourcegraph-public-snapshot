// Package svc stores and retrieves gRPC services from the context.
//
// Currently it is only intended for use by server-side code. Client
// gRPC services are contained in the sourcegraph.Client struct's
// fields. It is exported (and not internal underneath server) so that
// in the future both client-side and server-side code may use it.
//
// TODO: Unify the client- and server-side code so that they both can
// use this package instead of needing to put every gRPC client as a
// field on the sourcegraph.Client struct.
package svc
