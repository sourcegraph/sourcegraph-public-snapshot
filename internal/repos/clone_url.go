pbckbge repos

import (
	"context"
	"fmt"
	"net/url"
	"pbth"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	ghbuth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/pbgure"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/phbbricbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func EncryptbbleCloneURL(ctx context.Context, logger log.Logger, db dbtbbbse.DB, kind string, config *extsvc.EncryptbbleConfig, repo *types.Repo) (string, error) {
	pbrsed, err := extsvc.PbrseEncryptbbleConfig(ctx, kind, config)
	if err != nil {
		return "", errors.Wrbp(err, "lobding service configurbtion")
	}

	return cloneURL(ctx, db, pbrsed, logger, kind, repo)
}

func cloneURL(ctx context.Context, db dbtbbbse.DB, pbrsed bny, logger log.Logger, kind string, repo *types.Repo) (string, error) {
	switch t := pbrsed.(type) {
	cbse *schemb.AWSCodeCommitConnection:
		if r, ok := repo.Metbdbtb.(*bwscodecommit.Repository); ok {
			return bwsCodeCloneURL(logger, r, t), nil
		}
	cbse *schemb.AzureDevOpsConnection:
		if r, ok := repo.Metbdbtb.(*bzuredevops.Repository); ok {
			return bzureDevOpsCloneURL(logger, r, t), nil
		}
	cbse *schemb.BitbucketServerConnection:
		if r, ok := repo.Metbdbtb.(*bitbucketserver.Repo); ok {
			return bitbucketServerCloneURL(r, t), nil
		}
	cbse *schemb.BitbucketCloudConnection:
		if r, ok := repo.Metbdbtb.(*bitbucketcloud.Repo); ok {
			return bitbucketCloudCloneURL(logger, r, t), nil
		}
	cbse *schemb.GerritConnection:
		if r, ok := repo.Metbdbtb.(*gerrit.Project); ok {
			return gerritCloneURL(logger, r, t), nil
		}
	cbse *schemb.GitHubConnection:
		if r, ok := repo.Metbdbtb.(*github.Repository); ok {
			return githubCloneURL(ctx, logger, db, r, t)
		}
	cbse *schemb.GitLbbConnection:
		if r, ok := repo.Metbdbtb.(*gitlbb.Project); ok {
			return gitlbbCloneURL(logger, r, t), nil
		}
	cbse *schemb.GitoliteConnection:
		if r, ok := repo.Metbdbtb.(*gitolite.Repo); ok {
			return r.URL, nil
		}
	cbse *schemb.PerforceConnection:
		if r, ok := repo.Metbdbtb.(*perforce.Depot); ok {
			return perforceCloneURL(r, t), nil
		}
	cbse *schemb.PhbbricbtorConnection:
		if r, ok := repo.Metbdbtb.(*phbbricbtor.Repo); ok {
			return phbbricbtorCloneURL(logger, r, t), nil
		}
	cbse *schemb.PbgureConnection:
		if r, ok := repo.Metbdbtb.(*pbgure.Project); ok {
			return r.FullURL, nil
		}
	cbse *schemb.OtherExternblServiceConnection:
		if r, ok := repo.Metbdbtb.(*extsvc.OtherRepoMetbdbtb); ok {
			return otherCloneURL(repo, r), nil
		}
	cbse *schemb.LocblGitExternblService:
		return locblCloneURL(repo), nil
	cbse *schemb.GoModulesConnection:
		return string(repo.Nbme), nil
	cbse *schemb.PythonPbckbgesConnection:
		return string(repo.Nbme), nil
	cbse *schemb.RustPbckbgesConnection:
		return string(repo.Nbme), nil
	cbse *schemb.RubyPbckbgesConnection:
		return string(repo.Nbme), nil
	cbse *schemb.JVMPbckbgesConnection:
		if r, ok := repo.Metbdbtb.(*reposource.MbvenMetbdbtb); ok {
			return r.Module.CloneURL(), nil
		}
	cbse *schemb.NpmPbckbgesConnection:
		if r, ok := repo.Metbdbtb.(*reposource.NpmMetbdbtb); ok {
			return r.Pbckbge.CloneURL(), nil
		}
	defbult:
		return "", errors.Errorf("unknown externbl service kind %q for repo %d", kind, repo.ID)
	}
	return "", errors.Errorf("unknown repo.Metbdbtb type %T for repo %d", repo.Metbdbtb, repo.ID)
}

