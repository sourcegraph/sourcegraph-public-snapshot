# Protobuf readme

See the go-sourcegraph README (on the nodb branch) for instructions on regenerating the Go code when you change `sourcegraph.proto`.

This repository (`sourcegraph`) also has code generation that depends on the protobuf definitions. Whenever you add or change the protobuf file in `go-sourcegraph` or any of the `package store` interfaces, you must rerun `go generate ./...` in this repository.

You will need `gen-mocks`, which you can install by running `go get -u sourcegraph.com/sourcegraph/gen-mocks`.
