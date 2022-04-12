package src

// This function exists only to make lsif-go succeed
// Without this function, lsif-go fails with the error message
// error: failed to index: failed to load packages: packages.Load: err: exit status 1: stderr: go build github.com/sourcegraph/sourcegraph/lib/codeintel/reprolang/src: build constraints exclude all Go files in /__w/sourcegraph/sourcegraph/lib/codeintel/reprolang/src
// The error appears related to the usage of CGO in the sibling file binding.go https://github.com/golang/go/issues/24068
// By adding this file, there exists at least one non-test file in this package that doesn't use CGO.
//lint:ignore U1000 This function is intentionally unused
func lsifWorkaround() {}
