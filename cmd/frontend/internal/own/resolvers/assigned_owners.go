package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/own"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *ownResolver) computeAssignedOwners(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver, repoID api.RepoID) ([]reasonAndReference, error) {
	assignedOwnership, err := r.ownService().AssignedOwnership(ctx, repoID, api.CommitID(blob.Commit().OID()))
	if err != nil {
		return nil, errors.Wrap(err, "computing assigned ownership")
	}
	var rrs []reasonAndReference
	for _, o := range assignedOwnership.Match(blob.Path()) {
		rrs = append(rrs, reasonAndReference{
			reason: ownershipReason{
				assignedOwnerPath: []string{o.FilePath},
			},
			reference: own.Reference{
				UserID: o.OwnerUserID,
			},
		})
	}
	return rrs, nil
}

func (r *ownResolver) computeAssignedTeams(ctx context.Context, blob *graphqlbackend.GitTreeEntryResolver, repoID api.RepoID) ([]reasonAndReference, error) {
	assignedTeams, err := r.ownService().AssignedTeams(ctx, repoID, api.CommitID(blob.Commit().OID()))
	if err != nil {
		return nil, errors.Wrap(err, "computing assigned ownership")
	}
	var rrs []reasonAndReference
	for _, summary := range assignedTeams.Match(blob.Path()) {
		rrs = append(rrs, reasonAndReference{
			reason: ownershipReason{
				assignedOwnerPath: []string{summary.FilePath},
			},
			reference: own.Reference{
				TeamID: summary.OwnerTeamID,
			},
		})
	}
	return rrs, nil
}

type assignedOwner struct {
	directMatch bool
}

func (a *assignedOwner) Title() (string, error) {
	return "assigned owner", nil
}

func (a *assignedOwner) Description() (string, error) {
	return "Owner is manually assigned.", nil
}

func (a *assignedOwner) IsDirectMatch() bool {
	return a.directMatch
}
