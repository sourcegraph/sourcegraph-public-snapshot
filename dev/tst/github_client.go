package tst

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"os"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/dev/tst/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubClient struct {
	cfg *config.GitHub
	c   *github.Client
}

func NewGitHubClient(ctx context.Context, cfg config.GitHub) (*GitHubClient, error) {
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	))

	tc.Transport.(*oauth2.Transport).Base = http.DefaultTransport
	tc.Transport.(*oauth2.Transport).Base.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	gh, err := github.NewEnterpriseClient(cfg.URL, cfg.URL, tc)
	if err != nil {
		return nil, err
	}

	c := GitHubClient{
		cfg: &cfg,
		c:   gh,
	}
	return &c, nil
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
