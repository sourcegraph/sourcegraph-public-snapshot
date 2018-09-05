//docker:user sourcegraph

// Package frontend contains the enterprise frontend implementation.
//
// It wraps the open source frontend command and merely injects a few
// proprietary things into it via e.g. blank/underscore imports in this file
// which register side effects with the frontend package.
package main

import "github.com/sourcegraph/sourcegraph/cmd/frontend/shared"

func main() {
	shared.Main()
}
