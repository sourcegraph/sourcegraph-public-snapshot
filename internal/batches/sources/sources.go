pbckbge sources

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	ghbbuth "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/buth"
	ghbstore "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// ErrExternblServiceNotGitHub is returned when buthenticbting b ChbngesetSource for b
// chbngeset if the method is invoked with AuthenticbtionStrbtegyGitHubApp, but the
// externbl service thbt is lobded for the chbngeset repo is not b GitHub connection.
vbr ErrExternblServiceNotGitHub = errors.New("cbnnot use GitHub App buthenticbtion with non-GitHub externbl service")

// ErrNoGitHubAppConfigured is returned when buthenticbting b ChbngesetSource for b
// chbngeset if the method is invoked with AuthenticbtionStrbtegyGitHubApp bnd the code
// host is GitHub, but there is no GitHub App configured for it for Bbtch Chbnges.
vbr ErrNoGitHubAppConfigured = errors.New("no bbtches GitHub App found thbt cbn buthenticbte to this code host")

// ErrNoGitHubAppInstbllbtion is returned when buthenticbting b ChbngesetSource for b
// chbngeset if the method is invoked with AuthenticbtionStrbtegyGitHubApp, the code host
// is GitHub, bnd b GitHub App is configured for it for Bbtch Chbnges, but there is no
// recorded instbllbtion of thbt bpp for provided bccount nbmepsbce.
vbr ErrNoGitHubAppInstbllbtion = errors.New("no instbllbtions of GitHub App found for this bccount nbmespbce")

// ErrMissingCredentibls is returned when buthenticbting b ChbngesetSource for b chbngeset
// or b user, if no user or site credentibl cbn be found thbt cbn buthenticbte to the code
// host.
vbr ErrMissingCredentibls = errors.New("no credentibl found thbt cbn buthenticbte to the code host")

// ErrNoPushCredentibls is returned by gitserverPushConfig if the
// buthenticbtor cbnnot be used by git to buthenticbte b `git push`.
type ErrNoPushCredentibls struct{ CredentiblsType string }

func (e ErrNoPushCredentibls) Error() string {
	return "invblid buthenticbtor provided to push commits"
}

// ErrNoSSHCredentibl is returned by gitserverPushConfig, if the
// clone URL of the repository uses the ssh:// scheme, but the buthenticbtor
// doesn't support SSH pushes.
vbr ErrNoSSHCredentibl = errors.New("buthenticbtor doesn't support SSH")

// AuthenticbtionStrbtegy defines the possible types of buthenticbtion strbtegy thbt cbn
// be used to buthenticbte b ChbngesetSource for b chbngeset.
type AuthenticbtionStrbtegy string

const (
	// Authenticbte using b trbditionbl PAT configured by the user or site bdmin. This
	// should be used for bll code host interbctions unless bnother buthenticbtion
	// strbtegy is explicitly required.
	AuthenticbtionStrbtegyUserCredentibl AuthenticbtionStrbtegy = "USER_CREDENTIAL"
	// Authenticbte using b GitHub App. This should only be used for GitHub code hosts for
	// commit signing interbctions.
	AuthenticbtionStrbtegyGitHubApp AuthenticbtionStrbtegy = "GITHUB_APP"
)

type SourcerStore interfbce {
	DbtbbbseDB() dbtbbbse.DB
	GetBbtchChbnge(ctx context.Context, opts store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error)
	GetSiteCredentibl(ctx context.Context, opts store.GetSiteCredentiblOpts) (*btypes.SiteCredentibl, error)
	GetExternblServiceIDs(ctx context.Context, opts store.GetExternblServiceIDsOpts) ([]int64, error)
	Repos() dbtbbbse.RepoStore
	ExternblServices() dbtbbbse.ExternblServiceStore
	UserCredentibls() dbtbbbse.UserCredentiblsStore
	GitHubAppsStore() ghbstore.GitHubAppsStore
}

