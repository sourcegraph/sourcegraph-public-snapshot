package sourcegraph

// Empty is whether there are no search results for any result type.
func (r *SearchResults) Empty() bool {
	return len(r.Defs) == 0 && len(r.People) == 0 && len(r.Repos) == 0 && len(r.Tree) == 0
}
