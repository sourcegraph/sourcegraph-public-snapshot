package githubapp

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
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

// MarshalGitHubAppID converts a GitHub App ID (database ID) to a GraphQL ID.
func MarshalGitHubAppID(id int64) graphql.ID {
	return relay.MarshalID(gitHubAppIDKind, id)
}

// UnmarshalGitHubAppID converts a GitHub App GraphQL ID into a database ID.
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

func (r *resolver) GitHubApps(ctx context.Context) (graphqlbackend.GitHubAppConnectionResolver, error) {
	// 🚨 SECURITY: Check whether user is site-admin
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	apps, err := r.db.GitHubApps().List(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.GitHubAppResolver, len(apps))
	for i := range apps {
		resolvers[i] = NewGitHubAppResolver(apps[i])
	}

	gitHubAppConnection := &gitHubAppConnectionResolver{
		resolvers:  resolvers,
		totalCount: len(resolvers),
	}

	return gitHubAppConnection, nil
}

// NewGitHubAppResolver creates a new GitHubAppResolver from a GitHubApp.
func NewGitHubAppResolver(app *types.GitHubApp) *gitHubAppResolver {
	return &gitHubAppResolver{app: app}
}

type gitHubAppConnectionResolver struct {
	resolvers  []graphqlbackend.GitHubAppResolver
	totalCount int
}

func (r *gitHubAppConnectionResolver) Nodes(ctx context.Context) []graphqlbackend.GitHubAppResolver {
	return r.resolvers
}

func (r *gitHubAppConnectionResolver) TotalCount(ctx context.Context) int32 {
	return int32(r.totalCount)
}

// gitHubAppResolver is a GraphQL node resolver for GitHubApps.
type gitHubAppResolver struct {
	app *types.GitHubApp
}

func (r *gitHubAppResolver) ID() graphql.ID {
	return MarshalGitHubAppID(int64(r.app.ID))
}

func (r *gitHubAppResolver) AppID() int32 {
	return int32(r.app.AppID)
}

func (r *gitHubAppResolver) Name() string {
	return r.app.Name
}

func (r *gitHubAppResolver) Slug() string {
	return r.app.Slug
}

func (r *gitHubAppResolver) BaseURL() string {
	return r.app.BaseURL
}

func (r *gitHubAppResolver) AppURL() string {
	return r.app.AppURL
}

func (r *gitHubAppResolver) ClientID() string {
	return r.app.ClientID
}

func (r *gitHubAppResolver) Logo() string {
	return r.app.Logo
}

func (r *gitHubAppResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.app.CreatedAt}
}

func (r *gitHubAppResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.app.UpdatedAt}
}
