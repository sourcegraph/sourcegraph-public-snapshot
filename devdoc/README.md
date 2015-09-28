# Developers

This directory contains the code for developer.sourcegraph.com, the documentation site for the Sourcegraph API.

The documentation is generated on-the-fly from a binary-dump (protobuf encoded) representation of the protobuf files (`assets/sourcegraph.dump`). This dump file representation is built by [protoc-gen-dump](https://sourcegraph.com/sourcegraph/prototools). That is, documentation is auto-generated from the `.proto` API definition source files indirectly such that `protoc` is not invoked when the program is running (as it requires a complex environment setup).

Only release binaries (i.e. ones built with `make dist`) will contain the `sourcegraph.dump` file: other binaries will not contain documentation.

# Generating

To regenerate the documentation you'll need to do two things:

1. Generate the `sourcegraph.dump` documentation file using `protoc-gen-dump`.
2. Repackage the assets into the binary (or use `-tags=dev` when building) using `go generate`.

Running `protoc` is unfortunately a bit tedious due to the environment-specific paths, so a `Makefile` is provided in this directory to streamline the process.

- `make dist` produces the `sourcegraph.dump` file and runs `go generate` for you so that any future binary built (`./cmd/developer` or `../cmd/src`) will have documentation built into it.
- `make dev` works just like `make dist` except it pulls from the `.proto` files inside your ``$GOPATH`, useful during development of the API.
- `make clean` will remove the `sourcegraph.dump` file and runs `go generate` for you so that any future binary built will not have documentation built in.

The parent Makefile (the one that builds `../cmd/src`) will automatically do the correct thing when running `make distclean` (which runs `devdocs` `make clean`) or `make dist` which will produce `src` binaries with developer docs built in.

# Development

You can use the `-tags=dev` flag when building this package (i.e. `./cmd/developer` or `../cmd/src`) to have go-bindata load assets from disk. Because you need to regenerate the `sourcegraph.dump` each time you change the API, you can just build with `-tags=dev` and re-run `make dev` to regenerate the dump file (then just refresh the page in your browser to see the changes).

The developer site can be ran independent of everything else with e.g. `go run -tags=dev ./cmd/developer/main.go`.

Note: Ensure that you run `make dist` before committing your changes, as the `sourcegraph.dump` file must be updated / committed!
