# rpc  [![Build Status](https://travis-ci.org/keegancsmith/rpc.svg?branch=master)](https://travis-ci.org/keegancsmith/rpc)

This is a fork of the stdlib [net/rpc](https://golang.org/pkg/net/rpc/) which
is frozen. It adds support for `context.Context` on the client and server,
including propagating cancellation.

The API is exactly the same, except `Client.Call` takes a `context.Context`,
and Server methods are expected to take a `context.Context` as the first
argument. Additionally the wire protocol is unchanged, so is backwards
compatible with `net/rpc` clients.

`DialHTTPPathTimeout` function is also added. A future release of rpc may
update all Dial functions to instead take a context.

## Why use net/rpc

There are many alternatives for RPC in Go, the most popular being
[GRPC](https://grpc.io/). However, `net/rpc` has the following nice
properties:

- Nice API
- No need for IDL
- Good performance

The nice API is subjective. However, the API is small, simple and composable.
which makes it quite powerful. IDL tools are things like GRPC requiring protoc
to generate go code from the protobuf files. `net/rpc` has no third party
dependencies nor code generation step, simplify the use of it. A benchmark
done on the [6 Sep
2016](https://github.com/golang/go/issues/16844#issuecomment-245261755)
indicated `net/rpc` was 4x faster than GRPC. This is an outdated benchmark,
but is an indication at the surprisingly good performance `net/rpc` provides.

For more discussion on the pros and cons of `net/rpc` see the issue [proposal:
freeze net/rpc](https://github.com/golang/go/issues/16844).

## Details

Last forked from commit [08ab820](https://github.com/golang/go/commit/08ab820)
on 25 September 2018.

Cancellation implemented via the rpc call `_goRPC_.Cancel`.
