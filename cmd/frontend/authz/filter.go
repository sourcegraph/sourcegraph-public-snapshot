package authz

// Filterer provides an interface for filtering a list of items keyed by repository name. It should
// be passed to the Filter function.
type Filterer interface {
	// Len returns the length of the list of items to filter.
	Len() int

	// RepoNameForElem returns the repo name associated with the element.
	RepoNameForElem(i int) string

	// Select is invoked when calling Filter to select an element. Elements that are not explicitly
	// selected should be considered "filtered out". It is guaranteed to be called at most once.
	Select(i int)

	// FilterRepoNames maps from a set of repo names to the set of repo names to keep. Note that the
	// returned repo set is allowed to have entries not present in the input repo set. The
	// intersection of the input repo set and the returned repo set is what is selected in the
	// filter.
	FilterRepoNames(repoNames map[string]struct{}) (filteredRepoNames map[string]struct{})
}

// Filter accepts a Filterer and invokes its Select method exactly once for each item in the list
// that should be included.
func Filter(f Filterer) {
	l := f.Len()
	repoNameSet := map[string]struct{}{}
	repoNames := make([]string, l)
	for i := 0; i < l; i++ {
		repoName := f.RepoNameForElem(i)
		repoNames[i] = repoName
		repoNameSet[repoName] = struct{}{}
	}
	filteredRepoNames := f.FilterRepoNames(repoNameSet)
	for i, repoName := range repoNames {
		if _, in := filteredRepoNames[repoName]; in {
			f.Select(i)
		}
	}
}
