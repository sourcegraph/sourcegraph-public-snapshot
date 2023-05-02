package githubapp

import (
	"context"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/auth"
)

// NewResolver returns a new Resolver that uses the given database
func NewResolver(logger log.Logger, db edb.EnterpriseDB) graphqlbackend.GitHubAppsResolver {
	return &resolver{logger: logger, db: db}
}

type resolver struct {
	logger log.Logger
	db     edb.EnterpriseDB
}

// DeleteGitHubApp deletes a GitHub App along with all of its associated
// code host connections and auth providers.
func (r *resolver) DeleteGitHubApp(ctx context.Context, args *graphqlbackend.DeleteGitHubAppArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete GitHub Apps.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var appID int
	if err := args.GitHubApp.UnmarshalGraphQL(&appID); err != nil {
		return nil, err
	}

	return nil, r.db.GitHubApps().Delete(ctx, appID)
}

type gitHubAppResolver struct {
	// cache results because they are used by multiple fields
	once       sync.Once
	gitHubApps []*types.GitHubApp
	err        error
	db         edb.EnterpriseDB
}

func (r *resolver) GitHubApps(ctx context.Context, args *graphqlbackend.GitHubAppsArgs) (*gitHubAppResolver, error) {
	// 🚨 SECURITY: Check whether user is site-admin
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &gitHubAppResolver{db: r.db}, nil
}

func (r *gitHubAppResolver) compute(ctx context.Context) ([]*types.GitHubApp, error) {
	r.once.Do(func() {
		r.gitHubApps, r.err = r.db.GitHubApps().List(ctx)
	})
	return r.gitHubApps, r.err
}
