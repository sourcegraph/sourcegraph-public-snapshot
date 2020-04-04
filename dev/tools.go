// +build tools

package main

import (
	// dev/go-install.sh has debug support
	_ "github.com/go-delve/delve/cmd/dlv"

	// zoekt-* used in sourcegraph/server docker image build
	_ "github.com/google/zoekt/cmd/zoekt-archive-index"
	_ "github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver"
	_ "github.com/google/zoekt/cmd/zoekt-webserver"

	// go-bindata is used in lots of our gen.go files
	_ "github.com/kevinburke/go-bindata/go-bindata"

	// goreman is used by our local dev environment
	_ "github.com/mattn/goreman"

	// vfsgendev is used for packing static assets into .go files.
	_ "github.com/shurcooL/vfsgen/cmd/vfsgendev"

	// used in schema pkg
	_ "github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler"

	// used in many places
	_ "golang.org/x/tools/cmd/stringer"
)
