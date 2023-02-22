package result

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type OwnerMatch struct {
	ResolvedOwner codeowners.ResolvedOwner

	// The following contain information about what search the owner was matched from.
	InputRev *string           `json:"-"`
	Repo     types.MinimalRepo `json:"-"`
	CommitID api.CommitID      `json:"-"`
	Path     string

	LimitHit int
}

func (om *OwnerMatch) RepoName() types.MinimalRepo {
	// todo(own): this might not make sense forever. Right now we derive ownership from files within a repo but if we
	// extend this with external sources then it might not be mandatory to attach an owner to repo.
	// as an alternative we can also conduct a check that nothing expects RepoName to always exist.
	return om.Repo
}

func (om *OwnerMatch) ResultCount() int {
	// just a safeguard
	if om.ResolvedOwner == nil {
		return 0
	}
	return 1
}

func (om *OwnerMatch) Select(filter.SelectPath) Match {
	// There is nothing to "select" from an owner, so we return nil.
	return nil
}

func (om *OwnerMatch) Limit(limit int) int {
	matchCount := om.ResultCount()
	if matchCount == 0 {
		return limit
	}
	return limit - matchCount
}

func (om *OwnerMatch) Key() Key {
	k := Key{
		TypeRank: rankOwnerMatch,
		Repo:     om.Repo.Name,
		Commit:   om.CommitID,
		Path:     om.Path, // TODO: we should decide whether to omit this. If we leave it in then we will get duplicate owners if we match different paths.
	}
	if om.ResolvedOwner != nil {
		k.OwnerMetadata = string(om.ResolvedOwner.Type()) + om.ResolvedOwner.Identifier()
	}
	if om.InputRev != nil {
		k.Rev = *om.InputRev
	}
	return k
}

func (om *OwnerMatch) searchResultMarker() {}
