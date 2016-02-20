package store

import (
	"fmt"
	"reflect"
)

// scopeRepos returns a list of repos that are matched by the
// filters. If potentially all repos could match, or if enough repos
// could potentially match that it would probably be cheaper to
// iterate through all of them, then a nil slice is returned. If none
// match, an empty slice is returned.
//
// TODO(sqs): return an error if the filters are mutually exclusive?
func scopeRepos(filters []interface{}) ([]string, error) {
	repos := map[string]struct{}{}
	everHadAny := false // whether unitIDs ever contained any repos

	for _, f := range filters {
		switch f := f.(type) {
		case ByReposFilter:
			if len(repos) == 0 && !everHadAny {
				everHadAny = true
				for _, r := range f.ByRepos() {
					repos[r] = struct{}{}
				}
			} else {
				// Intersect.
				newRepos := make(map[string]struct{}, (len(repos)+len(f.ByRepos()))/2)
				for _, r := range f.ByRepos() {
					if _, present := repos[r]; present {
						newRepos[r] = struct{}{}
					}
				}
				repos = newRepos
			}
		}
	}

	if len(repos) == 0 && !everHadAny {
		// No unit scoping filters were present, so scope includes
		// potentially all repos.
		return nil, nil
	}

	repos2 := make([]string, 0, len(repos))
	for repo := range repos {
		repos2 = append(repos2, repo)
	}
	return repos2, nil
}

// A repoStoreOpener opens the RepoStore for the specified repo.
type repoStoreOpener interface {
	openRepoStore(repo string) RepoStore
	openAllRepoStores() (map[string]RepoStore, error)
}

// openRepoStores is a helper func that calls o.openRepoStore for each
// repo returned by scopeRepoStores(filters...).
func openRepoStores(o repoStoreOpener, filters interface{}) (map[string]RepoStore, error) {
	repos, err := scopeRepos(storeFilters(filters))
	if err != nil {
		return nil, err
	}

	if repos == nil {
		return o.openAllRepoStores()
	}

	rss := make(map[string]RepoStore, len(repos))
	for _, repo := range repos {
		rss[repo] = o.openRepoStore(repo)
	}
	return rss, nil
}

// filtersForRepo modifies the filters list to remove filters or
// conditions inside filters that are guaranteed to be true or
// unnecessary when using the filters on a call to a specific repo
// store.
func filtersForRepo(repo string, filters interface{}) interface{} {
	// Copy filters so that it can be used concurrently.
	sf := storeFilters(filters)
	repoFilters := make([]interface{}, len(sf))
	copy(repoFilters, sf)

	d := 0 // deleted (-) and added (+) indexes in repoFilters vs. sf
	for i, f := range sf {
		switch f := f.(type) {

		case byRepoCommitIDsFilter:
			found := false
			for _, v := range f.ByRepoCommitIDs() {
				if v.Repo == repo {
					repoFilters[i+d] = ByCommitIDs(v.CommitID)
					found = true
					break
				}
			}
			if !found {
				panic(fmt.Sprintf("in ByRepoCommitIDsFilter, no version.Repo == %q", repo))
			}

		case byReposFilter:
			found := false
			for _, r := range f.ByRepos() {
				if r == repo {
					repoFilters = append(repoFilters[:i+d], repoFilters[i+d+1:]...)
					found = true
					d--
					break
				}
			}
			if !found {
				panic(fmt.Sprintf("in ByReposFilter, no version.Repo == %q", repo))
			}
		}
	}

	return toTypedFilterSlice(reflect.TypeOf(filters), repoFilters)
}
