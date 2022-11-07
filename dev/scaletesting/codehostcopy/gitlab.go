package main

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/xanzy/go-gitlab"
)

type GitLabCodeHost struct {
	def *CodeHostDefinition
	c   *gitlab.Client
}

var _ CodeHostDestination = (*GitLabCodeHost)(nil)

func NewGitLabCodeHost(ctx context.Context, def *CodeHostDefinition) (*GitLabCodeHost, error) {
	baseURL, err := url.Parse(def.URL)
	if err != nil {
		return nil, err
	}
	baseURL.Path = "/api/v4"

	gl, err := gitlab.NewClient(def.Token, gitlab.WithBaseURL(baseURL.String()))
	if err != nil {
		return nil, err
	}
	return &GitLabCodeHost{
		def: def,
		c:   gl,
	}, nil
}

func (g *GitLabCodeHost) GitOpts() []GitOpt {
	return []GitOpt{}
}

func (g *GitLabCodeHost) AddSSHKey(ctx context.Context) (int64, error) {
	return 0, errors.New("Not implemented")
}

func (g *GitLabCodeHost) DropSSHKey(ctx context.Context, keyID int64) error {
	return errors.New("Not implemented")
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

	project, _, err := g.c.Projects.CreateProject(&gitlab.CreateProjectOptions{
		Name:        gitlab.String(name),
		NamespaceID: &group.ID,
	})
	if err != nil {
		return nil, err
	}

	gitURL, err := url.Parse(project.WebURL)
	if err != nil {
		return nil, err
	}

	gitURL.User = url.UserPassword(g.def.Username, g.def.Password)
	gitURL.Path = gitURL.Path + ".git"

	return gitURL, nil
}