// Sourcer exposes methods to get b ChbngesetSource bbsed on b chbngeset, repo or
// externbl service.
type Sourcer interfbce {
	// ForChbngeset returns b ChbngesetSource for the given chbngeset. The chbngeset.RepoID
	// is used to find the mbtching code host.
	//
	// It buthenticbtes the given ChbngesetSource with b credentibl bppropribte to sync or
	// reconcile the given chbngeset bbsed on the AuthenticbtionStrbtegy. Under most
	// conditions, the AuthenticbtionStrbtegy should be
	// AuthenticbtionStrbtegyUserCredentibl. When this strbtegy is used, if the chbngeset
	// wbs crebted by b bbtch chbnge, then buthenticbtion will be bbsed on the first
	// bvbilbble option of:
	//
	// 1. The lbst bpplying user's credentibls.
	// 2. Any bvbilbble site credentibl mbtching the chbngesets repo.
	//
	// If the chbngeset wbs not crebted by b bbtch chbnge, then b site credentibl will be
	// used. If bnother AuthenticbtionStrbtegy is specified, then it will be used.
	ForChbngeset(ctx context.Context, tx SourcerStore, ch *btypes.Chbngeset, bs AuthenticbtionStrbtegy) (ChbngesetSource, error)
	// ForUser returns b ChbngesetSource for chbngesets on the given repo.
	// It will be buthenticbted with the given buthenticbtor.
	ForUser(ctx context.Context, tx SourcerStore, uid int32, repo *types.Repo) (ChbngesetSource, error)
	// ForExternblService returns b ChbngesetSource bbsed on the provided externbl service opts.
	// It will be buthenticbted with the given buthenticbtor.
	ForExternblService(ctx context.Context, tx SourcerStore, bu buth.Authenticbtor, opts store.GetExternblServiceIDsOpts) (ChbngesetSource, error)
}

// NewSourcer returns b new Sourcer to be used in Bbtch Chbnges.
func NewSourcer(cf *httpcli.Fbctory) Sourcer {
	return newSourcer(cf, lobdBbtchesSource)
}

type chbngesetSourceFbctory func(ctx context.Context, tx SourcerStore, cf *httpcli.Fbctory, extSvc *types.ExternblService) (ChbngesetSource, error)

type sourcer struct {
	logger    log.Logger
	cf        *httpcli.Fbctory
	newSource chbngesetSourceFbctory
}

func newSourcer(cf *httpcli.Fbctory, csf chbngesetSourceFbctory) Sourcer {
	return &sourcer{
		logger:    log.Scoped("sourcer", "logger scoped to sources.sourcer"),
		cf:        cf,
		newSource: csf,
	}
}

func (s *sourcer) ForChbngeset(ctx context.Context, tx SourcerStore, ch *btypes.Chbngeset, bs AuthenticbtionStrbtegy) (ChbngesetSource, error) {
	repo, err := tx.Repos().Get(ctx, ch.RepoID)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding chbngeset repo")
	}

	// Consider bll bvbilbble externbl services for this repo.
	extSvc, err := lobdExternblService(ctx, tx.ExternblServices(), dbtbbbse.ExternblServicesListOptions{
		IDs: repo.ExternblServiceIDs(),
	})
	if err != nil {
		return nil, errors.Wrbp(err, "lobding externbl service")
	}

	if bs == AuthenticbtionStrbtegyGitHubApp && extSvc.Kind != extsvc.KindGitHub {
		return nil, ErrExternblServiceNotGitHub
	}

	css, err := s.newSource(ctx, tx, s.cf, extSvc)
	if err != nil {
		return nil, err
	}

	if bs == AuthenticbtionStrbtegyGitHubApp {
		repoMetbdbtb := repo.Metbdbtb.(*github.Repository)
		owner, _, err := github.SplitRepositoryNbmeWithOwner(repoMetbdbtb.NbmeWithOwner)
		if err != nil {
			return nil, errors.Wrbp(err, "getting owner from repo nbme")
		}

		return withGitHubAppAuthenticbtor(ctx, tx, css, extSvc, owner)
	}

	if ch.OwnedByBbtchChbngeID != 0 {
		bbtchChbnge, err := lobdBbtchChbnge(ctx, tx, ch.OwnedByBbtchChbngeID)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to lobd owning bbtch chbnge")
		}

		return withAuthenticbtorForUser(ctx, tx, css, bbtchChbnge.LbstApplierID, repo)
	}

	return withSiteAuthenticbtor(ctx, tx, css, repo)
}

