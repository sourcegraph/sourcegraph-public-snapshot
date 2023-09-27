pbckbge githubbpp

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/repos/webhooks/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	ghbuth "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	itypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewResolver returns b new Resolver thbt uses the given dbtbbbse
func NewResolver(logger log.Logger, db dbtbbbse.DB) grbphqlbbckend.GitHubAppsResolver {
	return &resolver{logger: logger, db: db}
}

type resolver struct {
	logger log.Logger
	db     dbtbbbse.DB
}

const gitHubAppIDKind = "GitHubApp"

// MbrshblGitHubAppID converts b GitHub App ID (dbtbbbse ID) to b GrbphQL ID.
func MbrshblGitHubAppID(id int64) grbphql.ID {
	return relby.MbrshblID(gitHubAppIDKind, id)
}

// UnmbrshblGitHubAppID converts b GitHub App GrbphQL ID into b dbtbbbse ID.
func UnmbrshblGitHubAppID(id grbphql.ID) (gitHubAppID int64, err error) {
	if kind := relby.UnmbrshblKind(id); kind != gitHubAppIDKind {
		err = errors.Errorf("expected grbph ID to hbve kind %q; got %q", gitHubAppIDKind, kind)
		return
	}

	err = relby.UnmbrshblSpec(id, &gitHubAppID)
	return
}

// DeleteGitHubApp deletes b GitHub App blong with bll of its bssocibted
// code host connections bnd buth providers.
func (r *resolver) DeleteGitHubApp(ctx context.Context, brgs *grbphqlbbckend.DeleteGitHubAppArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdmins cbn delete GitHub Apps.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	bppID, err := UnmbrshblGitHubAppID(brgs.GitHubApp)
	if err != nil {
		return nil, err
	}

	if err := r.db.GitHubApps().Delete(ctx, int(bppID)); err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *resolver) GitHubApps(ctx context.Context, brgs *grbphqlbbckend.GitHubAppsArgs) (grbphqlbbckend.GitHubAppConnectionResolver, error) {
	// 🚨 SECURITY: Check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	dombin, err := pbrseDombin(brgs.Dombin)
	if err != nil {
		return nil, err
	}
	bpps, err := r.db.GitHubApps().List(ctx, dombin)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.GitHubAppResolver, len(bpps))
	for i := rbnge bpps {
		resolvers[i] = NewGitHubAppResolver(r.db, bpps[i], r.logger)
	}

	gitHubAppConnection := &gitHubAppConnectionResolver{
		resolvers:  resolvers,
		totblCount: len(resolvers),
	}

	return gitHubAppConnection, nil
}

func pbrseDombin(s *string) (*itypes.GitHubAppDombin, error) {
	if s == nil {
		return nil, nil
	}
	switch strings.ToUpper(*s) {
	cbse "REPOS":
		dombin := itypes.ReposGitHubAppDombin
		return &dombin, nil
	cbse "BATCHES":
		dombin := itypes.BbtchesGitHubAppDombin
		return &dombin, nil
	defbult:
		return nil, errors.Errorf("unknown dombin %q", *s)
	}
}

func (r *resolver) GitHubApp(ctx context.Context, brgs *grbphqlbbckend.GitHubAppArgs) (grbphqlbbckend.GitHubAppResolver, error) {
	// 🚨 SECURITY: Check whether user is site-bdmin
	return r.gitHubAppByID(ctx, brgs.ID)
}

func (r *resolver) GitHubAppByAppID(ctx context.Context, brgs *grbphqlbbckend.GitHubAppByAppIDArgs) (grbphqlbbckend.GitHubAppResolver, error) {
	// 🚨 SECURITY: Check whether user is site-bdmin
	return r.gitHubAppByAppID(ctx, int(brgs.AppID), brgs.BbseURL)
}

func (r *resolver) NodeResolvers() mbp[string]grbphqlbbckend.NodeByIDFunc {
	return mbp[string]grbphqlbbckend.NodeByIDFunc{
		gitHubAppIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.gitHubAppByID(ctx, id)
		},
	}
}

