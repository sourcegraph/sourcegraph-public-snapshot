package codehost_scenario

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"os"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/dev/codehost_scenario/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubClient struct {
	cfg *config.GitHub
	c   *github.Client
}

func (gh *GitHubClient) GetOrg(ctx context.Context, name string) (*github.Organization, error) {
	org, resp, err := gh.c.Organizations.Get(ctx, name)
	if resp.StatusCode >= 299 {
		return nil, errors.Newf("failed to find org: %s", name)
	} else if err != nil {
		return nil, err
	}
	return org, err
}

func (gh *GitHubClient) CreateOrg(ctx context.Context, name string) (*github.Organization, error) {
	newOrg := github.Organization{
		Login: &name,
	}

	org, resp, err := gh.c.Admin.CreateOrg(ctx, &newOrg, gh.cfg.AdminUser)
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

func (gh *GitHubClient) CreateUser(ctx context.Context, name, email string) (*github.User, error) {
	user, resp, err := gh.c.Admin.CreateUser(ctx, name, email)
	if resp.StatusCode >= 400 {
		return nil, errors.Newf("failed to create user %q - GitHub status code %d: %v", name, resp.StatusCode, err)
	}
	return user, nil
}

func (gh *GitHubClient) GetUser(ctx context.Context, name string) (*github.User, error) {
	user, resp, err := gh.c.Users.Get(ctx, name)
	if resp.StatusCode >= 400 {
		return nil, errors.Newf("failed to get user %q - GitHub status code %d: %v", name, resp.StatusCode, err)
	}
	return user, nil
}

func (gh *GitHubClient) DeleteUser(ctx context.Context, username string) error {
	resp, err := gh.c.Admin.DeleteUser(ctx, username)
	if resp.StatusCode >= 400 {
		return errors.Newf("failed to delete user %q - GitHub status code %d: %v", username, resp.StatusCode, err)
	}

	return nil
}

func (gh *GitHubClient) GetTeam(ctx context.Context, org string, name string) (*github.Team, error) {
	team, resp, err := gh.c.Teams.GetTeamBySlug(ctx, org, name)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, errors.Newf("server error[%d]: %v", resp.StatusCode, err)
	}

	return team, err
}

func (gh *GitHubClient) CreateTeam(ctx context.Context, org *github.Organization, name string) (*github.Team, error) {
	newTeam := github.NewTeam{
		Name:        name,
		Description: strp("auto created team"),
		Privacy:     strp("closed"),
	}
	team, _, err := gh.c.Teams.CreateTeam(ctx, org.GetLogin(), newTeam)
	return team, err
}

func (gh *GitHubClient) DeleteTeam(ctx context.Context, org *github.Organization, name string) error {
	resp, err := gh.c.Teams.DeleteTeamBySlug(ctx, org.GetLogin(), name)
	if resp.StatusCode >= 400 {
		return errors.Newf("failed to delete team %q - GitHub status code %d: %v", name, resp.StatusCode, err)
	}
	return err
}

func (gh *GitHubClient) AssignTeamMembership(ctx context.Context, org *github.Organization, team *github.Team, user *github.User) (*github.Team, error) {
	_, _, err := gh.c.Teams.AddTeamMembershipBySlug(ctx, org.GetLogin(), team.GetSlug(), user.GetLogin(), &github.TeamAddTeamMembershipOptions{
		Role: "member",
	})

	if err != nil {
		return nil, err
	}

	return team, nil
}

func (gh *GitHubClient) GetRepo(ctx context.Context, owner, repoName string) (*github.Repository, error) {
	repo, resp, err := gh.c.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, errors.Newf("failed to get repo %q - GitHub response code %d: err", repoName, resp.StatusCode, err)
	}

	return repo, nil
}

func (gh *GitHubClient) CreateRepo(ctx context.Context, org *github.Organization, repoName string, private bool) (*github.Repository, error) {
	repo, resp, err := gh.c.Repositories.Create(ctx, org.GetLogin(), &github.Repository{
		Name:    &repoName,
		Private: &private,
	})

	if resp.StatusCode >= 400 {
		return nil, errors.Newf("failed to create repo %q - GitHub response code %d: err", repoName, resp.StatusCode, err)
	}

	return repo, err
}

func (gh *GitHubClient) ForkRepo(ctx context.Context, org *github.Organization, owner, repoName string) error {
	_, resp, err := gh.c.Repositories.CreateFork(ctx, owner, repoName, &github.RepositoryCreateForkOptions{
		Organization:      org.GetLogin(),
		Name:              repoName,
		DefaultBranchOnly: true,
	})
	if resp.StatusCode == 202 {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func (gh *GitHubClient) UpdateRepo(ctx context.Context, org *github.Organization, repo *github.Repository) (*github.Repository, error) {
	result, resp, err := gh.c.Repositories.Edit(ctx, org.GetLogin(), repo.GetName(), repo)

	if resp.StatusCode >= 400 {
		return nil, errors.Newf("failed to edit repository %q - github response status code %d: %v", repo.GetName(), resp.StatusCode, err)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (gh *GitHubClient) DeleteRepo(ctx context.Context, org *github.Organization, repo *github.Repository) error {
	resp, err := gh.c.Repositories.Delete(ctx, org.GetLogin(), repo.GetName())

	if resp.StatusCode >= 400 {
		return errors.Newf("failed to edit repository %q - github response status code %d: %v", repo.GetName(), resp.StatusCode, err)
	}

	if err != nil {
		return err
	}

	return nil
}

func (gh *GitHubClient) UpdateTeamRepoPermissions(ctx context.Context, org *github.Organization, team *github.Team, repo *github.Repository) error {
	resp, err := gh.c.Teams.AddTeamRepoByID(ctx, org.GetID(), team.GetID(), org.GetLogin(), repo.GetName(), &github.TeamAddTeamRepoOptions{})
	if resp.StatusCode != 204 {
		return errors.Newf("failed to update repo %q permissions for team %q: %v", repo.GetName(), team.GetSlug(), err)
	}
	return nil
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
