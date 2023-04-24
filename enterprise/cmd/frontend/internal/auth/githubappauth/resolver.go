package githubapp

import (
	"context"

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

// CreateGitHubApp creates a GitHub App.
func (r *resolver) CreateGitHubApp(ctx context.Context, args *graphqlbackend.CreateGitHubAppArgs) (*int32, error) {

	// 🚨 SECURITY: Only site admins can create GitHub Apps.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	// only set the fields we need, we can add Slug, Name and Logo URL once we have an installation access token
	input := args.Input
	app := &types.GitHubApp{
		AppID:        int(input.AppID),
		BaseURL:      input.BaseURL,
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		PrivateKey:   input.PrivateKey,
	}
	app, err := r.db.GitHubApps().Create(ctx, app)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, nil
	}
	id := int32(app.ID)
	return &id, nil
}
