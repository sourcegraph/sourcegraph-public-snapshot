// Package serverctx is a registry for context initialization
// functions that need to be run before each gRPC method call to
// initialize the server's context (e.g., by inserting DB handles,
// config, auth, etc.).
package serverctx

import "golang.org/x/net/context"

// Funcs are called to alter the ctx before responding to gRPC method
// calls.
//
// TODO(sqs): separate into funcs that run before EACH method call and
// funcs that run once at startup.
var Funcs []func(context.Context) (context.Context, error)
