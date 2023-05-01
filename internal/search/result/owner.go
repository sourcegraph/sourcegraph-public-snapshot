package result

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type Owner interface {
	Type() string
	Identifier() string
}

type OwnerPerson struct {
	Handle string
	Email  string
	User   *types.User
}

func (o OwnerPerson) Identifier() string {
	return "Person:" + o.Handle + o.Email
}

func (o OwnerPerson) Type() string {
	return "person"
}

type OwnerTeam struct {
	Handle string
	Email  string
	Team   *types.Team
}

func (o OwnerTeam) Identifier() string {
	return "Team:" + o.Team.Name
}

func (o OwnerTeam) Type() string {
	return "team"
}

type OwnerMatch struct {
	ResolvedOwner Owner

	// The following contain information about what search the owner was matched from.
	InputRev *string           `json:"-"`
	Repo     types.MinimalRepo `json:"-"`
	CommitID api.CommitID      `json:"-"`

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
	}
	if om.ResolvedOwner != nil {
		k.OwnerMetadata = om.ResolvedOwner.Type() + om.ResolvedOwner.Identifier()
	}
	if om.InputRev != nil {
		k.Rev = *om.InputRev
	}
	return k
}

func (om *OwnerMatch) searchResultMarker() {}
