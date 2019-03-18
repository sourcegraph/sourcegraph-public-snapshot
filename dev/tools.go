// +build tools

package main

import (
	_ "github.com/google/zoekt/cmd/zoekt-archive-index"
	_ "github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver"
	_ "github.com/google/zoekt/cmd/zoekt-webserver"
	_ "github.com/kevinburke/differ"
	_ "github.com/kevinburke/go-bindata/go-bindata"
	_ "github.com/mattn/goreman"
	_ "github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler"
	_ "github.com/sourcegraph/godockerize"
	_ "golang.org/x/tools/cmd/stringer"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
