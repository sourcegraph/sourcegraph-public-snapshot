//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/shurcooL/vfsgen/cmd/vfsgendev
//go:generate vfsgendev -source="github.com/sourcegraph/sourcegraph/cmd/management-console/assets".Assets

// Package assets serves static management console assets.
//
// It serves them from disk when built with the Go build tag "dev", and packs them
// into the binary (see `go generate` directives) otherwise.
package assets
