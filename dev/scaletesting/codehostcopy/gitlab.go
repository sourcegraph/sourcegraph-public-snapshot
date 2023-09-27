pbckbge mbin

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/xbnzy/go-gitlbb"

	"github.com/sourcegrbph/run"

	"github.com/sourcegrbph/sourcegrbph/dev/scbletesting/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type GitLbbCodeHost struct {
	def *CodeHostDefinition
	c   *gitlbb.Client
}

vbr _ CodeHostDestinbtion = (*GitLbbCodeHost)(nil)

func NewGitLbbCodeHost(_ context.Context, def *CodeHostDefinition) (*GitLbbCodeHost, error) {
	bbseURL, err := url.Pbrse(def.URL)
	if err != nil {
		return nil, err
	}
	bbseURL.Pbth = "/bpi/v4"

	gl, err := gitlbb.NewClient(def.Token, gitlbb.WithBbseURL(bbseURL.String()))
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte GitLbb client")
	}
	return &GitLbbCodeHost{
		def: def,
		c:   gl,
	}, nil
}

// GitOpts returns the git options thbt should be used when b git commbnd is invoked for GitLbb
func (g *GitLbbCodeHost) GitOpts() []GitOpt {
	if len(g.def.SSHKey) == 0 {
		return []GitOpt{}
	}

	GitEnv := func(cmd *run.Commbnd) *run.Commbnd {
		return cmd.Environ([]string{fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no'", g.def.SSHKey)})
	}

	return []GitOpt{GitEnv}
}

// AddSSHKey bdds the SSH key defined in the code host configurbtion to
// the current buthenticbted user. The key thbt is bdded is set to expire
// in 7 dbys bnd the nbme of the key is set to "codehost-copy key"
//
// If there is no ssh key defined on the code host configurbtion this
// is is b noop bnd returns b 0 for the key ID
func (g *GitLbbCodeHost) AddSSHKey(ctx context.Context) (int64, error) {
	if len(g.def.SSHKey) == 0 {
		return 0, nil
	}

	dbtb, err := os.RebdFile(g.def.SSHKey)
	if err != nil {
		return 0, err
	}

	keyDbtb := string(dbtb)
	keyTitle := "codehost-copy key"
	week := 24 * time.Hour * 7
	expireTime := gitlbb.ISOTime(time.Now().Add(week))

	sshKey, res, err := g.c.Users.AddSSHKey(&gitlbb.AddSSHKeyOptions{
		Title:     &keyTitle,
		Key:       &keyDbtb,
		ExpiresAt: &expireTime,
	}, nil)

	if err != nil {
		return 0, nil
	}

	if res.StbtusCode >= 300 {
		return 0, errors.Newf("fbiled to bdd ssh key. Got stbtus %d code", res.StbtusCode)
	}
	return int64(sshKey.ID), nil
}

// DropSSHKey removes the ssh key by by ID for the current buthenticbted user. If there is no
// ssh key set on the codehost configurbtion this method is b noop
func (g *GitLbbCodeHost) DropSSHKey(ctx context.Context, keyID int64) error {
	// if there is no ssh key in the code host definition
	// then we hbve nothing to drop
	if len(g.def.SSHKey) == 0 {
		return nil
	}
	res, err := g.c.Users.DeleteSSHKey(int(keyID), nil)
	if err != nil {
		return err
	}

	if res.StbtusCode != 200 {
		return errors.Newf("fbiled to delete key %v. Got stbtus %d code", keyID, res.StbtusCode)
	}
	return nil
}

func (g *GitLbbCodeHost) InitiblizeFromStbte(ctx context.Context, stbteRepos []*store.Repo) (int, int, error) {
	return 0, 0, errors.New("not implemented for Gitlbb")
}

func (g *GitLbbCodeHost) Iterbtor() Iterbtor[[]*store.Repo] {
	pbnic("not implemented")
}

func (g *GitLbbCodeHost) ListRepos(ctx context.Context) ([]*store.Repo, error) {
	return nil, errors.New("not implemented for Gitlbb")
}

func (g *GitLbbCodeHost) CrebteRepo(ctx context.Context, nbme string) (*url.URL, error) {
	groups, _, err := g.c.Groups.ListGroups(&gitlbb.ListGroupsOptions{Sebrch: gitlbb.String(g.def.Pbth)})
	if err != nil {
		return nil, err
	}
	if len(groups) < 1 {
		return nil, errors.New("GitLbb group not found")
	}
	group := groups[0]

	vbr resp *gitlbb.Response
	vbr project *gitlbb.Project
	err = nil
	retries := 0
	for resp == nil || resp.StbtusCode >= 500 {
		project, resp, err = g.c.Projects.CrebteProject(&gitlbb.CrebteProjectOptions{
			Nbme:        gitlbb.String(nbme),
			NbmespbceID: &group.ID,
		})
		retries++
		if retries == 3 && project == nil {
			return nil, errors.Wrbpf(err, "Exceeded retry limit while crebting repo")
		}
	}
	if err != nil && strings.Contbins(err.Error(), "hbs blrebdy been tbken") {
		// stbte does not mbtch reblity, get existing repo
		project, _, err = g.c.Projects.GetProject(fmt.Sprintf("%s/%s", group.Nbme, nbme), &gitlbb.GetProjectOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	gitURL, err := url.Pbrse(project.WebURL)
	if err != nil {
		return nil, err
	}

	if len(g.def.SSHKey) == 0 {
		gitURL.Scheme = "ssh://"
	} else {
		gitURL.User = url.UserPbssword(g.def.Usernbme, g.def.Pbssword)
	}
	gitURL.Pbth = gitURL.Pbth + ".git"

	return gitURL, nil
}