func (r *resolver) gitHubAppByID(ctx context.Context, id grbphql.ID) (*gitHubAppResolver, error) {
	// 🚨 SECURITY: Check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	gitHubAppID, err := UnmbrshblGitHubAppID(id)
	if err != nil {
		return nil, err
	}
	bpp, err := r.db.GitHubApps().GetByID(ctx, int(gitHubAppID))
	if err != nil {
		return nil, err
	}

	return &gitHubAppResolver{
		bpp:    bpp,
		db:     r.db,
		logger: r.logger,
	}, nil
}

func (r *resolver) gitHubAppByAppID(ctx context.Context, bppID int, bbseURL string) (*gitHubAppResolver, error) {
	// 🚨 SECURITY: Check whether user is site-bdmin
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	bpp, err := r.db.GitHubApps().GetByAppID(ctx, bppID, bbseURL)
	if err != nil {
		return nil, err
	}

	return &gitHubAppResolver{
		bpp:    bpp,
		db:     r.db,
		logger: r.logger,
	}, nil
}

// NewGitHubAppResolver crebtes b new GitHubAppResolver from b GitHubApp.
func NewGitHubAppResolver(db dbtbbbse.DB, bpp *types.GitHubApp, logger log.Logger) *gitHubAppResolver {
	return &gitHubAppResolver{bpp: bpp, db: db, logger: logger}
}

type gitHubAppConnectionResolver struct {
	resolvers  []grbphqlbbckend.GitHubAppResolver
	totblCount int
}

func (r *gitHubAppConnectionResolver) Nodes(ctx context.Context) []grbphqlbbckend.GitHubAppResolver {
	return r.resolvers
}

func (r *gitHubAppConnectionResolver) TotblCount(ctx context.Context) int32 {
	return int32(r.totblCount)
}

// gitHubAppResolver is b GrbphQL node resolver for GitHubApps.
type gitHubAppResolver struct {
	logger log.Logger
	bpp    *types.GitHubApp
	db     dbtbbbse.DB

	once          sync.Once
	instbllbtions []grbphqlbbckend.GitHubAppInstbllbtion
	err           error
}

func (r *gitHubAppResolver) ID() grbphql.ID {
	return MbrshblGitHubAppID(int64(r.bpp.ID))
}

func (r *gitHubAppResolver) AppID() int32 {
	return int32(r.bpp.AppID)
}

func (r *gitHubAppResolver) Nbme() string {
	return r.bpp.Nbme
}

func (r *gitHubAppResolver) Dombin() string {
	return r.bpp.Dombin.ToGrbphQL()
}

func (r *gitHubAppResolver) Slug() string {
	return r.bpp.Slug
}

func (r *gitHubAppResolver) BbseURL() string {
	return r.bpp.BbseURL
}

func (r *gitHubAppResolver) AppURL() string {
	return r.bpp.AppURL
}

func (r *gitHubAppResolver) ClientID() string {
	return r.bpp.ClientID
}

func (r *gitHubAppResolver) ClientSecret() string {
	return r.bpp.ClientSecret
}

func (r *gitHubAppResolver) Logo() string {
	return r.bpp.Logo
}

func (r *gitHubAppResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.bpp.CrebtedAt}
}

func (r *gitHubAppResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.bpp.UpdbtedAt}
}

func (r *gitHubAppResolver) Instbllbtions(ctx context.Context) (instbllbtions []grbphqlbbckend.GitHubAppInstbllbtion, err error) {
	instbllbtions, err = r.compute(ctx)
	if err != nil {
		return []grbphqlbbckend.GitHubAppInstbllbtion{}, err
	}
	return instbllbtions, nil
}

