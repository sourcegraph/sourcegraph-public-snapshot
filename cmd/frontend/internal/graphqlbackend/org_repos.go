package graphqlbackend

import (
	"context"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

type orgRepoResolver struct {
	org  *types.Org
	repo *types.OrgRepo
}

func (o *orgRepoResolver) ID() int32 {
	return int32(o.repo.ID)
}

func (o *orgRepoResolver) Org() *orgResolver {
	return &orgResolver{o.org}
}

// DEPRECATED: to be replaced by CanonicalRemoteID.
func (o *orgRepoResolver) RemoteURI() string {
	return string(o.repo.CanonicalRemoteID)
}

func (o *orgRepoResolver) CanonicalRemoteID() string {
	return string(o.repo.CanonicalRemoteID)
}

func (o *orgRepoResolver) CloneURL() string {
	return o.repo.CloneURL
}

func (o *orgRepoResolver) CreatedAt() string {
	return o.repo.CreatedAt.Format(time.RFC3339)
}

func (o *orgRepoResolver) UpdatedAt() string {
	return o.repo.UpdatedAt.Format(time.RFC3339)
}

func (o *orgRepoResolver) Threads(ctx context.Context, args *struct {
	File   *string
	Branch *string
	Limit  *int32
}) *threadConnectionResolver {
	return &threadConnectionResolver{o.org, []*types.OrgRepo{o.repo}, []api.RepoURI{o.repo.CanonicalRemoteID}, args.File, args.Branch, args.Limit}
}

func (o *orgRepoResolver) Repository(ctx context.Context) (*repositoryResolver, error) {
	// See if a repository exists whose URI matches the canonical remote ID.
	repo, err := db.Repos.GetByURI(ctx, api.RepoURI(o.repo.CanonicalRemoteID))
	if errcode.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}
