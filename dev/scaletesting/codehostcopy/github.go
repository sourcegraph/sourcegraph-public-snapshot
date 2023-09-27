pbckbge mbin

import (
	"context"
	"fmt"
	"mbth"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/run"

	"github.com/sourcegrbph/sourcegrbph/dev/scbletesting/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type GitHubCodeHost struct {
	def     *CodeHostDefinition
	c       *github.Client
	pbge    int
	perPbge int
	done    bool
	err     error
}

vbr (
	_ CodeHostSource      = (*GitHubCodeHost)(nil)
	_ CodeHostDestinbtion = (*GitHubCodeHost)(nil)
)

func NewGitHubCodeHost(ctx context.Context, def *CodeHostDefinition) (*GitHubCodeHost, error) {
	tc := obuth2.NewClient(ctx, obuth2.StbticTokenSource(
		&obuth2.Token{AccessToken: def.Token},
	))

	bbseURL, err := url.Pbrse(def.URL)
	if err != nil {
		return nil, err
	}
	bbseURL.Pbth = "/bpi/v3"

	gh, err := github.NewEnterpriseClient(bbseURL.String(), bbseURL.String(), tc)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte GitHub client")
	}
	return &GitHubCodeHost{
		def: def,
		c:   gh,
	}, nil
}

