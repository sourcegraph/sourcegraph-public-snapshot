// Package federated wraps server methods (using codegen) and uses the
// method args to determine where to route the method call, to allow
// for federation.
//
// For example, consider the Repos.Get method. If the repo URI arg is
// "src:///foo", then the repo should be fetched locally. If the arg
// is "src://example.com/bar", then it should be fetched from
// example.com.
//
// This package contains 3 components:
//
// 1. The server method wrapper codegen (regenerate using `go
//    generate`).
//
// 2. The automated mapping from the arg to the expression
//    representing the repo URI or user spec. This mapping is
//    performed during codegen, but it's a separate concept. This is
//    implemented in package gen.
//
//    For example, given a gRPC method Defs.Get that accepts an arg of
//    type *DefGetOptions, how do we derive the repo URI (which is
//    required to know where to route the call) from the arg? That is
//    determined at codegen-time.
//
// 3. The functions (lookupRepo and lookupUser) that accept a repo URI
//    (or user spec) and return the context.Context that should be
//    used. This context might, e.g., hit a remote server or consult
//    the local methods but using a different underlying store.
package federated