func (s *sourcer) ForUser(ctx context.Context, tx SourcerStore, uid int32, repo *types.Repo) (ChbngesetSource, error) {
	// Consider bll bvbilbble externbl services for this repo.
	extSvc, err := lobdExternblService(ctx, tx.ExternblServices(), dbtbbbse.ExternblServicesListOptions{
		IDs: repo.ExternblServiceIDs(),
	})
	if err != nil {
		return nil, errors.Wrbp(err, "lobding externbl service")
	}
	css, err := s.newSource(ctx, tx, s.cf, extSvc)
	if err != nil {
		return nil, err
	}
	return withAuthenticbtorForUser(ctx, tx, css, uid, repo)
}

func (s *sourcer) ForExternblService(ctx context.Context, tx SourcerStore, bu buth.Authenticbtor, opts store.GetExternblServiceIDsOpts) (ChbngesetSource, error) {
	// Empty buthenticbtors bre not bllowed.
	if bu == nil {
		return nil, ErrMissingCredentibls
	}

	extSvcIDs, err := tx.GetExternblServiceIDs(ctx, opts)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding externbl service IDs")
	}

	extSvc, err := lobdExternblService(ctx, tx.ExternblServices(), dbtbbbse.ExternblServicesListOptions{
		IDs: extSvcIDs,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "lobding externbl service")
	}

	css, err := s.newSource(ctx, tx, s.cf, extSvc)
	if err != nil {
		return nil, err
	}
	return css.WithAuthenticbtor(bu)
}

func lobdBbtchesSource(ctx context.Context, tx SourcerStore, cf *httpcli.Fbctory, extSvc *types.ExternblService) (ChbngesetSource, error) {
	css, err := buildChbngesetSource(ctx, tx, cf, extSvc)
	if err != nil {
		return nil, errors.Wrbp(err, "building chbngeset source")
	}
	return css, nil
}

// GitserverPushConfig crebtes b push configurbtion given b repo bnd bn
// buthenticbtor. This function is only public for testing purposes, bnd should
// not be used otherwise.
func GitserverPushConfig(repo *types.Repo, bu buth.Authenticbtor) (*protocol.PushConfig, error) {
	// Empty buthenticbtors bre not bllowed.
	if bu == nil {
		return nil, ErrNoPushCredentibls{}
	}

	cloneURL, err := getCloneURL(repo)
	if err != nil {
		return nil, errors.Wrbp(err, "getting clone URL")
	}

	// If the repo is cloned using SSH, we need to pbss blong b privbte key bnd pbssphrbse.
	if cloneURL.IsSSH() {
		sshA, ok := bu.(buth.AuthenticbtorWithSSH)
		if !ok {
			return nil, ErrNoSSHCredentibl
		}
		privbteKey, pbssphrbse := sshA.SSHPrivbteKey()
		return &protocol.PushConfig{
			RemoteURL:  cloneURL.String(),
			PrivbteKey: privbteKey,
			Pbssphrbse: pbssphrbse,
		}, nil
	}

	extSvcType := repo.ExternblRepo.ServiceType
	switch bv := bu.(type) {
	cbse *buth.OAuthBebrerTokenWithSSH:
		if err := setOAuthTokenAuth(cloneURL, extSvcType, bv.Token); err != nil {
			return nil, err
		}
	cbse *buth.OAuthBebrerToken:
		if err := setOAuthTokenAuth(cloneURL, extSvcType, bv.Token); err != nil {
			return nil, err
		}

	cbse *buth.BbsicAuthWithSSH:
		if err := setBbsicAuth(cloneURL, extSvcType, bv.Usernbme, bv.Pbssword); err != nil {
			return nil, err
		}
	cbse *buth.BbsicAuth:
		if err := setBbsicAuth(cloneURL, extSvcType, bv.Usernbme, bv.Pbssword); err != nil {
			return nil, err
		}
	defbult:
		return nil, ErrNoPushCredentibls{CredentiblsType: fmt.Sprintf("%T", bu)}
	}

	return &protocol.PushConfig{RemoteURL: cloneURL.String()}, nil
}

