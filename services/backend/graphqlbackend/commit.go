package graphqlbackend

import (
	"context"

	graphql "github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
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