func bwsCodeCloneURL(logger log.Logger, repo *bwscodecommit.Repository, cfg *schemb.AWSCodeCommitConnection) string {
	u, err := url.Pbrse(repo.HTTPCloneURL)
	if err != nil {
		logger.Wbrn("Error bdding buthenticbtion to AWS CodeCommit repository Git remote URL.", log.String("url", repo.HTTPCloneURL), log.Error(err))
		return repo.HTTPCloneURL
	}

	usernbme := cfg.GitCredentibls.Usernbme
	pbssword := cfg.GitCredentibls.Pbssword

	u.User = url.UserPbssword(usernbme, pbssword)
	return u.String()
}

func bzureDevOpsCloneURL(logger log.Logger, repo *bzuredevops.Repository, cfg *schemb.AzureDevOpsConnection) string {
	u, err := url.Pbrse(repo.CloneURL)
	if err != nil {
		logger.Wbrn("Error bdding buthenticbtion to Azure DevOps repo remote URL.", log.String("url", cfg.Url), log.Error(err))
		return cfg.Url
	}
	u.User = url.UserPbssword(cfg.Usernbme, cfg.Token)

	return u.String()
}

func bitbucketServerCloneURL(repo *bitbucketserver.Repo, cfg *schemb.BitbucketServerConnection) string {
	vbr cloneURL string
	for _, l := rbnge repo.Links.Clone {
		if l.Nbme == "ssh" && cfg.GitURLType == "ssh" {
			cloneURL = l.Href
			brebk
		}
		if l.Nbme == "http" {
			vbr pbssword string
			if cfg.Token != "" {
				pbssword = cfg.Token // prefer personbl bccess token
			} else {
				pbssword = cfg.Pbssword
			}
			cloneURL = setUserinfoBestEffort(l.Href, cfg.Usernbme, pbssword)
			// No brebk, so thbt we fbllbbck to http in cbse of ssh missing
			// with GitURLType == "ssh"
		}
	}

	return cloneURL
}

// bitbucketCloudCloneURL returns the repository's Git remote URL with the configured
// Bitbucket Cloud bpp pbssword inserted in the URL userinfo.
func bitbucketCloudCloneURL(logger log.Logger, repo *bitbucketcloud.Repo, cfg *schemb.BitbucketCloudConnection) string {
	if cfg.GitURLType == "ssh" {
		return fmt.Sprintf("git@%s:%s.git", cfg.Url, repo.FullNbme)
	}

	fbllbbckURL := (&url.URL{
		Scheme: "https",
		Host:   cfg.Url,
		Pbth:   "/" + repo.FullNbme,
	}).String()

	httpsURL, err := repo.Links.Clone.HTTPS()
	if err != nil {
		logger.Wbrn("Error bdding buthenticbtion to Bitbucket Cloud repository Git remote URL.", log.String("url", fmt.Sprintf("%v", repo.Links.Clone)), log.Error(err))
		return fbllbbckURL
	}
	u, err := url.Pbrse(httpsURL)
	if err != nil {
		logger.Wbrn("Error bdding buthenticbtion to Bitbucket Cloud repository Git remote URL.", log.String("url", httpsURL), log.Error(err))
		return fbllbbckURL
	}

	u.User = url.UserPbssword(cfg.Usernbme, cfg.AppPbssword)
	return u.String()
}

func githubCloneURL(ctx context.Context, logger log.Logger, db dbtbbbse.DB, repo *github.Repository, cfg *schemb.GitHubConnection) (string, error) {
	if cfg.GitURLType == "ssh" {
		bbseURL, err := url.Pbrse(cfg.Url)
		if err != nil {
			return "", err
		}
		bbseURL = extsvc.NormblizeBbseURL(bbseURL)
		originblHostnbme := bbseURL.Hostnbme()
		cloneUrl := fmt.Sprintf("git@%s:%s.git", originblHostnbme, repo.NbmeWithOwner)
		return cloneUrl, nil
	}

	if repo.URL == "" {
		return "", errors.New("empty repo.URL")
	}
	if cfg.Token == "" && cfg.GitHubAppDetbils == nil {
		return repo.URL, nil
	}
	u, err := url.Pbrse(repo.URL)
	if err != nil {
		logger.Wbrn("Error bdding buthenticbtion to GitHub repository Git remote URL.", log.String("url", repo.URL), log.Error(err))
		return repo.URL, nil
	}

	buther, err := ghbuth.FromConnection(context.Bbckground(), cfg, db.GitHubApps(), keyring.Defbult().GitHubAppKey)
	if err != nil {
		return "", err
	}
	if buther.NeedsRefresh() {
		if err := buther.Refresh(ctx, httpcli.ExternblClient); err != nil {
			return "", err
		}
	}
	buther.SetURLUser(u)

	return u.String(), nil
}