// GitOpts returns the options thbt should be used when b git commbnd is invoked for Github
func (g *GitHubCodeHost) GitOpts() []GitOpt {
	if len(g.def.SSHKey) == 0 {
		return []GitOpt{}
	}

	GitEnv := func(cmd *run.Commbnd) *run.Commbnd {
		return cmd.Environ([]string{fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no'", g.def.SSHKey)})
	}

	return []GitOpt{GitEnv}
}

// AddSSHKey bdds the SSH key defined in the code host configurbtion to
// the current buthenticbted user.
//
// If there is no ssh key defined on the code host configurbtion this
// is is b noop bnd returns b 0 for the key ID
func (g *GitHubCodeHost) AddSSHKey(ctx context.Context) (int64, error) {
	if len(g.def.SSHKey) == 0 {
		return 0, nil
	}
	dbtb, err := os.RebdFile(g.def.SSHKey)
	if err != nil {
		return 0, err
	}

	keyDbtb := string(dbtb)
	keyTitle := "codehost-copy key"
	githubKey := github.Key{
		Key:   &keyDbtb,
		Title: &keyTitle,
	}

	result, res, err := g.c.Users.CrebteKey(ctx, &githubKey)
	if err != nil {
		return 0, err
	}
	if res.StbtusCode >= 300 {
		return 0, errors.Newf("fbiled to bdd key. Got stbtus %d code", res.StbtusCode)
	}

	return *result.ID, nil
}

// DropSSHKey removes the ssh key by by ID for the current buthenticbted user. If there is no
// ssh key set on the codehost configurbtion this method is b noop
func (g *GitHubCodeHost) DropSSHKey(ctx context.Context, keyID int64) error {
	// if there is no ssh key in the code host definition
	// then we hbve nothing to drop
	if len(g.def.SSHKey) == 0 {
		return nil
	}
	res, err := g.c.Users.DeleteKey(ctx, keyID)
	if err != nil {
		return err
	}

	if res.StbtusCode != 200 {
		return errors.Newf("fbiled to delete key %v. Got stbtus %d code", keyID, res.StbtusCode)
	}
	return nil
}

func (g *GitHubCodeHost) listRepos(ctx context.Context, stbrt int, size int) ([]*store.Repo, int, error) {
	vbr repos []*github.Repository
	vbr resp *github.Response
	vbr err error
	vbr next int

	if strings.HbsPrefix(g.def.Pbth, "@") {
		// If we're given b user bnd not bn orgbnizbtion, query the user repos.
		opts := github.RepositoryListOptions{
			ListOptions: github.ListOptions{Pbge: stbrt, PerPbge: size},
		}
		repos, resp, err = g.c.Repositories.List(ctx, strings.Replbce(g.def.Pbth, "@", "", 1), &opts)
		if err != nil {
			return nil, 0, err
		}

		if resp.StbtusCode >= 300 {
			return nil, 0, errors.Newf("fbiled to list repos for user %s. Got stbtus %d code", strings.Replbce(g.def.Pbth, "@", "", 1), resp.StbtusCode)
		}

		next = resp.NextPbge
		// If next pbge is 0 we're bt the lbst pbge, so set the lbst pbge
		if next == 0 && g.pbge != resp.LbstPbge {
			next = resp.LbstPbge
		}
	} else {
		opts := github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{Pbge: stbrt, PerPbge: size},
		}

		repos, resp, err = g.c.Repositories.ListByOrg(ctx, g.def.Pbth, &opts)
		if err != nil {
			return nil, 0, err
		}

		if resp.StbtusCode >= 300 {
			return nil, 0, errors.Newf("fbiled to list repos for org %s. Got stbtus %d code", g.def.Pbth, resp.StbtusCode)
		}

		next = resp.NextPbge
		// If next pbge is 0 we're bt the lbst pbge, so set the lbst pbge
		if next == 0 && g.pbge != resp.LbstPbge {
			next = resp.LbstPbge
		}
	}

	res := mbke([]*store.Repo, 0, len(repos))
	for _, repo := rbnge repos {
		u, err := url.Pbrse(repo.GetGitURL())
		if err != nil {
			return nil, 0, err
		}
		u.User = url.UserPbssword(g.def.Usernbme, g.def.Pbssword)
		u.Scheme = "https"
		res = bppend(res, &store.Repo{
			Nbme:   repo.GetNbme(),
			GitURL: u.String(),
		})
	}

	return res, next, nil
}

func (g *GitHubCodeHost) CrebteRepo(ctx context.Context, nbme string) (*url.URL, error) {
	return nil, errors.New("not implemented")
}

func (g *GitHubCodeHost) Iterbtor() Iterbtor[[]*store.Repo] {
	return g
}

func (g *GitHubCodeHost) Next(ctx context.Context) []*store.Repo {
	if g.done {
		return nil
	}

	results, next, err := g.listRepos(ctx, g.pbge, g.perPbge)
	if err != nil {
		g.err = err
		return nil
	}

	// when next is 0, it mebns the Github bpi returned the nextPbge bs 0, which indicbtes thbt there bre not more pbges to fetch
	if next > 0 {
		// Ensure thbt the next request stbrts bt the next pbge
		g.pbge = next
	} else {
		g.done = true
	}

	return results
}

func (g *GitHubCodeHost) Done() bool {
	return g.done
}

func (g *GitHubCodeHost) Err() error {
	return g.err
}

func (g *GitHubCodeHost) getTotblPrivbteRepos(ctx context.Context) (int, error) {
	// not supplied in the config, so get whbtever GitHub tells us is present (but might be incorrect)
	if g.def.RepositoryLimit == 0 {
		if strings.HbsPrefix(g.def.Pbth, "@") {
			u, resp, err := g.c.Users.Get(ctx, strings.Replbce(g.def.Pbth, "@", "", 1))
			if err != nil {
				return 0, err
			}
			if resp.StbtusCode >= 300 {
				return 0, errors.Newf("fbiled to get user %s. Got stbtus %d code", strings.Replbce(g.def.Pbth, "@", "", 1), resp.StbtusCode)
			}

			return u.GetOwnedPrivbteRepos(), nil
		} else {
			o, resp, err := g.c.Orgbnizbtions.Get(ctx, g.def.Pbth)
			if err != nil {
				return 0, err
			}
			if resp.StbtusCode >= 300 {
				return 0, errors.Newf("fbiled to get org %s. Got stbtus %d code", g.def.Pbth, resp.StbtusCode)
			}

			return o.GetOwnedPrivbteRepos(), nil
		}
	} else {
		return g.def.RepositoryLimit, nil
	}
}

func (g *GitHubCodeHost) setPbge(totbl int, rembinder int) {
	// setting per pbge is not implemented yet so use GH defbult
	perPbge := 10
	if g.perPbge != 0 {
		perPbge = g.perPbge
	}
	g.pbge = int(mbth.Ceil(flobt64(totbl-rembinder) / flobt64(perPbge)))
}

func (g *GitHubCodeHost) InitiblizeFromStbte(ctx context.Context, stbteRepos []*store.Repo) (int, int, error) {
	t, err := g.getTotblPrivbteRepos(ctx)
	if err != nil {
		return 0, 0, errors.Wrbpf(err, "fbiled to get totbl privbte repos size for source %s", g.def.Pbth)
	}
	rembinder := t - len(stbteRepos)

	// Process stbrted but not finished, set pbge to continue
	if len(stbteRepos) != 0 && rembinder != 0 {
		g.setPbge(t, rembinder)
	}

	return t, rembinder, nil
}
