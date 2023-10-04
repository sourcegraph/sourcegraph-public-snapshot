package codehost_scenario

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/dev/codehost_scenario/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GitHubClient provides methods for creating and retrieving resources with the configured GitHub codehost.
// It is recommended that the configered token passed to the client has the following scopes:
// - admin:enterprise
// - delete_repo
// - repo
// - site_admin
// - user
// - write:org
type GitHubClient struct {
	t   *testing.T
	cfg *config.GitHub
	c   *github.Client
}

// GetOrg returns the GitHub organization with the given name.
func (gh *GitHubClient) GetOrg(ctx context.Context, name string) (*github.Organization, error) {
	org, resp, err := gh.c.Organizations.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return nil, errors.Newf("failed to find org: %s - %s", name, respErr)
	}
	return org, err
}

// CreateOrg creates a new GitHub organization with the given name using the Admin GitHub API.
func (gh *GitHubClient) CreateOrg(ctx context.Context, name string) (*github.Organization, error) {
	newOrg := github.Organization{
		Login: &name,
	}

	org, resp, err := gh.c.Admin.CreateOrg(ctx, &newOrg, gh.cfg.AdminUser)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return nil, errors.Newf("failed to create org %q - %s", name, respErr)
	}
	return org, err
}

// UpdateOrg updates an existing GitHub organization with the given Org values
func (gh *GitHubClient) UpdateOrg(ctx context.Context, org *github.Organization) (*github.Organization, error) {
	_, resp, err := gh.c.Organizations.Edit(ctx, org.GetLogin(), org)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return nil, errors.Newf("failed to update actions permissions for org %q - %s", org.GetLogin(), respErr)
	}
	return org, err
}

// CreateUser creates a new GitHub user with the given username using the GitHub Admin API
func (gh *GitHubClient) CreateUser(ctx context.Context, name, email string) (*github.User, error) {
	user, resp, err := gh.c.Admin.CreateUser(ctx, name, email)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return nil, errors.Newf("failed to create user %q - %s", name, respErr)
	}
	return user, nil
}

// GetUser returns the GitHub user with the given username.
func (gh *GitHubClient) GetUser(ctx context.Context, name string) (*github.User, error) {
	user, resp, err := gh.c.Users.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return nil, errors.Newf("failed to get user %q - %s", name, respErr)
	}
	return user, nil
}

// DeleteUser deletes a GitHub user with the given username using the GitHub Admin API.
func (gh *GitHubClient) DeleteUser(ctx context.Context, username string) error {
	resp, err := gh.c.Admin.DeleteUser(ctx, username)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return errors.Newf("failed to delete user %q - %s", username, respErr)
	}

	return nil
}

// GetTeam returns the GitHub team with the given name in the given Organization name.
func (gh *GitHubClient) GetTeam(ctx context.Context, org string, name string) (*github.Team, error) {
	team, resp, err := gh.c.Teams.GetTeamBySlug(ctx, org, name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return nil, errors.Newf("failed to get team %q - %s", name, respErr)
	}

	return team, err
}

// CreateTeam creates a new GitHub team with the given name in the given Organization.
func (gh *GitHubClient) CreateTeam(ctx context.Context, org *github.Organization, name string) (*github.Team, error) {
	newTeam := github.NewTeam{
		Name:        name,
		Description: strp("auto created team"),
		Privacy:     strp("closed"),
	}
	team, resp, err := gh.c.Teams.CreateTeam(ctx, org.GetLogin(), newTeam)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return nil, errors.Newf("failed to create team %q - %s", name, respErr)
	}
	return team, err
}

// DeleteTeam deletes a GitHub team with the given name in the given Organization.
func (gh *GitHubClient) DeleteTeam(ctx context.Context, org *github.Organization, name string) error {
	resp, err := gh.c.Teams.DeleteTeamBySlug(ctx, org.GetLogin(), name)
	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return errors.Newf("failed to delete team %q - %s", name, respErr)
	}
	return err
}

