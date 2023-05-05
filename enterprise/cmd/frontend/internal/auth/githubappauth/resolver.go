package githubapp

import (
	"context"
	"net/url"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	ghauth "github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
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
		resolvers[i] = NewGitHubAppResolver(r.db, apps[i])
	}

	gitHubAppConnection := &gitHubAppConnectionResolver{
		resolvers:  resolvers,
		totalCount: len(resolvers),
	}

	return gitHubAppConnection, nil
}

func (r *resolver) GitHubApp(ctx context.Context, args *graphqlbackend.GitHubAppArgs) (graphqlbackend.GitHubAppResolver, error) {
	// 🚨 SECURITY: Check whether user is site-admin
	return r.GitHubAppByID(ctx, args.ID)
}

func (r *resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		gitHubAppIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.GitHubAppByID(ctx, id)
		},
	}
}

func (r *resolver) GitHubAppByID(ctx context.Context, id graphql.ID) (*gitHubAppResolver, error) {
	// 🚨 SECURITY: Check whether user is site-admin
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	gitHubAppID, err := UnmarshalGitHubAppID(id)
	if err != nil {
		return nil, err
	}
	app, err := r.db.GitHubApps().GetByID(ctx, int(gitHubAppID))
	if err != nil {
		return nil, err
	}

	return &gitHubAppResolver{
		app: app,
		db:  r.db,
	}, nil
}

// NewGitHubAppResolver creates a new GitHubAppResolver from a GitHubApp.
func NewGitHubAppResolver(db edb.EnterpriseDB, app *types.GitHubApp) *gitHubAppResolver {
	return &gitHubAppResolver{app: app, db: db}
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
	db  edb.EnterpriseDB
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

func (r *gitHubAppResolver) ExternalServices(ctx context.Context, args *struct{ graphqlutil.ConnectionArgs }) *graphqlbackend.ComputedExternalServiceConnectionResolver {
	extsvcs, err := r.db.ExternalServices().List(ctx, database.ExternalServicesListOptions{
		Kinds: []string{extsvc.KindGitHub},
	})
	if err != nil {
		return nil
	}

	var filteredExtsvc []*itypes.ExternalService
	for _, es := range extsvcs {
		parsed, err := extsvc.ParseEncryptableConfig(ctx, extsvc.KindGitHub, es.Config)
		if err != nil {
			continue
		}
		c := parsed.(*schema.GitHubConnection)
		if c.GitHubAppDetails == nil || c.GitHubAppDetails.AppID != r.app.AppID || c.Url != r.app.BaseURL {
			continue
		}
		filteredExtsvc = append(filteredExtsvc, es)
	}

	return graphqlbackend.NewComputedExternalServiceConnectionResolver(r.db, filteredExtsvc, args.ConnectionArgs)
}

func (r *gitHubAppResolver) Installations(ctx context.Context) (installations []graphqlbackend.GitHubAppInstallation) {
	auther, err := ghauth.NewGitHubAppAuthenticator(int(r.AppID()), []byte(r.app.PrivateKey))
	if err != nil {
		return nil
	}

	baseURL, err := url.Parse(r.app.BaseURL)
	if err != nil {
		return nil
	}
	apiURL, _ := github.APIRoot(baseURL)

	cli := github.NewV3Client(log.Scoped("GitHubAppResolver", ""), "", apiURL, auther, nil)
	installs, err := cli.GetAppInstallations(ctx)
	if err != nil {
		return nil
	}

	for _, install := range installs {
		installations = append(installations, graphqlbackend.GitHubAppInstallation{
			InstallID: int32(*install.ID),
			InstallAccount: graphqlbackend.GitHubAppInstallationAccount{
				AccountLogin:     install.Account.GetLogin(),
				AccountAvatarURL: install.Account.GetAvatarURL(),
				AccountURL:       install.Account.GetURL(),
			},
		})
	}

	return
}
