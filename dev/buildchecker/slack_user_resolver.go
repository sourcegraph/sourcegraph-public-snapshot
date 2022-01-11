package main

import (
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v41/github"
	"github.com/slack-go/slack"
	"golang.org/x/net/context/ctxhttp"
	"gopkg.in/yaml.v2"
)

type SlackUserResolver interface {
	ResolveByCommit(ctx context.Context, commit string) (string, error)
}

const teamDataURL = "https://raw.githubusercontent.com/sourcegraph/handbook/main/data/team.yml"

type teamMember struct {
	Email  string `yaml:"email"`
	GitHub string `yaml:"github"`
}

type githubSlackUserResolver struct {
	ghClient     *github.Client
	slackClient  *slack.Client
	organization string
	repository   string
	team         map[string]teamMember
	sync.Once
}

func NewGithubSlackUserResolver(ghClient *github.Client, slackClient *slack.Client, organization string, repository string) SlackUserResolver {
	return &githubSlackUserResolver{
		ghClient:     ghClient,
		slackClient:  slackClient,
		organization: organization,
		repository:   repository,
	}
}

func (r *githubSlackUserResolver) ResolveByCommit(ctx context.Context, commit string) (string, error) {
	resp, _, err := r.ghClient.Repositories.GetCommit(ctx, r.organization, r.repository, commit, nil)
	if err != nil {
		return "", errors.Wrap(err, "cannot resolve author from commit")
	}
	return r.getSlackUserIDbyCommit(ctx, resp.Author.GetLogin())
}

func (r *githubSlackUserResolver) getSlackUserIDbyCommit(ctx context.Context, handle string) (string, error) {
	err := r.fetchTeamData(ctx)
	if err != nil {
		return "", err
	}
	var email string
	for _, member := range r.team {
		if member.GitHub == handle {
			email = member.Email
			break
		}
	}
	if email == "" {
		return "", errors.Newf("cannot find slack user for GitHub handle %s", handle)
	}
	user, err := r.slackClient.GetUserByEmail(email)
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

func (r *githubSlackUserResolver) fetchTeamData(ctx context.Context) error {
	var outerErr error
	r.Once.Do(func() {
		team, err := fetchTeamData(ctx)
		if err != nil {
			outerErr = err
			return
		}
		r.team = team
	})
	return outerErr
}

func fetchTeamData(ctx context.Context) (map[string]teamMember, error) {
	resp, err := ctxhttp.Get(ctx, http.DefaultClient, teamDataURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	team := map[string]teamMember{}
	err = yaml.Unmarshal(body, &team)
	if err != nil {
		return nil, err
	}
	return team, nil
}

type mockSlackUserResolver struct {
	commit string
	err    error
}

func (r *mockSlackUserResolver) ResolveByCommit(_ context.Context, commit string) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return r.commit, nil
}

func NewMockSlackUserResolver(commit string, err error) SlackUserResolver {
	return &mockSlackUserResolver{
		commit: commit,
		err:    err,
	}
}
