package result

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type OwnerMatch struct {
	Owner *codeownerspb.Owner

	// The following contain information about what search the owner was matched from.
	InputRev *string           `json:"-"`
	Repo     types.MinimalRepo `json:"-"`
	CommitID api.CommitID      `json:"-"`
	Path     string

	LimitHit int

	// Debug is optionally set with a debug message explaining the result.
	//
	// Note: this is a pointer since usually this is unset. Pointer is 8 bytes
	// vs an empty string which is 16 bytes.
	Debug *string `json:"-"`
}

func (om *OwnerMatch) RepoName() types.MinimalRepo {
	// todo(own): this might not make sense forever. right now we derive ownership from files within a repo but if we
	// extend this with external sources then it might not be mandatory to attach an owner to repo.
	// as an alternative we can also conduct a check that nothing expects RepoName to always exist.
	return om.Repo
}

func (om *OwnerMatch) ResultCount() int {
	// just a safeguard
	if om.Owner == nil {
		return 0
	}
	return 1
}

func (om *OwnerMatch) Select(selectPath filter.SelectPath) Match {
	// todo can this be safely ignored?
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
		Path:     om.Path,
	}
	if om.Owner != nil {
		k.Metadata = om.Owner.Handle + om.Owner.Email
	}
	if om.InputRev != nil {
		k.Rev = *om.InputRev
	}
	return k
}

func (om *OwnerMatch) searchResultMarker() {}
