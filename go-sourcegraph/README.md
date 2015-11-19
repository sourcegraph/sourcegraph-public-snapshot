# go-sourcegraph [![Build Status](https://travis-ci.org/sourcegraph/go-sourcegraph.png?branch=master)](https://travis-ci.org/sourcegraph/go-sourcegraph)

[Sourcegraph](https://sourcegraph.com) API client library for [Go](http://golang.org).

**Work in progress. If you want to use this, [post an issue](https://github.com/sourcegraph/go-sourcegraph/issues) or contact us [@srcgraph](https://twitter.com/srcgraph).**

## Development

### Protocol buffers

This repository uses the `sourcegraph/sourcegraph.proto` [protocol buffers](https://developers.google.com/protocol-buffers/) definition file to generate Go structs as well as [gRPC](http://grpc.io) clients and servers for various service interfaces.

### First-time installation of protobuf and other codegen tools

You need to install and run the protobuf compiler before you can regenerate Go code after you change the `sourcegraph.proto` file.

If you run into errors while compiling protobufs, try again with these versions that are known to work:

-  `protoc` - version `github.com/google/protobuf@v3.0.0-beta-1`.
-  `protoc-gen-gogo` - commit `github.com/gogo/protobuf@200875106f3bf0eb01eb297dae30b250a25ffc84`.
-  `grpc-go` - commit `google.golang.org/grpc@f7d1653e300d6ad9f019bce7a5f5ab3b4821f637`.

1. **Install protoc**, the protobuf compiler. Find more details in the [protobuf README](https://github.com/google/protobuf/tree/v3.0.0-beta-1#c-installation---unix).

   Make sure the `protoc` binary is in your `$PATH`.

2. **Install [gogo/protobuf](https://github.com/gogo/protobuf)**.

   ```
   go get -u github.com/gogo/protobuf/...
   ```

3. **Install [grpc](https://github.com/grpc/grpc-go)**:

   ```
   go get google.golang.org/grpc
   ```

4. **Install [gen-mocks](https://sourcegraph.com/sourcegraph/gen-mocks)** by running:

   ```
   go get -u sourcegraph.com/sourcegraph/gen-mocks
   ```

5. **Install `gopathexec`**:

   ```
   go get -u sourcegraph.com/sourcegraph/gopathexec
   ```

6. **Install `grpccache-gen`**:

   ```
   go get -u sourcegraph.com/sourcegraph/grpccache/grpccache-gen
   ```

### Regenerating Go code after changing `sourcegraph.proto`

1. In `go-sourcegraph` (this repository), run:

   ```
   go generate ./...
   ```
