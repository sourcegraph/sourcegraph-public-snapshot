package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// This file just contains stub GraphQL resolvers and data types for GitHub apps which merely
// return an error if not running in enterprise mode. The actual resolvers can be found in
// cmd/frontend/internal/auth/githubappauth/

type GitHubAppsResolver interface {
	NodeResolvers() map[string]NodeByIDFunc

	// Queries
	GitHubApps(ctx context.Context, args *GitHubAppsArgs) (GitHubAppConnectionResolver, error)
	GitHubApp(ctx context.Context, args *GitHubAppArgs) (GitHubAppResolver, error)
	GitHubAppByAppID(ctx context.Context, args *GitHubAppByAppIDArgs) (GitHubAppResolver, error)

	// Mutations
	DeleteGitHubApp(ctx context.Context, args *DeleteGitHubAppArgs) (*EmptyResponse, error)
}

type GitHubAppConnectionResolver interface {
	Nodes(ctx context.Context) []GitHubAppResolver
	TotalCount(ctx context.Context) int32
}

type GitHubAppResolver interface {
	ID() graphql.ID
	AppID() int32
	Name() string
	Domain() string
	Slug() string
	BaseURL() string
	AppURL() string
	ClientID() string
	ClientSecret() string
	Logo() string
	CreatedAt() gqlutil.DateTime
	UpdatedAt() gqlutil.DateTime
	Installations(context.Context) ([]GitHubAppInstallation, error)
	Webhook(context.Context) WebhookResolver
}

type DeleteGitHubAppArgs struct {
	GitHubApp graphql.ID
}

type GitHubAppsArgs struct {
	Domain *string
}

type GitHubAppArgs struct {
	ID graphql.ID
}

type GitHubAppByAppIDArgs struct {
	AppID   int32
	BaseURL string
}

type GitHubAppInstallationAccount struct {
	AccountLogin     string
	AccountName      string
	AccountAvatarURL string
	AccountURL       string
	AccountType      string
}

func (ghai GitHubAppInstallationAccount) Login() string {
	return ghai.AccountLogin
}

func (ghai GitHubAppInstallationAccount) Name() string {
	return ghai.AccountName
}

func (ghai GitHubAppInstallationAccount) AvatarURL() string {
	return ghai.AccountAvatarURL
}

func (ghai GitHubAppInstallationAccount) URL() string {
	return ghai.AccountURL
}

func (ghai GitHubAppInstallationAccount) Type() string {
	return ghai.AccountType
}

type GitHubAppInstallation struct {
	DB                      database.DB
	InstallID               int32
	InstallURL              string
	InstallAccount          GitHubAppInstallationAccount
	InstallExternalServices []*types.ExternalService
}

func (ghai GitHubAppInstallation) ID() int32 {
	return ghai.InstallID
}

func (ghai GitHubAppInstallation) URL() string {
	return ghai.InstallURL
}

func (ghai GitHubAppInstallation) Account() GitHubAppInstallationAccount {
	return ghai.InstallAccount
}

func (ghai GitHubAppInstallation) ExternalServices(args *struct{ graphqlutil.ConnectionArgs }) *ComputedExternalServiceConnectionResolver {
	return NewComputedExternalServiceConnectionResolver(ghai.DB, ghai.InstallExternalServices, args.ConnectionArgs)
}
