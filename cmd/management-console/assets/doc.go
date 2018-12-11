// Package assets serves static management console assets.
//
// It serves them from disk when built with the Go build tag "dev", and packs them
// into the binary (see `go generate` directives) otherwise.
package assets
