package validation

import (
	reader "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/protocol/reader"
)

// forEachInV calls the given function on each sink vertex adjacent to the given
// edge. If any invocation returns false, iteration of the adjacent vertices will
// not complete and false will be returned immediately.
func forEachInV(edge reader.Edge, f func(inV int) bool) bool {
	if edge.InV != 0 {
		if !f(edge.InV) {
			return false
		}
	}
	for _, inV := range edge.InVs {
		if !f(inV) {
			return false
		}
	}

	return true
}

// eachInV returns a slice containing the InV/InVs values of the given edge.
func eachInV(edge reader.Edge) (inVs []int) {
	_ = forEachInV(edge, func(inV int) bool {
		inVs = append(inVs, inV)
		return true
	})

	return inVs
}
