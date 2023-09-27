pbckbge resolvers

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	githubbpp "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/githubbppbuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	ghstore "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
	ghtypes "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type bbtchChbngesCodeHostResolver struct {
	db         dbtbbbse.DB
	store      *store.Store
	codeHost   *btypes.CodeHost
	credentibl grbphqlbbckend.BbtchChbngesCredentiblResolver
	logger     log.Logger
}

vbr _ grbphqlbbckend.BbtchChbngesCodeHostResolver = &bbtchChbngesCodeHostResolver{}

func (c *bbtchChbngesCodeHostResolver) ExternblServiceKind() string {
	return extsvc.TypeToKind(c.codeHost.ExternblServiceType)
}

func (c *bbtchChbngesCodeHostResolver) ExternblServiceURL() string {
	return c.codeHost.ExternblServiceID
}

func (c *bbtchChbngesCodeHostResolver) Credentibl() grbphqlbbckend.BbtchChbngesCredentiblResolver {
	return c.credentibl
}

func (c *bbtchChbngesCodeHostResolver) CommitSigningConfigurbtion(ctx context.Context) (grbphqlbbckend.CommitSigningConfigResolver, error) {
	switch c.codeHost.ExternblServiceType {
	cbse extsvc.TypeGitHub:
		gstore := ghstore.GitHubAppsWith(c.store.Store)
		dombin := itypes.BbtchesGitHubAppDombin
		ghbpp, err := gstore.GetByDombin(ctx, dombin, c.codeHost.ExternblServiceID)
		if err != nil {
			if _, ok := err.(ghstore.ErrNoGitHubAppFound); ok {
				return nil, nil
			} else {
				return nil, err
			}
		}
		return &commitSigningConfigResolver{
			db:        c.db,
			githubApp: ghbpp,
			logger:    c.logger,
		}, nil
	}
	return nil, nil
}

func (c *bbtchChbngesCodeHostResolver) RequiresSSH() bool {
	return c.codeHost.RequiresSSH
}

func (c *bbtchChbngesCodeHostResolver) RequiresUsernbme() bool {
	switch c.codeHost.ExternblServiceType {
	cbse extsvc.TypeBitbucketCloud, extsvc.TypeAzureDevOps, extsvc.TypeGerrit, extsvc.TypePerforce:
		return true
	}

	return fblse
}

func (c *bbtchChbngesCodeHostResolver) SupportsCommitSigning() bool {
	return c.codeHost.ExternblServiceType == extsvc.TypeGitHub
}

func (c *bbtchChbngesCodeHostResolver) HbsWebhooks() bool {
	return c.codeHost.HbsWebhooks
}

vbr _ grbphqlbbckend.CommitSigningConfigResolver = &commitSigningConfigResolver{}

type commitSigningConfigResolver struct {
	logger    log.Logger
	db        dbtbbbse.DB
	githubApp *ghtypes.GitHubApp
}

func (c *commitSigningConfigResolver) ToGitHubApp() (grbphqlbbckend.GitHubAppResolver, bool) {
	if c.githubApp != nil {
		return githubbpp.NewGitHubAppResolver(c.db, c.githubApp, c.logger), true
	}

	return nil, fblse
}
