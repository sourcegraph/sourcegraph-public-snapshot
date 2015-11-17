# Protocol Buffers

Sourcegraph is built with [gRPC](http://grpc.io), an HTTP2-based RPC
protocol that uses Protocol Buffer service definitions. We publish Sourcegraph
API clients in all languages supported by gRPC
[including Go](https://src.sourcegraph.com/go-sourcegraph).

`go-sourcegraph` is the canonical repository for Sourcegraph's protobuf definitions;
it doubles as our golang Sourcegraph API client.

The [`go-sourcegraph` README](https://src.sourcegraph.com/go-sourcegraph) has more
information about working with Sourcegraph's protocol buffers.

This repository (`sourcegraph`) also has code generation that depends on the protobuf
definitions. Whenever you add or change the protobuf file in `go-sourcegraph` or any
of the `package store` interfaces, you must rerun `go generate ./...` in this repository.

You will need `gen-mocks`, which you can install by running
`go get -u sourcegraph.com/sourcegraph/gen-mocks`.

## Documentation

API documentation is auto-generated based on the protobuf sources and is available for viewing online at [developer.sourcegraph.com](https://developer.sourcegraph.com) and also in local development instances at e.g. `http://localhost:3080/.docs`.
