# gRPC

As of Sourcegraph `5.3.X`, [gRPC](https://grpc.io/about/) has supplanted REST as our default mode of communication between our microservices for our internal APIs.

<Callout type="note">An "internal" API is one that's solely used for intra-service communication/RPCs (think `searcher` fetching an archive from `gitserver`). Internal APIs don't include things like the GraphQL API that external people can use (including our web interface).</Callout>

## gRPC Tutorial

The [`internal/grpc/example`](https://github.com/sourcegraph/sourcegraph/tree/main/internal/grpc/example) package in the [sourcegraph/sourcegraph monorepo](https://github.com/sourcegraph/sourcegraph) contains a simple, runnable example of a gRPC service and client. It is a good starting point for understanding how to write a gRPC service that covers the following topics:

- All the basic Protobuf types (e.g. primitives, enums, messages, one-ofs, etc.)
- All the basic RPC types (e.g. unary, server streaming, client streaming, bidirectional streaming)
- Error handling (e.g. gRPC errors, wrapping status errors, etc.)
- Implementing a gRPC server (with proper separation of concerns)
- Implementing a gRPC client
- Some known footguns (non-utf8 strings, huge messages, etc.)
- Some Sourcegraph-specific helper packages and patterns ([grpc/defaults](https://github.com/sourcegraph/sourcegraph/tree/main/internal/grpc/defaults), [grpc/streamio](https://github.com/sourcegraph/sourcegraph/tree/main/internal/grpc/streamio), etc.)

When going through this example for the first time, it is recommended to:

1. Read the protobuf definitions in [weather/v1/weather.proto](https://github.com/sourcegraph/sourcegraph/blob/main/internal/grpc/example/weather/v1/weather.proto) to get a sense of the service.
2. Run the server and client examples in [server/](https://github.com/sourcegraph/sourcegraph/tree/main/internal/grpc/example/server) and [client/](https://github.com/sourcegraph/sourcegraph/tree/main/internal/grpc/example/client) (via [server/run-server.sh](https://github.com/sourcegraph/sourcegraph/blob/main/internal/grpc/example/server/run-server.sh) and [client/run-client.sh](https://github.com/sourcegraph/sourcegraph/blob/main/internal/grpc/example/client/run-client.sh)) respectively to see the service in action. You can see a recording of this below:
    - [![asciicast](https://asciinema.org/a/wFAVGl59oxSWuLSazBgdpO5ks.svg)](https://asciinema.org/a/wFAVGl59oxSWuLSazBgdpO5ks)
3. Read the implementation of the [server](https://github.com/sourcegraph/sourcegraph/tree/main/internal/grpc/example/server) and [client](https://github.com/sourcegraph/sourcegraph/tree/main/internal/grpc/example/client) to get a sense of how things are implemented, and follow the explanatory comments in the code.
