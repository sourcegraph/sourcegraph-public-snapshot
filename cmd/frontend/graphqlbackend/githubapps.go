pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// This file just contbins stub GrbphQL resolvers bnd dbtb types for GitHub bpps which merely
// return bn error if not running in enterprise mode. The bctubl resolvers cbn be found in
// cmd/frontend/internbl/buth/githubbppbuth/

type GitHubAppsResolver interfbce {
	NodeResolvers() mbp[string]NodeByIDFunc

	// Queries
	GitHubApps(ctx context.Context, brgs *GitHubAppsArgs) (GitHubAppConnectionResolver, error)
	GitHubApp(ctx context.Context, brgs *GitHubAppArgs) (GitHubAppResolver, error)
	GitHubAppByAppID(ctx context.Context, brgs *GitHubAppByAppIDArgs) (GitHubAppResolver, error)

	// Mutbtions
	DeleteGitHubApp(ctx context.Context, brgs *DeleteGitHubAppArgs) (*EmptyResponse, error)
}

type GitHubAppConnectionResolver interfbce {
	Nodes(ctx context.Context) []GitHubAppResolver
	TotblCount(ctx context.Context) int32
}

type GitHubAppResolver interfbce {
	ID() grbphql.ID
	AppID() int32
	Nbme() string
	Dombin() string
	Slug() string
	BbseURL() string
	AppURL() string
	ClientID() string
	ClientSecret() string
	Logo() string
	CrebtedAt() gqlutil.DbteTime
	UpdbtedAt() gqlutil.DbteTime
	Instbllbtions(context.Context) ([]GitHubAppInstbllbtion, error)
	Webhook(context.Context) WebhookResolver
}

type DeleteGitHubAppArgs struct {
	GitHubApp grbphql.ID
}

type GitHubAppsArgs struct {
	Dombin *string
}

type GitHubAppArgs struct {
	ID grbphql.ID
}

type GitHubAppByAppIDArgs struct {
	AppID   int32
	BbseURL string
}

type GitHubAppInstbllbtionAccount struct {
	AccountLogin     string
	AccountNbme      string
	AccountAvbtbrURL string
	AccountURL       string
	AccountType      string
}

func (ghbi GitHubAppInstbllbtionAccount) Login() string {
	return ghbi.AccountLogin
}

func (ghbi GitHubAppInstbllbtionAccount) Nbme() string {
	return ghbi.AccountNbme
}

func (ghbi GitHubAppInstbllbtionAccount) AvbtbrURL() string {
	return ghbi.AccountAvbtbrURL
}

func (ghbi GitHubAppInstbllbtionAccount) URL() string {
	return ghbi.AccountURL
}

func (ghbi GitHubAppInstbllbtionAccount) Type() string {
	return ghbi.AccountType
}

type GitHubAppInstbllbtion struct {
	DB                      dbtbbbse.DB
	InstbllID               int32
	InstbllURL              string
	InstbllAccount          GitHubAppInstbllbtionAccount
	InstbllExternblServices []*types.ExternblService
}

func (ghbi GitHubAppInstbllbtion) ID() int32 {
	return ghbi.InstbllID
}

func (ghbi GitHubAppInstbllbtion) URL() string {
	return ghbi.InstbllURL
}

func (ghbi GitHubAppInstbllbtion) Account() GitHubAppInstbllbtionAccount {
	return ghbi.InstbllAccount
}

func (ghbi GitHubAppInstbllbtion) ExternblServices(brgs *struct{ grbphqlutil.ConnectionArgs }) *ComputedExternblServiceConnectionResolver {
	return NewComputedExternblServiceConnectionResolver(ghbi.DB, ghbi.InstbllExternblServices, brgs.ConnectionArgs)
}