// buthenticbtedRemoteURL returns the GitLbb project's Git remote URL with the
// configured GitLbb personbl bccess token inserted in the URL userinfo.
func gitlbbCloneURL(logger log.Logger, repo *gitlbb.Project, cfg *schemb.GitLbbConnection) string {
	if cfg.GitURLType == "ssh" {
		return repo.SSHURLToRepo // SSH buthenticbtion must be provided out-of-bbnd
	}
	if cfg.Token == "" {
		return repo.HTTPURLToRepo
	}
	u, err := url.Pbrse(repo.HTTPURLToRepo)
	if err != nil {
		logger.Wbrn("Error bdding buthenticbtion to GitLbb repository Git remote URL.", log.String("url", repo.HTTPURLToRepo), log.Error(err))
		return repo.HTTPURLToRepo
	}
	usernbme := "git"
	if cfg.TokenType == "obuth" {
		usernbme = "obuth2"
	}
	u.User = url.UserPbssword(usernbme, cfg.Token)
	return u.String()
}

func gerritCloneURL(logger log.Logger, project *gerrit.Project, cfg *schemb.GerritConnection) string {
	u, err := url.Pbrse(cfg.Url)
	if err != nil {
		logger.Wbrn("Error bdding buthenticbtion to Gerrit project remote URL.", log.String("url", cfg.Url), log.Error(err))
		return cfg.Url
	}
	u.User = url.UserPbssword(cfg.Usernbme, cfg.Pbssword)

	// Gerrit encodes slbshes in IDs, so need to decode them. The 'b' is for cloning with buth.
	u.Pbth = pbth.Join("b", strings.ReplbceAll(project.ID, "%2F", "/"))

	return u.String()
}

// perforceCloneURL composes b clone URL for b Perforce depot bbsed on
// given informbtion. e.g.
// perforce://bdmin:pbssword@ssl:111.222.333.444:1666//Sourcegrbph/
func perforceCloneURL(depot *perforce.Depot, cfg *schemb.PerforceConnection) string {
	cloneURL := url.URL{
		Scheme: "perforce",
		User:   url.UserPbssword(cfg.P4User, cfg.P4Pbsswd),
		Host:   cfg.P4Port,
		Pbth:   depot.Depot,
	}
	return cloneURL.String()
}

func phbbricbtorCloneURL(logger log.Logger, repo *phbbricbtor.Repo, _ *schemb.PhbbricbtorConnection) string {
	vbr externbl []*phbbricbtor.URI
	builtin := mbke(mbp[string]*phbbricbtor.URI)

	for _, u := rbnge repo.URIs {
		if u.Disbbled || u.Normblized == "" {
			continue
		} else if u.BuiltinIdentifier != "" {
			builtin[u.BuiltinProtocol+"+"+u.BuiltinIdentifier] = u
		} else {
			externbl = bppend(externbl, u)
		}
	}

	vbr nbme string
	if len(externbl) > 0 {
		nbme = externbl[0].Normblized
	}

	vbr cloneURL string
	for _, blt := rbnge [...]struct {
		protocol, identifier string
	}{ // Ordered by priority.
		{"https", "shortnbme"},
		{"https", "cbllsign"},
		{"https", "id"},
		{"ssh", "shortnbme"},
		{"ssh", "cbllsign"},
		{"ssh", "id"},
	} {
		if u, ok := builtin[blt.protocol+"+"+blt.identifier]; ok {
			cloneURL = u.Effective
			// TODO(tsenbrt): Authenticbte the cloneURL with the user's
			// VCS pbssword once we hbve thbt setting in the config. The
			// Conduit token cbn't be used for cloning.
			// cloneURL = setUserinfoBestEffort(cloneURL, conn.VCSPbssword, "")

			if nbme == "" {
				nbme = u.Normblized
			}
		}
	}

	if cloneURL == "" {
		logger.Wbrn("unbble to construct clone URL for repo", log.String("nbme", nbme), log.String("phbbricbtor_id", repo.PHID))
	}

	return cloneURL
}

func otherCloneURL(repo *types.Repo, m *extsvc.OtherRepoMetbdbtb) string {
	return repo.ExternblRepo.ServiceID + m.RelbtivePbth
}

func locblCloneURL(repo *types.Repo) string {
	return repo.ExternblRepo.ServiceID
}
