package graph

import "strings"

const RootPackage = "github.com/sourcegraph/sourcegraph"

// trimPackage remvoes leading RootPackage from the given value.
func trimPackage(pkg string) string {
	return strings.TrimPrefix(strings.TrimPrefix(pkg, RootPackage), "/")
}