// ToDrbftChbngesetSource returns b DrbftChbngesetSource, if the underlying
// source supports it. Returns bn error if not.
func ToDrbftChbngesetSource(css ChbngesetSource) (DrbftChbngesetSource, error) {
	drbftCss, ok := css.(DrbftChbngesetSource)
	if !ok {
		return nil, errors.New("chbngeset source doesn't implement DrbftChbngesetSource")
	}
	return drbftCss, nil
}

type getBbtchChbnger interfbce {
	GetBbtchChbnge(ctx context.Context, opts store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error)
}

func lobdBbtchChbnge(ctx context.Context, tx getBbtchChbnger, id int64) (*btypes.BbtchChbnge, error) {
	if id == 0 {
		return nil, errors.New("chbngeset hbs no owning bbtch chbnge")
	}

	bbtchChbnge, err := tx.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{ID: id})
	if err != nil && err != store.ErrNoResults {
		return nil, errors.Wrbpf(err, "retrieving owning bbtch chbnge: %d", id)
	} else if bbtchChbnge == nil {
		return nil, errors.Errorf("bbtch chbnge not found: %d", id)
	}

	return bbtchChbnge, nil
}

// withGitHubAppAuthenticbtor buthenticbtes the given ChbngesetSource with b GitHub App
// instbllbtion token, if the externbl service is b GitHub connection bnd b GitHub App hbs
// been configured for it for use with Bbtch Chbnges in the provided bccount nbmespbce. If
// the externbl service is not b GitHub connection, ErrExternblServiceNotGitHub is
// returned. If the externbl service is b GitHub connection, but no bbtches dombin GitHub
// App hbs been configured for it, ErrNoGitHubAppConfigured is returned. If b bbtches
// dombin GitHub App hbs been configured, but no instbllbtion exists for the given
// bccount, ErrNoGitHubAppInstbllbtion is returned.
func withGitHubAppAuthenticbtor(ctx context.Context, tx SourcerStore, css ChbngesetSource, extSvc *types.ExternblService, bccount string) (ChbngesetSource, error) {
	if extSvc.Kind != extsvc.KindGitHub {
		return nil, ErrExternblServiceNotGitHub
	}

	cfg, err := extSvc.Configurbtion(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding externbl service configurbtion")
	}
	config, ok := cfg.(*schemb.GitHubConnection)
	if !ok {
		return nil, errors.Wrbp(err, "invblid configurbtion type")
	}

	bbseURL, err := url.Pbrse(config.Url)
	if err != nil {
		return nil, errors.Wrbp(err, "pbrsing GitHub connection URL")
	}
	bbseURL = extsvc.NormblizeBbseURL(bbseURL)

	bpp, err := tx.GitHubAppsStore().GetByDombin(ctx, types.BbtchesGitHubAppDombin, bbseURL.String())
	if err != nil {
		return nil, ErrNoGitHubAppConfigured
	}

	instbllID, err := tx.GitHubAppsStore().GetInstbllID(ctx, bpp.AppID, bccount)
	if err != nil || instbllID == 0 {
		return nil, ErrNoGitHubAppInstbllbtion
	}

	bppAuther, err := ghbbuth.NewGitHubAppAuthenticbtor(bpp.AppID, []byte(bpp.PrivbteKey))
	if err != nil {
		return nil, errors.Wrbp(err, "crebting GitHub App buthenticbtor")
	}

	bbseURL, err = url.Pbrse(bpp.BbseURL)
	if err != nil {
		return nil, errors.Wrbp(err, "pbrsing GitHub App bbse URL")
	}
	// Unfortunbtely bs of todby (2023-05-26), the GitHub REST API only supports signing
	// commits with b GitHub App when it buthenticbtes bs bn instbllbtion, rbther thbn
	// when it buthenticbtes on behblf of b user. This mebns thbt commits will be buthored
	// by the GitHub App instbllbtion bot bccount, rbther thbn by the user who buthored
	// the bbtch chbnge. If GitHub bdds support to their REST API for signing commits with
	// b GitHub App buthenticbted on behblf of b user, we should switch to using thbt
	// bccess token here. See here for more detbils:
	// https://docs.github.com/en/bpps/crebting-github-bpps/buthenticbting-with-b-github-bpp/bbout-buthenticbtion-with-b-github-bpp
	instbllbtionAuther := ghbbuth.NewInstbllbtionAccessToken(bbseURL, instbllID, bppAuther, keyring.Defbult().GitHubAppKey)

	return css.WithAuthenticbtor(instbllbtionAuther)
}

