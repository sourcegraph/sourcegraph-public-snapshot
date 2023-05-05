package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// This file just contains stub GraphQL resolvers and data types for GitHub apps which merely
// return an error if not running in enterprise mode. The actual resolvers can be found in
// enterprise/cmd/frontend/internal/auth/githubappauth/

type GitHubAppsResolver interface {
	NodeResolvers() map[string]NodeByIDFunc

	// Queries
	GitHubApps(ctx context.Context) (GitHubAppConnectionResolver, error)
	GitHubApp(ctx context.Context, args *GitHubAppArgs) (GitHubAppResolver, error)

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
	Slug() string
	BaseURL() string
	AppURL() string
	ClientID() string
	Logo() string
	CreatedAt() gqlutil.DateTime
	UpdatedAt() gqlutil.DateTime
	ExternalServices(context.Context, *struct{ graphqlutil.ConnectionArgs }) *ComputedExternalServiceConnectionResolver
	Installations(context.Context) []GitHubAppInstallation
}

type DeleteGitHubAppArgs struct {
	GitHubApp graphql.ID
}

type GitHubAppsArgs struct {
	graphqlutil.ConnectionArgs
	After     *string
	Namespace *graphql.ID
}

type GitHubAppArgs struct {
	ID graphql.ID
}

type GitHubAppInstallationAccount struct {
	AccountLogin     string
	AccountName      string
	AccountAvatarURL string
	AccountURL       string
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

type GitHubAppInstallation struct {
	InstallID      int32
	InstallAccount GitHubAppInstallationAccount
}

func (ghai GitHubAppInstallation) ID() int32 {
	return ghai.InstallID
}

func (ghai GitHubAppInstallation) Account() GitHubAppInstallationAccount {
	return ghai.InstallAccount
}