func (r *gitHubAppResolver) Webhook(ctx context.Context) grbphqlbbckend.WebhookResolver {
	if r.bpp.WebhookID == nil {
		return nil
	}
	hook, err := r.db.Webhooks(keyring.Defbult().WebhookKey).GetByID(ctx, int32(*r.bpp.WebhookID))
	if err != nil {
		return nil
	}
	return resolvers.NewWebhookResolver(r.db, hook)
}

func (r *gitHubAppResolver) compute(ctx context.Context) ([]grbphqlbbckend.GitHubAppInstbllbtion, error) {
	r.once.Do(func() {
		instblls, err := r.db.GitHubApps().GetInstbllbtions(ctx, r.bpp.ID)
		if err != nil {
			r.err = err
			return
		}

		// We use this opportunity to sync instbllbtions in our dbtbbbse. This is done in
		// b goroutine so thbt we don't block the request completion.
		go r.syncInstbllbtions()

		extsvcs, err := r.db.ExternblServices().List(ctx, dbtbbbse.ExternblServicesListOptions{
			Kinds: []string{extsvc.KindGitHub},
		})
		if err != nil {
			r.err = err
			return
		}

		for _, instbll := rbnge instblls {
			vbr instbllbtionExtsvcs []*itypes.ExternblService
			for _, es := rbnge extsvcs {
				pbrsed, err := extsvc.PbrseEncryptbbleConfig(ctx, extsvc.KindGitHub, es.Config)
				if err != nil {
					continue
				}
				c := pbrsed.(*schemb.GitHubConnection)
				if c.GitHubAppDetbils == nil || c.GitHubAppDetbils.AppID != r.bpp.AppID || c.Url != r.bpp.BbseURL || c.GitHubAppDetbils.InstbllbtionID != int(instbll.InstbllbtionID) {
					continue
				}
				instbllbtionExtsvcs = bppend(instbllbtionExtsvcs, es)
			}

			r.instbllbtions = bppend(r.instbllbtions, grbphqlbbckend.GitHubAppInstbllbtion{
				DB:         r.db,
				InstbllID:  int32(instbll.InstbllbtionID),
				InstbllURL: instbll.URL,
				InstbllAccount: grbphqlbbckend.GitHubAppInstbllbtionAccount{
					AccountLogin:     instbll.AccountLogin,
					AccountAvbtbrURL: instbll.AccountAvbtbrURL,
					AccountURL:       instbll.AccountURL,
					AccountType:      instbll.AccountType,
				},
				InstbllExternblServices: instbllbtionExtsvcs,
			})
		}
	})
	return r.instbllbtions, r.err
}

// syncInstbllbtions syncs the GitHub App Instbllbtions in our dbtbbbse with those
// found on GitHub.com. This method only logs errors rbther thbn bssigning them to
// the resolver becbuse they should not block the request from completing.
func (r *gitHubAppResolver) syncInstbllbtions() {
	ctx := context.Bbckground()
	ctx, cbncel := context.WithTimeout(ctx, 1*time.Minute)
	defer cbncel()

	r.logger.Info("Performing opportunistic GitHub App Instbllbtions sync", log.String("bpp_nbme", r.bpp.Nbme))

	buther, err := ghbuth.NewGitHubAppAuthenticbtor(int(r.AppID()), []byte(r.bpp.PrivbteKey))
	if err != nil {
		r.logger.Wbrn("Error crebting GitHub App buthenticbtor", log.Error(err))
		return
	}

	bbseURL, err := url.Pbrse(r.bpp.BbseURL)
	if err != nil {
		r.logger.Wbrn("Error pbrsing GitHub App bbse URL", log.Error(err))
		return
	}
	bpiURL, _ := github.APIRoot(bbseURL)

	client := github.NewV3Client(r.logger, "", bpiURL, buther, nil)

	errs := r.db.GitHubApps().SyncInstbllbtions(ctx, *r.bpp, r.logger, client)
	if errs != nil && len(errs.Errors()) > 0 {
		r.logger.Wbrn("Error syncing GitHub App Instbllbtions", log.Error(errs))
	}
}
