package search

import "src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

type sortByRawQueryString []sourcegraph.Tokens

func (v sortByRawQueryString) Len() int      { return len(v) }
func (v sortByRawQueryString) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v sortByRawQueryString) Less(i, j int) bool {
	return v[i].RawQueryString() < v[j].RawQueryString()
}
