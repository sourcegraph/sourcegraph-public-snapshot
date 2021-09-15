package search

import (
	"testing"
)

func BenchmarkGitSearch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Search("/Users/ccheek/src/sourcegraph/sourcegraph", nil, nil, nil)
	}
}
