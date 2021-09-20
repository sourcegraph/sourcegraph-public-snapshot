package search

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"testing"
)

func BenchmarkGitSearch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		predicate := &And{[]MatchTree{
			&AuthorMatches{Regexp{regexp.MustCompile("camden")}},
			&DiffMatches{Regexp{regexp.MustCompile("camden")}},
		}}

		buf := bufio.NewWriter(os.Stdout)
		err := Search("/Users/ccheek/src/sourcegraph/sourcegraph", nil, predicate, func(lc *LazyCommit, hl *HighlightedCommit) bool {
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
