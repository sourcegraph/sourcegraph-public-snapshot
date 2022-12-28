package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/scaletesting/internal/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitLabCodeHost struct {
	def *CodeHostDefinition
	c   *gitlab.Client
}

var _ CodeHostDestination = (*GitLabCodeHost)(nil)

func NewGitLabCodeHost(_ context.Context, def *CodeHostDefinition) (*GitLabCodeHost, error) {
	baseURL, err := url.Parse(def.URL)
	if err != nil {
		return nil, err
	}
	baseURL.Path = "/api/v4"

	gl, err := gitlab.NewClient(def.Token, gitlab.WithBaseURL(baseURL.String()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GitLab client")
	}
	return &GitLabCodeHost{
		def: def,
		c:   gl,
	}, nil
}

// GitOpts returns the git options that should be used when a git command is invoked for GitLab
func (g *GitLabCodeHost) GitOpts() []GitOpt {
	if len(g.def.SSHKey) == 0 {
		return []GitOpt{}
	}

	GitEnv := func(cmd *run.Command) *run.Command {
		return cmd.Environ([]string{fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no'", g.def.SSHKey)})
	}

	return []GitOpt{GitEnv}
}

// AddSSHKey adds the SSH key defined in the code host configuration to
// the current authenticated user. The key that is added is set to expire
// in 7 days and the name of the key is set to "codehost-copy key"
//
// If there is no ssh key defined on the code host configuration this
// is is a noop and returns a 0 for the key ID
func (g *GitLabCodeHost) AddSSHKey(ctx context.Context) (int64, error) {
	if len(g.def.SSHKey) == 0 {
		return 0, nil
	}

	data, err := os.ReadFile(g.def.SSHKey)
	if err != nil {
		return 0, err
	}

	keyData := string(data)
	keyTitle := "codehost-copy key"
	week := 24 * time.Hour * 7
	expireTime := gitlab.ISOTime(time.Now().Add(week))

	sshKey, res, err := g.c.Users.AddSSHKey(&gitlab.AddSSHKeyOptions{
		Title:     &keyTitle,
		Key:       &keyData,
		ExpiresAt: &expireTime,
	}, nil)

	if err != nil {
		return 0, nil
	}

	if res.StatusCode >= 300 {
		return 0, errors.Newf("failed to add ssh key. Got status %d code", res.StatusCode)
	}
	return int64(sshKey.ID), nil
}

// DropSSHKey removes the ssh key by by ID for the current authenticated user. If there is no
// ssh key set on the codehost configuration this method is a noop
func (g *GitLabCodeHost) DropSSHKey(ctx context.Context, keyID int64) error {
	// if there is no ssh key in the code host definition
	// then we have nothing to drop
	if len(g.def.SSHKey) == 0 {
		return nil
	}
	res, err := g.c.Users.DeleteSSHKey(int(keyID), nil)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.Newf("failed to delete key %v. Got status %d code", keyID, res.StatusCode)
	}
	return nil
}

func (g *GitLabCodeHost) InitializeFromState(ctx context.Context, stateRepos []*store.Repo) (int, int, error) {
	return 0, 0, errors.New("not implemented for Gitlab")
}

func (g *GitLabCodeHost) Iterator() Iterator[[]*store.Repo] {
	panic("not implemented")
}

func (g *GitLabCodeHost) ListRepos(ctx context.Context) ([]*store.Repo, error) {
	return nil, errors.New("not implemented for Gitlab")
}

func (g *GitLabCodeHost) CreateRepo(ctx context.Context, name string) (*url.URL, error) {
	groups, _, err := g.c.Groups.ListGroups(&gitlab.ListGroupsOptions{Search: gitlab.String(g.def.Path)})
	if err != nil {
		return nil, err
	}
	if len(groups) < 1 {
		return nil, errors.New("GitLab group not found")
	}
	group := groups[0]

	var resp *gitlab.Response
	var project *gitlab.Project
	err = nil
	retries := 0
	for resp == nil || resp.StatusCode >= 500 {
		project, resp, err = g.c.Projects.CreateProject(&gitlab.CreateProjectOptions{
			Name:        gitlab.String(name),
			NamespaceID: &group.ID,
		})
		retries++
		if retries == 3 && project == nil {
			return nil, errors.Wrapf(err, "Exceeded retry limit while creating repo")
		}
	}
	if err != nil && strings.Contains(err.Error(), "has already been taken") {
		// state does not match reality, get existing repo
		project, _, err = g.c.Projects.GetProject(fmt.Sprintf("%s/%s", group.Name, name), &gitlab.GetProjectOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	gitURL, err := url.Parse(project.WebURL)
	if err != nil {
		return nil, err
	}

	if len(g.def.SSHKey) == 0 {
		gitURL.Scheme = "ssh://"
	} else {
		gitURL.User = url.UserPassword(g.def.Username, g.def.Password)
	}
	gitURL.Path = gitURL.Path + ".git"

	return gitURL, nil
}
