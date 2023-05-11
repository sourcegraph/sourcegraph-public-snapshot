package githubapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/repos/webhooks/resolvers"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	ghauth "github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/auth"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
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

func (r *resolver) ConnectWebhookToGitHubApp(ctx context.Context, args *graphqlbackend.ConnectWebhookToGitHubAppArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete GitHub Apps.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	appID, err := UnmarshalGitHubAppID(args.GitHubApp)
	if err != nil {
		return nil, err
	}

	app, err := r.db.GitHubApps().GetByID(ctx, int(appID))
	if err != nil {
		return nil, err
	}

	webhookID, err := resolvers.UnmarshalWebhookID(args.Webhook)
	if err != nil {
		return nil, err
	}

	hook, err := r.db.Webhooks(keyring.Default().WebhookKey).GetByID(ctx, webhookID)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(app.BaseURL)
	if err != nil {
		return nil, err
	}
	apiURL, _ := github.APIRoot(baseURL)

	auther, err := ghauth.NewGitHubAppAuthenticator(int(app.AppID), []byte(app.PrivateKey))
	if err != nil {
		return nil, err
	}

	webhookEndpoint, err := url.Parse(conf.Get().ExternalURL)
	if err != nil {
		return nil, err
	}

	webhookEndpoint.Path = fmt.Sprintf(".api/webhooks/%s", hook.UUID.String())

	type webhookReq struct {
		ContentType string `json:"content_type"`
		InsecureSSL string `json:"insecure_ssl"`
		Secret      string `json:"secret,omitempty"`
		URL         string `json:"url"`
	}
	whReq := webhookReq{
		ContentType: "json",
		InsecureSSL: "0",
		URL:         webhookEndpoint.String(),
	}
	if hook.Secret != nil {
		hookSecret, err := hook.Secret.Decrypt(ctx)
		if err != nil {
			return nil, err
		}

		whReq.Secret = hookSecret
	}
	whBytes, err := json.Marshal(whReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL.JoinPath("/app/hook/config").String(), bytes.NewReader(whBytes))
	if err != nil {
		return nil, err
	}
	if err := auther.Authenticate(req); err != nil {
		return nil, err
	}

	resp, err := httpcli.ExternalClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("webhook update failed: %d", resp.StatusCode))
	}

	hookID := int(hook.ID)
	app.WebhookID = &hookID
	if _, err := r.db.GitHubApps().Update(ctx, app.ID, app); err != nil {
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
	return r.gitHubAppByID(ctx, args.ID)
}

func (r *resolver) GitHubAppByAppID(ctx context.Context, args *graphqlbackend.GitHubAppByAppIDArgs) (graphqlbackend.GitHubAppResolver, error) {
	// 🚨 SECURITY: Check whether user is site-admin
	return r.gitHubAppByAppID(ctx, int(args.AppID), args.BaseURL)
}

func (r *resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		gitHubAppIDKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.gitHubAppByID(ctx, id)
		},
	}
}

func (r *resolver) gitHubAppByID(ctx context.Context, id graphql.ID) (*gitHubAppResolver, error) {
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

func (r *resolver) gitHubAppByAppID(ctx context.Context, appID int, baseURL string) (*gitHubAppResolver, error) {
	// 🚨 SECURITY: Check whether user is site-admin
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	app, err := r.db.GitHubApps().GetByAppID(ctx, appID, baseURL)
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

func (r *gitHubAppResolver) ClientSecret() string {
	return r.app.ClientSecret
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
			InstallID:  int32(*install.ID),
			InstallURL: install.GetHTMLURL(),
			InstallAccount: graphqlbackend.GitHubAppInstallationAccount{
				AccountLogin:     install.Account.GetLogin(),
				AccountAvatarURL: install.Account.GetAvatarURL(),
				AccountURL:       install.Account.GetHTMLURL(),
				AccountType:      install.Account.GetType(),
			},
		})
	}

	return
}

func (r *gitHubAppResolver) Webhook(ctx context.Context) graphqlbackend.WebhookResolver {
	if r.app.WebhookID == nil {
		return nil
	}
	hook, err := r.db.Webhooks(keyring.Default().WebhookKey).GetByID(ctx, int32(*r.app.WebhookID))
	if err != nil {
		return nil
	}
	return resolvers.NewWebhookResolver(r.db, hook)
}