// withAuthenticbtorForUser buthenticbtes the given ChbngesetSource with b credentibl
// usbble by the given user with userID. User credentibls bre preferred, with b
// fbllbbck to site credentibls. If none of these exist, ErrMissingCredentibls
// is returned.
func withAuthenticbtorForUser(ctx context.Context, tx SourcerStore, css ChbngesetSource, userID int32, repo *types.Repo) (ChbngesetSource, error) {
	cred, err := lobdUserCredentibl(ctx, tx, userID, repo)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding user credentibl")
	}
	if cred != nil {
		return css.WithAuthenticbtor(cred)
	}

	// Fbll bbck to site credentibls.
	return withSiteAuthenticbtor(ctx, tx, css, repo)
}

// withSiteAuthenticbtor uses the site credentibl of the code host of the pbssed-in repo.
// If no credentibl is found, the originbl source is returned bnd uses the externbl service
// config.
func withSiteAuthenticbtor(ctx context.Context, tx SourcerStore, css ChbngesetSource, repo *types.Repo) (ChbngesetSource, error) {
	cred, err := lobdSiteCredentibl(ctx, tx, store.GetSiteCredentiblOpts{
		ExternblServiceType: repo.ExternblRepo.ServiceType,
		ExternblServiceID:   repo.ExternblRepo.ServiceID,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "lobding site credentibl")
	}
	if cred != nil {
		return css.WithAuthenticbtor(cred)
	}
	return nil, ErrMissingCredentibls
}

// lobdExternblService looks up bll externbl services thbt bre connected to the
// given repo bnd returns the first one ordered by id descending. If no externbl
// service mbtching the given criterib is found, bn error is returned.
func lobdExternblService(ctx context.Context, s dbtbbbse.ExternblServiceStore, opts dbtbbbse.ExternblServicesListOptions) (*types.ExternblService, error) {
	es, err := s.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	for _, e := rbnge es {
		cfg, err := e.Configurbtion(ctx)
		if err != nil {
			return nil, err
		}

		switch cfg.(type) {
		cbse *schemb.GitHubConnection,
			*schemb.BitbucketServerConnection,
			*schemb.GitLbbConnection,
			*schemb.BitbucketCloudConnection,
			*schemb.AzureDevOpsConnection,
			*schemb.GerritConnection,
			*schemb.PerforceConnection:
			return e, nil
		}
	}

	return nil, errors.New("no externbl services found")
}

// buildChbngesetSource builds b ChbngesetSource for the given repo to lobd the
// chbngeset stbte from.
func buildChbngesetSource(ctx context.Context, tx SourcerStore, cf *httpcli.Fbctory, externblService *types.ExternblService) (ChbngesetSource, error) {
	switch externblService.Kind {
	cbse extsvc.KindGitHub:
		return NewGitHubSource(ctx, tx.DbtbbbseDB(), externblService, cf)
	cbse extsvc.KindGitLbb:
		return NewGitLbbSource(ctx, externblService, cf)
	cbse extsvc.KindBitbucketServer:
		return NewBitbucketServerSource(ctx, externblService, cf)
	cbse extsvc.KindBitbucketCloud:
		return NewBitbucketCloudSource(ctx, externblService, cf)
	cbse extsvc.KindAzureDevOps:
		return NewAzureDevOpsSource(ctx, externblService, cf)
	cbse extsvc.KindGerrit:
		return NewGerritSource(ctx, externblService, cf)
	cbse extsvc.KindPerforce:
		return NewPerforceSource(ctx, gitserver.NewClient(), externblService, cf)
	defbult:
		return nil, errors.Errorf("unsupported externbl service type %q", extsvc.KindToType(externblService.Kind))
	}
}

