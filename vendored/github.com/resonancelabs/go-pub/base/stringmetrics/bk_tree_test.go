package stringmetrics

import (
	"sort"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/vendored/github.com/resonancelabs/go-pub/base/imath"

	"gopkg.in/check.v1"
)

const MAX_WORD_LENGTH = 100

// Hook up check to "go test"
func Test(t *testing.T) { check.TestingT(t) }

func lengthMetric(a, b string) int {
	return imath.Abs(len(a) - len(b))
}

type bkTreeTestSuite struct {
	bkTree *BkTree
	words  []string
}

func (s *bkTreeTestSuite) SetUpSuite(c *check.C) {
	// build a list of words and insert into the tree
	s.words = []string{}
	for i, w := 0, ""; i < MAX_WORD_LENGTH; i++ {
		s.words = append(s.words, w)
		w = w + "a"
	}

	s.bkTree = NewBkTree(s.words, lengthMetric)
}

func (s *bkTreeTestSuite) referenceAnswer(query string, distance int) []string {
	reference := []string{}
	for _, w := range s.words {
		if lengthMetric(query, w) <= distance {
			reference = append(reference, w)
		}
	}
	sort.Strings(reference)
	return reference
}

func (s *bkTreeTestSuite) TestRetrieval(c *check.C) {
	for _, w := range s.words {
		for d := 0; d < MAX_WORD_LENGTH; d++ {
			ref := s.referenceAnswer(w, d)
			tree := s.bkTree.FindSimilarWords(w, d)
			sort.Strings(tree)
			c.Assert(ref, check.DeepEquals, tree)
		}
	}
}

func (s *bkTreeTestSuite) BenchmarkRetrieval(c *check.C) {
	for i := 0; i < c.N; i++ {
		for _, w := range s.words {
			// Test just a reasonable difference
			for d := 0; d < 5; d++ {
				s.bkTree.FindSimilarWords(w, d)
			}
		}
	}
}
