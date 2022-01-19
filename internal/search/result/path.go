package result

import (
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// PathMatch represents a match of the file itself, rather than
// a match on the content of the file.
type PathMatch struct {
	File

	// MatchedPathRanges is a set of ranges of the embedded File.Path
	// that matched the query.
	MatchedPathRanges Ranges
}

func (p *PathMatch) ResultCount() int {
	return 1
}

// Limit will mutate p such that it only has limit results. limit is a number
// greater than 0.
func (p *PathMatch) Limit(limit int) int {
	return limit - 1
}

func (p *PathMatch) Key() Key {
	k := Key{
		TypeRank: rankPathMatch,
		Repo:     p.Repo.Name,
		Commit:   p.CommitID,
		Path:     p.Path,
	}
	if p.InputRev != nil {
		k.Rev = *p.InputRev
	}
	return k
}

func (p *PathMatch) Select(selectPath filter.SelectPath) Match {
	switch selectPath.Root() {
	case filter.Repository:
		return &RepoMatch{
			Name: p.Repo.Name,
			ID:   p.Repo.ID,
		}
	case filter.File:
		return p
	default:
		return nil
	}
}

func (p *PathMatch) RepoName() types.MinimalRepo {
	return p.Repo
}

func (p *PathMatch) searchResultMarker() {}
