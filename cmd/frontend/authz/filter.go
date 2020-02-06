package authz

type Filterer interface {
	Len() int
	RepoNameForElem(i int) string
	Select(i int) // guaranteed to be called at most once

	// Returns the set of repo names to keep. Does not have to be a subset of repoNames argument.
	FilterRepoNames(repoNames map[string]struct{}) (filteredRepoNames map[string]struct{})
}

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