// AssignTeamMembership adds team membership for a user in a team
func (gh *GitHubClient) AssignTeamMembership(ctx context.Context, org *github.Organization, team *github.Team, user *github.User) (*github.Team, error) {
	_, resp, err := gh.c.Teams.AddTeamMembershipBySlug(ctx, org.GetLogin(), team.GetSlug(), user.GetLogin(), &github.TeamAddTeamMembershipOptions{
		Role: "member",
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return team, nil
}

// GetRepo returns the GitHub repository with the given name in the given owner which should typically get the Organization name.
func (gh *GitHubClient) GetRepo(ctx context.Context, owner, repoName string) (*github.Repository, error) {
	repo, resp, err := gh.c.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return nil, errors.Newf("failed to get repo %q - %s", repoName, respErr)
	}

	return repo, nil
}

// CreateRepo creates a new GitHub repository with the given name under the given org.
func (gh *GitHubClient) CreateRepo(ctx context.Context, org *github.Organization, repoName string, private bool) (*github.Repository, error) {
	repo, resp, err := gh.c.Repositories.Create(ctx, org.GetLogin(), &github.Repository{
		Name:    &repoName,
		Private: &private,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return nil, errors.Newf("failed to create repo %q - %s", repoName, respErr)
	}

	return repo, err
}

// ForkRepo forks a repository into an Organization. The repostiry will have the same name but the owner will be the given
// organization. Note that only teh default branch is forked.
func (gh *GitHubClient) ForkRepo(ctx context.Context, org *github.Organization, owner, repoName string) error {
	_, resp, err := gh.c.Repositories.CreateFork(ctx, owner, repoName, &github.RepositoryCreateForkOptions{
		Organization:      org.GetLogin(),
		Name:              repoName,
		DefaultBranchOnly: true,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return errors.Newf("failed to fork repo %q - %s", repoName, respErr)
	}
	return nil
}

// UpdateRepo updates an existing GitHub repository with the given repo values
func (gh *GitHubClient) UpdateRepo(ctx context.Context, org *github.Organization, repo *github.Repository) (*github.Repository, error) {
	result, resp, err := gh.c.Repositories.Edit(ctx, org.GetLogin(), repo.GetName(), repo)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return nil, errors.Newf("failed to edit repository %q - %s", repo.GetName(), respErr)
	}

	return result, nil
}

// DeleteTeam deletes a GitHub repo with the given name in the given Organization.
func (gh *GitHubClient) DeleteRepo(ctx context.Context, org *github.Organization, repo *github.Repository) error {
	resp, err := gh.c.Repositories.Delete(ctx, org.GetLogin(), repo.GetName())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respErr := githubResponseError(gh.t, resp)
		return errors.Newf("failed to edit repository %q - %s", repo.GetName(), respErr)
	}

	if err != nil {
		return err
	}

	return nil
}

// UpdateTeamRepoPermissions updates the permissions of the given team for the given repository in the provided Organization.
func (gh *GitHubClient) UpdateTeamRepoPermissions(ctx context.Context, org *github.Organization, team *github.Team, repo *github.Repository) error {
	resp, err := gh.c.Teams.AddTeamRepoByID(ctx, org.GetID(), team.GetID(), org.GetLogin(), repo.GetName(), &github.TeamAddTeamRepoOptions{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		respErr := githubResponseError(gh.t, resp)
		return errors.Newf("failed to update repo %q permissions for team %q: %v", repo.GetName(), team.GetSlug(), respErr)
	}
	return nil
}

func githubResponseError(t *testing.T, resp *github.Response) string {
	code := resp.StatusCode
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Logf("failed to read response body: %v", err)
		return ""
	}
	return fmt.Sprintf("Status Code: %d\nBody: %s\n", code, string(raw))

}

// NewGitHubClient returns a new GitHub client from the given config. Note that the client sets InsecureSkipVerify to true
func NewGitHubClient(ctx context.Context, t *testing.T, cfg config.GitHub) (*GitHubClient, error) {
	t.Helper()
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
