// Package serverctx is a registry for context initialization
// functions that need to be run before each gRPC method call to
// initialize the server's context (e.g., by inserting DB handles,
// config, auth, etc.).
package serverctx

import "golang.org/x/net/context"

// Funcs are called to alter the ctx before responding to gRPC method
// calls. These funcs may not rely on any other values to be set in
// the ctx; if the func depends on other values (e.g., the actor) to
// be in the ctx, then add it to LastFuncs.
var Funcs []func(context.Context) (context.Context, error)

// LastFuncs are called to alter the ctx before responding to gRPC
// method calls but AFTER all of the funcs in Funcs are called and
// AFTER the actor is set in the ctx. These funcs may retrieve values
// from the ctx that the Funcs set.
var LastFuncs []func(context.Context) (context.Context, error)
