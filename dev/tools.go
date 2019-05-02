// +build tools

package main

import (
	_ "github.com/go-delve/delve/cmd/dlv"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/google/zoekt/cmd/zoekt-archive-index"
	_ "github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver"
	_ "github.com/google/zoekt/cmd/zoekt-webserver"
	_ "github.com/kevinburke/differ"
	_ "github.com/kevinburke/go-bindata/go-bindata"
	_ "github.com/mattn/goreman"
	_ "github.com/shurcooL/vfsgen/cmd/vfsgendev"
	_ "github.com/sourcegraph/docsite/cmd/docsite"
	_ "github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler"
	_ "golang.org/x/tools/cmd/stringer"
)
