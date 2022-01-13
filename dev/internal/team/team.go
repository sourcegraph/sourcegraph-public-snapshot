package team

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/slack-go/slack"
	"golang.org/x/net/context/ctxhttp"
	"gopkg.in/yaml.v3"
)

// TeammateResolver provides an interface to find information about teammates.
//
//go:generate ../../mockgen.sh github.com/sourcegraph/sourcegraph/dev/internal/team -i TeammateResolver -o mock.go
type TeammateResolver interface {
	// ResolveByName tries to resolve a teammate by name
	ResolveByName(ctx context.Context, name string) (*Teammate, error)
	// ResolveByGitHubHandle retrieves the Teammate associated with the given GitHub handle
	ResolveByGitHubHandle(ctx context.Context, handle string) (*Teammate, error)
	// ResolveByCommitAuthor retrieves the Teammate associated with the given commit
	ResolveByCommitAuthor(ctx context.Context, org, repo, commit string) (*Teammate, error)
}

const teamDataURL = "https://raw.githubusercontent.com/sourcegraph/handbook/main/data/team.yml"

type Teammate struct {
	// Key is the key for this teammate in team.yml
	Key string `yaml:"-"`
	// HandbookLink is generated from name
	HandbookLink string `yaml:"-"`

	// Slack data is not available in handbook data, we populate it once in getTeamData
	SlackID       string         `yaml:"-"`
	SlackNickname string         `yaml:"-"`
	SlackTimezone *time.Location `yaml:"-"`

	// Handbook team data fields
	Name        string `yaml:"name"`
	Email       string `yaml:"email"`
	GitHub      string `yaml:"github"`
	Description string `yaml:"description"`
	Location    string `yaml:"location"`
	Role        string `yaml:"role"`
}

type teammateResolver struct {
	slack  *slack.Client
	github *github.Client

	// Access via getTeamData
	cachedTeam     map[string]*Teammate
	cachedTeamOnce sync.Once
}

// NewTeammateResolver instantiates a TeammateResolver for querying teammate data.
//
// The GitHub client and Slack client are optional, but enable certain functions and
// extended teammate data.
func NewTeammateResolver(ghClient *github.Client, slackClient *slack.Client) TeammateResolver {
	return &teammateResolver{
		github: ghClient,
		slack:  slackClient,
	}
}

func (r *teammateResolver) ResolveByCommitAuthor(ctx context.Context, org, repo, commit string) (*Teammate, error) {
	if r.github == nil {
		return nil, fmt.Errorf("GitHub integration disabled")
	}

	resp, _, err := r.github.Repositories.GetCommit(ctx, org, repo, commit, nil)
	if err != nil {
		return nil, fmt.Errorf("GetCommit: %w", err)
	}
	return r.ResolveByGitHubHandle(ctx, resp.Author.GetLogin())
}

func (r *teammateResolver) ResolveByGitHubHandle(ctx context.Context, handle string) (*Teammate, error) {
	team, err := r.getTeamData(ctx)
	if err != nil {
		return nil, fmt.Errorf("getTeamData: %w", err)
	}

	var teammate *Teammate
	for _, tm := range team {
		if tm.GitHub == handle {
			teammate = tm
			break
		}
	}
	if teammate == nil {
		return nil, fmt.Errorf("no teammate with GitHub handle %q", handle)
	}
	return teammate, nil
}

func (r *teammateResolver) ResolveByName(ctx context.Context, name string) (*Teammate, error) {
	team, err := r.getTeamData(ctx)
	if err != nil {
		return nil, fmt.Errorf("getTeamData: %w", err)
	}

	// Generalize name
	name = strings.TrimPrefix(strings.ToLower(name), "@")

	// Try to find an exact match
	for _, tm := range team {
		if strings.ToLower(tm.Name) == name ||
			strings.ToLower(tm.SlackNickname) == name ||
			strings.ToLower(tm.GitHub) == name {
			return tm, nil
		}
	}

	// No user found, try to guess
	candidates := []*Teammate{}
	for _, tm := range team {
		if strings.Contains(strings.ToLower(tm.Name), name) ||
			strings.Contains(strings.ToLower(tm.SlackNickname), name) ||
			strings.Contains(strings.ToLower(tm.GitHub), name) {
			candidates = append(candidates, tm)
		}
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) > 1 {
		candidateNames := []string{}
		for _, c := range candidates {
			candidateNames = append(candidateNames, c.Name)
		}
		return nil, fmt.Errorf("multiple users found for name %q: %s", name, strings.Join(candidateNames, ", "))
	}

	return nil, fmt.Errorf("no users found matching name %q", name)
}

func (r *teammateResolver) getTeamData(ctx context.Context) (map[string]*Teammate, error) {
	var onceErr error
	r.cachedTeamOnce.Do(func() {
		team, err := fetchTeamData(ctx)
		if err != nil {
			onceErr = fmt.Errorf("fetchTeamData: %w", err)
			return
		}

		emails := map[string]*Teammate{}
		for _, tm := range team {
			// Create team keyed by email for populating Slack details
			if tm.Email != "" {
				emails[tm.Email] = tm
			}

			// Generate handbook link
			anchor := strings.ToLower(strings.ReplaceAll(tm.Name, " ", "-"))
			anchor = strings.ReplaceAll(anchor, "\"", "")
			tm.HandbookLink = fmt.Sprintf("https://handbook.sourcegraph.com/team#%s", anchor)
		}

		// Populate Slack details
		if r.slack != nil {
			slackUsers, err := r.slack.GetUsers()
			if err != nil {
				onceErr = fmt.Errorf("slack.GetUsers: %w", err)
				return
			}
			for _, user := range slackUsers {
				if teammate, exists := emails[user.Profile.Email]; exists {
					teammate.SlackID = user.ID
					teammate.SlackNickname = user.Name
					teammate.SlackTimezone, err = time.LoadLocation(user.TZ)
					if err != nil {
						onceErr = fmt.Errorf("teammate %q: time.LoadLocation: %w", teammate.Key, err)
						return
					}
				}
			}
		}

		r.cachedTeam = team
	})
	return r.cachedTeam, onceErr
}

func fetchTeamData(ctx context.Context) (map[string]*Teammate, error) {
	resp, err := ctxhttp.Get(ctx, http.DefaultClient, teamDataURL)
	if err != nil {
		return nil, fmt.Errorf("Get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}

	team := map[string]*Teammate{}
	if err = yaml.Unmarshal(body, &team); err != nil {
		return nil, fmt.Errorf("Unmarshal: %w", err)
	}
	for id, tm := range team {
		tm.Key = id
	}

	return team, nil
}
