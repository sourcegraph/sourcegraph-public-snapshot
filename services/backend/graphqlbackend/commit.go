package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

type commitSpec struct {
	RepoID   int32
	CommitID string
}

type commitResolver struct {
	nodeBase
	commit commitSpec
}

func commitByID(ctx context.Context, id graphql.ID) (nodeResolver, error) {
	var commit commitSpec
	if err := relay.UnmarshalSpec(id, &commit); err != nil {
		return nil, err
	}
	return &commitResolver{commit: commit}, nil
}

func (r *commitResolver) ToCommit() (*commitResolver, bool) {
	return r, true
}

func (r *commitResolver) ID() graphql.ID {
	return relay.MarshalID("Commit", r.commit)
}

func (r *commitResolver) SHA1() string {
	return r.commit.CommitID
}

func (r *commitResolver) Tree(ctx context.Context, args *struct {
	Path      string
	Recursive bool
}) (*treeResolver, error) {
	return makeTreeResolver(ctx, r.commit, args.Path, args.Recursive)
}

func (r *commitResolver) Languages(ctx context.Context) ([]string, error) {
	inventory, err := backend.Repos.GetInventory(ctx, &sourcegraph.RepoRevSpec{
		Repo:     r.commit.RepoID,
		CommitID: r.commit.CommitID,
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, len(inventory.Languages))
	for i, l := range inventory.Languages {
		names[i] = l.Name
	}
	return names, nil
}
