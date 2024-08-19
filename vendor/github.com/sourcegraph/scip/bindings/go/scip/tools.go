//go:build tools
// +build tools

package scip

// Hack for go.mod not supporting "dev dependencies" as a feature.
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

// Keep in sync with dev/proto-generate.sh
import (
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc"
	_ "golang.org/x/tools/cmd/goimports"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
