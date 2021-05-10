package graphqlbackend

import "github.com/sourcegraph/sourcegraph/internal/search/result"

type searchResultDeduper struct {
	seenFileMatches   map[result.Key]*FileMatchResolver
	seenRepoMatches   map[result.Key]*RepositoryResolver
	seenCommitMatches map[result.Key]*CommitSearchResultResolver
	seenDiffMatches   map[result.Key]*CommitSearchResultResolver
}

func NewDeduper() *searchResultDeduper {
	return &searchResultDeduper{
		seenFileMatches:   make(map[result.Key]*FileMatchResolver),
		seenRepoMatches:   make(map[result.Key]*RepositoryResolver),
		seenCommitMatches: make(map[result.Key]*CommitSearchResultResolver),
		seenDiffMatches:   make(map[result.Key]*CommitSearchResultResolver),
	}
}

// Add adds a SearchResultResolver to the deduper, merging it into
// a previously added SearchResultResolver if the result key has already been seen
func (d *searchResultDeduper) Add(r SearchResultResolver) {
	if fileMatch, ok := r.ToFileMatch(); ok {
		if prev, seen := d.seenFileMatches[fileMatch.FileMatch.Key()]; seen {
			prev.appendMatches(fileMatch)
			return
		}
		d.seenFileMatches[fileMatch.FileMatch.Key()] = fileMatch
		return
	}

	if repoMatch, ok := r.ToRepository(); ok {
		if _, seen := d.seenRepoMatches[repoMatch.Key()]; seen {
			return
		}
		d.seenRepoMatches[repoMatch.Key()] = repoMatch
		return
	}

	if commitMatch, ok := r.ToCommitSearchResult(); ok {
		if commitMatch.DiffPreview() != nil {
			if _, seen := d.seenDiffMatches[commitMatch.Key()]; seen {
				return
			}
			d.seenDiffMatches[commitMatch.Key()] = commitMatch
			return
		}

		if _, seen := d.seenCommitMatches[commitMatch.Key()]; seen {
			return
		}
		d.seenCommitMatches[commitMatch.Key()] = commitMatch
		return
	}
}

// Seen returns whether the given result key exists for a file type in the deduper without
// modifying the contents of the deduper
func (d *searchResultDeduper) Seen(r SearchResultResolver) (ok bool) {
	switch v := r.(type) {
	case *FileMatchResolver:
		_, ok = d.seenFileMatches[v.FileMatch.Key()]
	case *RepositoryResolver:
		_, ok = d.seenRepoMatches[v.Key()]
	case *CommitSearchResultResolver:
		if v.DiffPreview() != nil {
			_, ok = d.seenDiffMatches[v.Key()]
		} else {
			_, ok = d.seenCommitMatches[v.Key()]
		}
	}

	return ok
}

// Results returns a slice of SearchResultResolvers, deduplicated from
// the SearchResultResolvers that were added with Add
func (d *searchResultDeduper) Results() []SearchResultResolver {
	total := len(d.seenFileMatches) + len(d.seenRepoMatches) + len(d.seenCommitMatches) + len(d.seenDiffMatches)
	r := make([]SearchResultResolver, 0, total)
	for _, v := range d.seenFileMatches {
		r = append(r, v)
	}
	for _, v := range d.seenRepoMatches {
		r = append(r, v)
	}
	for _, v := range d.seenCommitMatches {
		r = append(r, v)
	}
	for _, v := range d.seenDiffMatches {
		r = append(r, v)
	}
	return r
}
