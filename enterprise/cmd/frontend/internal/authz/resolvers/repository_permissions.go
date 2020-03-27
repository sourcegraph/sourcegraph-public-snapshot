package resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

var _ graphqlbackend.RepositoryPermissionsConnectionResolver = (*repositoryPermissionsConnectionResolver)(nil)

type repositoryPermissionsConnectionResolver struct {
	perms      []*authz.RepoPermissions
	totalCount int
	limit      int

	// cache results because they are used by multiple fields
	once      sync.Once
	repoPerms []graphqlbackend.RepositoryPermissionsResolver
	pageInfo  *graphqlutil.PageInfo
	err       error
}

var _ graphqlbackend.RepositoryPermissionsResolver = (*repositoryPermissionsResolver)(nil)

type repositoryPermissionsResolver struct {
	repo  *graphqlbackend.RepositoryResolver
	perms graphqlbackend.RepositoryPermissionsInfoResolver
}

func (r *repositoryPermissionsResolver) Repository() *graphqlbackend.RepositoryResolver {
	return r.repo
}

func (r *repositoryPermissionsResolver) Permissions() graphqlbackend.RepositoryPermissionsInfoResolver {
	return r.perms
}

var _ graphqlbackend.RepositoryPermissionsInfoResolver = (*repositoryPermissionsInfoResolver)(nil)

type repositoryPermissionsInfoResolver struct {
	perms *authz.RepoPermissions
}

func (r *repositoryPermissionsInfoResolver) Perm() string {
	return strings.ToUpper(r.perms.Perm.String())
}

func (r *repositoryPermissionsInfoResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.perms.UpdatedAt}
}

// 🚨 SECURITY: It is the caller's responsibility to ensure the current authenticated user
// is the site admin because this method computes data from all available information in
// the database.
func (r *repositoryPermissionsConnectionResolver) compute(ctx context.Context) ([]graphqlbackend.RepositoryPermissionsResolver, *graphqlutil.PageInfo, error) {
	r.once.Do(func() {
		repoIDs := make([]api.RepoID, 0, len(r.perms))
		for _, p := range r.perms {
			repoIDs = append(repoIDs, api.RepoID(p.RepoID))
		}

		var repos []*types.Repo
		repos, r.err = db.Repos.GetByIDs(ctx, repoIDs...)
		if r.err != nil {
			return
		}

		reposSet := make(map[api.RepoID]*types.Repo, len(repos))
		for _, r := range repos {
			reposSet[r.ID] = r
		}

		r.repoPerms = make([]graphqlbackend.RepositoryPermissionsResolver, 0, len(r.perms))
		for _, p := range r.perms {
			r.repoPerms = append(r.repoPerms, &repositoryPermissionsResolver{
				repo:  graphqlbackend.NewRepositoryResolver(reposSet[api.RepoID(p.RepoID)]),
				perms: &repositoryPermissionsInfoResolver{perms: p},
			})
		}

		r.pageInfo = graphqlutil.HasNextPage(r.limit > 0 && r.totalCount > r.limit)
	})
	return r.repoPerms, r.pageInfo, r.err
}

func (r *repositoryPermissionsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.RepositoryPermissionsResolver, error) {
	// 🚨 SECURITY: Only site admins may access this method.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	repoPerms, _, err := r.compute(ctx)
	return repoPerms, err
}

func (r *repositoryPermissionsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins may access this method.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return 0, err
	}

	return int32(r.totalCount), nil
}

func (r *repositoryPermissionsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	// 🚨 SECURITY: Only site admins may access this method.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	_, pageInfo, err := r.compute(ctx)
	return pageInfo, err
}
