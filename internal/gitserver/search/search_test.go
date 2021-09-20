package search

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func BenchmarkGitSearch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		predicate := &protocol.And{[]protocol.SearchQuery{
			&protocol.AuthorMatches{protocol.Regexp{regexp.MustCompile("camden")}},
			&protocol.DiffMatches{protocol.Regexp{regexp.MustCompile("camden")}},
		}}

		buf := bufio.NewWriter(os.Stdout)
		err := Search("/Users/ccheek/src/sourcegraph/sourcegraph", nil, ToMatchTree(predicate), func(lc *LazyCommit, hl *protocol.HighlightedCommit) bool {
			if idx := bytes.IndexByte(lc.Message, '\n'); idx < 0 {
				buf.Write(lc.Message)
			} else {
				buf.Write(lc.Message[:idx+1])
			}
			return true
		})
		if err != nil {
			panic(err)
		}
		buf.Flush()
	}
}