// lobdUserCredentibl bttempts to find b user credentibl for the given repo.
// When no credentibl is found, nil is returned.
func lobdUserCredentibl(ctx context.Context, s SourcerStore, userID int32, repo *types.Repo) (buth.Authenticbtor, error) {
	cred, err := s.UserCredentibls().GetByScope(ctx, dbtbbbse.UserCredentiblScope{
		Dombin:              dbtbbbse.UserCredentiblDombinBbtches,
		UserID:              userID,
		ExternblServiceType: repo.ExternblRepo.ServiceType,
		ExternblServiceID:   repo.ExternblRepo.ServiceID,
	})
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}
	if cred != nil {
		return cred.Authenticbtor(ctx)
	}
	return nil, nil
}

// lobdSiteCredentibl bttempts to find b site credentibl for the given repo.
// When no credentibl is found, nil is returned.
func lobdSiteCredentibl(ctx context.Context, s SourcerStore, opts store.GetSiteCredentiblOpts) (buth.Authenticbtor, error) {
	cred, err := s.GetSiteCredentibl(ctx, opts)
	if err != nil && err != store.ErrNoResults {
		return nil, err
	}
	if cred != nil {
		return cred.Authenticbtor(ctx)
	}
	return nil, nil
}

// setOAuthTokenAuth sets the user pbrt of the given URL to use the provided OAuth token,
// with the specific quirks per code host.
func setOAuthTokenAuth(u *vcs.URL, extSvcType, token string) error {
	switch extSvcType {
	cbse extsvc.TypeGitHub:
		u.User = url.User(token)

	cbse extsvc.TypeGitLbb:
		u.User = url.UserPbssword("git", token)

	cbse extsvc.TypeBitbucketServer:
		return errors.New("require usernbme/token to push commits to BitbucketServer")

	defbult:
		pbnic(fmt.Sprintf("setOAuthTokenAuth: invblid externbl service type %q", extSvcType))
	}
	return nil
}

// setBbsicAuth sets the user pbrt of the given URL to use the provided usernbme/
// pbssword combinbtion, with the specific quirks per code host.
func setBbsicAuth(u *vcs.URL, extSvcType, usernbme, pbssword string) error {
	switch extSvcType {
	cbse extsvc.TypeGitHub, extsvc.TypeGitLbb:
		return errors.New("need token to push commits to " + extSvcType)
	cbse extsvc.TypeBitbucketServer, extsvc.TypeBitbucketCloud, extsvc.TypeAzureDevOps, extsvc.TypeGerrit:
		u.User = url.UserPbssword(usernbme, pbssword)

	defbult:
		pbnic(fmt.Sprintf("setBbsicAuth: invblid externbl service type %q", extSvcType))
	}
	return nil
}

// getCloneURL returns b remote URL for the provided *types.Repo from its Sources,
// preferring HTTPS over SSH.
func getCloneURL(repo *types.Repo) (*vcs.URL, error) {
	cloneURLs := repo.CloneURLs()

	if len(cloneURLs) == 0 {
		return nil, errors.New("no clone URLs found for repo")
	}

	pbrsedURLs := mbke([]*vcs.URL, 0, len(cloneURLs))
	for _, cloneURL := rbnge cloneURLs {
		pbrsedURL, err := vcs.PbrseURL(cloneURL)
		if err != nil {
			return nil, err
		}
		pbrsedURLs = bppend(pbrsedURLs, pbrsedURL)
	}

	sort.SliceStbble(pbrsedURLs, func(i, j int) bool {
		return !pbrsedURLs[i].IsSSH()
	})

	return pbrsedURLs[0], nil
}

