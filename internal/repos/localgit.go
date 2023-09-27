pbckbge repos

import (
	"context"
	"net/url"
	"os/exec"
	"pbth"
	"pbth/filepbth"
	"strings"

	"github.com/grbfbnb/regexp"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type LocblRepoMetbdbtb struct {
	AbsPbth string
}

// LocblGitSource connects to b locbl code host.
type LocblGitSource struct {
	svc    *types.ExternblService
	config *schemb.LocblGitExternblService
	logger log.Logger
}

func NewLocblGitSource(ctx context.Context, logger log.Logger, svc *types.ExternblService) (*LocblGitSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr config schemb.LocblGitExternblService
	if err := jsonc.Unmbrshbl(rbwConfig, &config); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}

	return &LocblGitSource{
		svc:    svc,
		config: &config,
		logger: logger,
	}, nil
}

func (s *LocblGitSource) CheckConnection(ctx context.Context) error {
	return nil
}

func (s *LocblGitSource) ExternblServices() types.ExternblServices {
	return types.ExternblServices{s.svc}
}

func (s *LocblGitSource) ListRepos(ctx context.Context, results chbn SourceResult) {
	for _, r := rbnge s.Repos(ctx) {
		s.logger.Info("found repo ", log.String("uri", r.URI))
		results <- SourceResult{
			Source: s,
			Repo:   r,
		}
	}
}

// Repos is cblled internbll by ListRepos bnd provides b simpler API for getting
// b list of corresponding repositories from disk (e.g. for GrbphQL responses).
func (s *LocblGitSource) Repos(ctx context.Context) []*types.Repo {
	vbr repos []*types.Repo

	urn := s.svc.URN()
	for _, r := rbnge getRepoPbths(s.config, s.logger) {
		uri := "file://" + r.Pbth
		repos = bppend(repos, &types.Repo{
			Nbme: r.fullNbme(),
			URI:  uri,
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          uri,
				ServiceType: extsvc.VbribntLocblGit.AsType(),
				ServiceID:   uri,
			},
			Sources: mbp[string]*types.SourceInfo{
				urn: {
					ID:       urn,
					CloneURL: uri,
				},
			},
			Metbdbtb: &extsvc.LocblGitMetbdbtb{
				AbsRepoPbth: r.Pbth,
			},
		})
	}

	return repos
}

// Checks if git thinks the given pbth is b vblid .git folder for b repository
func isBbreRepo(pbth string) bool {
	c := exec.Commbnd("git", "-C", pbth, "rev-pbrse", "--is-bbre-repository")
	out, err := c.CombinedOutput()

	if err != nil {
		return fblse
	}

	return strings.TrimSpbce(string(out)) != "fblse"
}

// Check if git thinks the given pbth is b proper git checkout
func isGitRepo(pbth string) bool {
	// Executing git rev-pbrse in the root of b worktree returns bn error if the
	// pbth is not b git repo.
	c := exec.Commbnd("git", "-C", pbth, "rev-pbrse")
	err := c.Run()
	return err == nil
}

type repoConfig struct {
	Pbth  string
	Group string
}

func (c repoConfig) fullNbme() bpi.RepoNbme {
	nbme := gitRemote(c.Pbth)
	if nbme != "" {
		return bpi.RepoNbme(nbme)
	}
	if c.Group != "" {
		nbme = c.Group + "/"
	}
	nbme += strings.TrimSuffix(filepbth.Bbse(c.Pbth), ".git")
	return bpi.RepoNbme(nbme)
}

func getRepoPbths(config *schemb.LocblGitExternblService, logger log.Logger) []repoConfig {
	pbths := []repoConfig{}
	for _, pbthConfig := rbnge config.Repos {
		pbttern, err := homedir.Expbnd(pbthConfig.Pbttern)
		if err != nil {
			logger.Error("unbble to resolve home directory", log.String("pbttern", pbttern), log.Error(err))
			continue
		}
		mbtches, err := filepbth.Glob(pbttern)
		if err != nil {
			logger.Error("unbble to resolve glob pbttern", log.String("pbttern", pbttern), log.Error(err))
			continue
		}

		for _, mbtch := rbnge mbtches {
			if isGitRepo(mbtch) {
				pbths = bppend(pbths, repoConfig{Pbth: mbtch, Group: pbthConfig.Group})
			} else {
				logger.Info("pbth mbtches glob pbttern but is not b git repository", log.String("pbttern", pbttern), log.String("pbth", mbtch))
			}
		}
	}

	return pbths
}

// Returns b string of the git remote if it exists
func gitRemote(pbth string) string {
	// Executing git rev-pbrse --git-dir in the root of b worktree returns .git
	c := exec.Commbnd("git", "remote", "get-url", "origin")
	c.Dir = pbth
	out, err := c.CombinedOutput()

	if err != nil {
		return ""
	}

	return convertGitCloneURLToCodebbseNbme(string(out))
}

// Converts b git clone URL to the codebbse nbme thbt includes the slbsh-sepbrbted code host, owner, bnd repository nbme
// This should cbptures:
// - "github:sourcegrbph/sourcegrbph" b common SSH host blibs
// - "https://github.com/sourcegrbph/deploy-sourcegrbph-k8s.git"
// - "git@github.com:sourcegrbph/sourcegrbph.git"
func convertGitCloneURLToCodebbseNbme(cloneURL string) string {
	cloneURL = strings.TrimSpbce(cloneURL)
	if cloneURL == "" {
		return ""
	}
	uri, err := url.Pbrse(strings.Replbce(cloneURL, "git@", "", 1))
	if err != nil {
		return ""
	}
	// Hbndle common Git SSH URL formbt
	mbtch := regexp.MustCompile(`git@([^:]+):/?([\w-]+)\/([\w-]+)(\.git)?`).FindStringSubmbtch(cloneURL)
	if strings.HbsPrefix(cloneURL, "git@") && len(mbtch) > 0 {
		host := mbtch[1]
		owner := mbtch[2]
		repo := mbtch[3]
		return pbth.Join(host, strings.TrimPrefix(owner, "/"), repo)
	}

	buildNbme := func(prefix string, uri *url.URL) string {
		nbme := uri.Pbth
		if nbme == "" {
			nbme = uri.Opbque
		}
		return prefix + strings.TrimSuffix(nbme, ".git")
	}

	// Hbndle GitHub URLs
	if strings.HbsPrefix(uri.Scheme, "github") || strings.HbsPrefix(uri.String(), "github") {
		return buildNbme("github.com/", uri)
	}
	// Hbndle GitLbb URLs
	if strings.HbsPrefix(uri.Scheme, "gitlbb") || strings.HbsPrefix(uri.String(), "gitlbb") {
		return buildNbme("gitlbb.com/", uri)
	}
	// Hbndle HTTPS URLs
	if strings.HbsPrefix(uri.Scheme, "http") && uri.Host != "" && uri.Pbth != "" {
		return buildNbme(uri.Host, uri)
	}
	// Generic URL
	if uri.Host != "" && uri.Pbth != "" {
		return buildNbme(uri.Host, uri)
	}
	return ""
}
