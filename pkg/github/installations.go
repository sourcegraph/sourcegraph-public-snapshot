package github

import (
	"context"

	gogithub "github.com/sourcegraph/go-github/github"
)

// ListAllAccessibleInstallations lists all GitHub app installations that are
// accessible via the currently authed user.
func ListAllAccessibleInstallations(ctx context.Context) ([]*gogithub.Installation, error) {
	if ListAllAccessibleInstallationsMock != nil {
		return ListAllAccessibleInstallationsMock(ctx)
	}

	if !HasAuthedUser(ctx) {
		return nil, nil
	}

	const maxPage = 1000
	opt := gogithub.ListOptions{
		PerPage: 100,
	}

	var allInstalls []*gogithub.Installation
	for page := 1; page <= maxPage; page++ {
		opt.Page = page
		installs, resp, err := Client(ctx).Users.ListInstallations(ctx, &opt)
		if err != nil {
			return nil, checkResponse(ctx, resp, err, "github.User.ListInstallations")
		}
		allInstalls = append(allInstalls, installs...)
		if len(installs) < opt.PerPage {
			break
		}
	}
	return allInstalls, nil
}

var ListAllAccessibleInstallationsMock func(ctx context.Context) ([]*gogithub.Installation, error)

func ListAllAccessibleInstallationsMock_Return(returns []*gogithub.Installation) (called *bool) {
	called = new(bool)
	ListAllAccessibleInstallationsMock = func(ctx context.Context) ([]*gogithub.Installation, error) {
		*called = true
		return returns, nil
	}
	return
}

// ListAllAccessibleReposForInstallation lists all GitHub repos for the given
// installation that are accessible via the currently authed user.
func ListAllAccessibleReposForInstallation(ctx context.Context, installID int) ([]*gogithub.Repository, error) {
	if ListAllAccessibleReposForInstallationMock != nil {
		return ListAllAccessibleReposForInstallationMock(ctx, installID)
	}

	if !HasAuthedUser(ctx) {
		return nil, nil
	}

	const maxPage = 1000
	opt := gogithub.ListOptions{
		PerPage: 100,
	}

	var allRepos []*gogithub.Repository
	for page := 1; page <= maxPage; page++ {
		opt.Page = page
		repos, resp, err := Client(ctx).Users.ListInstallationRepos(ctx, installID, &opt)
		if err != nil {
			return nil, checkResponse(ctx, resp, err, "github.User.ListInstallationRepos")
		}
		allRepos = append(allRepos, repos...)
		if len(repos) < opt.PerPage {
			break
		}
	}
	return allRepos, nil
}

var ListAllAccessibleReposForInstallationMock func(ctx context.Context, installID int) ([]*gogithub.Repository, error)

func ListAllAccessibleReposForInstallationMock_Return(returns []*gogithub.Repository) (called *bool) {
	called = new(bool)
	ListAllAccessibleReposForInstallationMock = func(ctx context.Context, installID int) ([]*gogithub.Repository, error) {
		*called = true
		return returns, nil
	}
	return
}
