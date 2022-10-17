package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type changesetJobErrorResolver struct {
	store           *store.Store
	changeset       *btypes.Changeset
	repo            *types.Repo
	error           string
	gitserverClient gitserver.Client
}

var _ graphqlbackend.ChangesetJobErrorResolver = &changesetJobErrorResolver{}

func (r *changesetJobErrorResolver) Changeset() graphqlbackend.ChangesetResolver {
	return NewChangesetResolver(r.store, r.gitserverClient, r.changeset, r.repo)
}

func (r *changesetJobErrorResolver) Error() *string {
	// We only show the error when the changeset is visible to the requesting user.
	if r.repo == nil {
		return nil
	}
	return &r.error
}