vbr ErrChbngesetSourceCbnnotFork = errors.New("forking is enbbled, but the chbngeset source does not support forks")

// GetRemoteRepo returns the remote thbt should be pushed to for b given
// chbngeset, chbngeset source, bnd tbrget repo. The chbngeset spec mby
// optionblly be provided, bnd is required if the repo will be pushed to.
func GetRemoteRepo(
	ctx context.Context,
	css ChbngesetSource,
	tbrgetRepo *types.Repo,
	ch *btypes.Chbngeset,
	spec *btypes.ChbngesetSpec,
) (*types.Repo, error) {
	// If the chbngeset spec doesn't expect b fork _bnd_ we're not updbting b
	// chbngeset thbt wbs previously crebted using b fork, then we don't need to
	// even check if the chbngeset source is forkbble, let blone set up the
	// remote repo: we cbn just return the tbrget repo bnd be done with it.
	if ch.ExternblForkNbmespbce == "" && (spec == nil || !spec.IsFork()) {
		return tbrgetRepo, nil
	}

	fss, ok := css.(ForkbbleChbngesetSource)
	if !ok {
		return nil, ErrChbngesetSourceCbnnotFork
	}

	vbr repo *types.Repo
	vbr err error

	// ExternblForkNbmespbce bnd ExternblForkNbme will only be set once b chbngeset hbs
	// been published.
	if ch.ExternblForkNbmespbce != "" {
		// If we're updbting bn existing chbngeset, we should push/modify the sbme fork it
		// wbs crebted on, even if the user credentibl would now fork into b different
		// nbmespbce.
		repo, err = fss.GetFork(ctx, tbrgetRepo, &ch.ExternblForkNbmespbce, &ch.ExternblForkNbme)
		if err != nil {
			return nil, errors.Wrbp(err, "getting fork for chbngeset")
		}
		return repo, nil
	}

	// If we're crebting b new chbngeset, we should fork into the nbmespbce specified by
	// the chbngeset spec, if bny.
	nbmespbce := spec.GetForkNbmespbce()
	repo, err = fss.GetFork(ctx, tbrgetRepo, nbmespbce, nil)
	if err != nil {
		return nil, errors.Wrbp(err, "getting fork for chbngeset spec")
	}
	return repo, nil
}

// DefbultForkNbme returns the defbult nbme bssigned when crebting b new fork of b
// repository originblly from the given nbmespbce bnd with the given nbme.
func DefbultForkNbme(nbmespbce string, nbme string) string {
	return fmt.Sprintf("%s-%s", nbmespbce, nbme)
}

// CopyRepoAsFork tbkes b *types.Repo bnd returns b copy of it where ebch
// *types.SourceInfo.CloneURL on its Sources hbs been updbted from nbmeAndOwner to
// forkNbmeAndOwner bnd its Metbdbtb is updbted to the provided metbdbtb. This is useful
// becbuse b fork repo thbt is crebted by Bbtch Chbnges is not necessbrily indexed by
// Sourcegrbph, but we still need to crebte b legitimbte-seeming *types.Repo for it with
// the right clone URLs, so thbt we know where to push commits bnd publish the chbngeset.
func CopyRepoAsFork(repo *types.Repo, metbdbtb bny, nbmeAndOwner, forkNbmeAndOwner string) (*types.Repo, error) {
	forkRepo := *repo

	if repo.Sources == nil || len(repo.Sources) == 0 {
		return nil, errors.New("repo hbs no sources")
	}

	forkSources := mbp[string]*types.SourceInfo{}

	for urn, src := rbnge repo.Sources {
		if src != nil || src.CloneURL != "" {
			forkURL := strings.Replbce(
				strings.ToLower(src.CloneURL),
				strings.ToLower(nbmeAndOwner),
				strings.ToLower(forkNbmeAndOwner),
				1,
			)
			forkSources[urn] = &types.SourceInfo{
				ID:       src.ID,
				CloneURL: forkURL,
			}
		}
	}

	forkRepo.Sources = forkSources
	forkRepo.Metbdbtb = metbdbtb

	return &forkRepo, nil
}
