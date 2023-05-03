package githubapp

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewResolver returns a new Resolver that uses the given database
func NewResolver(logger log.Logger, db edb.EnterpriseDB) graphqlbackend.GitHubAppsResolver {
	return &resolver{logger: logger, db: db}
}

type resolver struct {
	logger log.Logger
	db     edb.EnterpriseDB
}

const gitHubAppIDKind = "GitHubApp"

func MarshalGitHubAppID(id int64) graphql.ID {
	return relay.MarshalID(gitHubAppIDKind, id)
}

func UnmarshalGitHubAppID(id graphql.ID) (gitHubAppID int64, err error) {
	if kind := relay.UnmarshalKind(id); kind != gitHubAppIDKind {
		err = errors.Errorf("expected graph ID to have kind %q; got %q", gitHubAppIDKind, kind)
		return
	}

	err = relay.UnmarshalSpec(id, &gitHubAppID)
	return
}

// DeleteGitHubApp deletes a GitHub App along with all of its associated
// code host connections and auth providers.
func (r *resolver) DeleteGitHubApp(ctx context.Context, args *graphqlbackend.DeleteGitHubAppArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete GitHub Apps.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	appID, err := UnmarshalGitHubAppID(args.GitHubApp)
	if err != nil {
		return nil, err
	}

	if err := r.db.GitHubApps().Delete(ctx, int(appID)); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
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
