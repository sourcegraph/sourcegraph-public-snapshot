# Developers

This directory contains the code for developer.sourcegraph.com, the documentation site for the Sourcegraph API.

The documentation is generated on-the-fly from a binary-dump (protobuf encoded)
representation of Sourcegraph protobuf files (`assets/data/sourcegraph.dump`).
This dump file representation is built by [protoc-gen-dump](https://github.com/sourcegraph/prototools). That is, documentation is auto-generated from the `.proto` API definition source files
indirectly such that `protoc` is not invoked when the program is running (as it requires a
complex environment setup).

Development binaries (i.e., ones built without `dist` build tag) will have documentation
only if `assets/data/sourcegraph.dump` is already generated.

# Generating

To regenerate the documentation you'll need to:

1. Generate the `sourcegraph.dump` documentation file using `protoc-gen-dump`.
2. Repackage the assets into the binary (or use `-tags=dev` when building) using `go generate`.

Running `protoc` is unfortunately a bit tedious due to the environment-specific paths, so
go generate directives are relied on to streamline the process.

- `go generate ./...` produces the `sourcegraph.dump` file so that any future binary built
(`./cmd/developer` or `../cmd/src`) will have documentation built into it.

The parent Makefile (the one that builds `../cmd/src`) will automatically do the correct
thing when running `make dist` which will produce `src` binaries with developer docs built in.

# Development

You can use the `-tags=` flag when building this package (i.e. `./cmd/developer` or `../cmd/src`)
to have it load assets from disk. Because you need to regenerate the `sourcegraph.dump` each time
you change the API, you can just build with `-tags=` and re-run `go generate ./...` to regenerate
the dump file (then just refresh the page in your browser to see the changes).

The developer site can be run independent of everything else with e.g. `go run ./cmd/developer/main.go`.
