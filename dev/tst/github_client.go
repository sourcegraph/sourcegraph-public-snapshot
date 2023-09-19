package tst

import (
	"context"
	"io"
	"os"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/dev/tst/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubClient struct {
	cfg *config.GitHub
	c   *github.Client
}

func (gh *GitHubClient) CreateOrg(ctx context.Context, name string) (*github.Organization, error) {
	newOrg := github.Organization{
		Login: &name,
	}

	org, resp, err := gh.c.Admin.CreateOrg(ctx, &newOrg, gh.cfg.User)
	if resp.StatusCode >= 400 {
		io.Copy(os.Stdout, resp.Body)
		return nil, errors.Newf("failed to create org %q - GitHub status code %d: %v", name, resp.StatusCode, err)
	}
	return org, err
}

func (gh *GitHubClient) UpdateOrg(ctx context.Context, org *github.Organization) (*github.Organization, error) {
	_, resp, err := gh.c.Organizations.Edit(ctx, org.GetLogin(), org)
	if resp.StatusCode >= 400 {
		io.Copy(os.Stdout, resp.Body)
		return nil, errors.Newf("failed to update actions permissions for org %q - GitHub status code %d: %v", org.GetLogin(), resp.StatusCode, err)
	}
	return org, err
}

func (gh *GitHubClient) orgUsers(ctx context.Context, org *github.Organization) ([]*github.User, error) {
	users, _, err := gh.c.Organizations.ListMembers(ctx, org.GetLogin(), &github.ListMembersOptions{})
	if err != nil {
		return nil, err
	}
	return users, nil
}
