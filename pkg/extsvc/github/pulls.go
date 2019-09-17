package github

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// An Actor represents an object which can take actions on GitHub. Typically a User or Bot.
type Actor struct {
	AvatarURL string
	Login     string
	URL       string
}

// A GitActor represents an actor in a Git commit (ie. an author or committer).
type GitActor struct {
	AvatarURL string
	Email     string
	Name      string
	User      *Actor
}

// A Review of a PullRequest.
type Review struct {
	Body        string
	State       string
	URL         string
	Author      Actor
	Commit      Commit
	CreatedAt   time.Time
	SubmittedAt time.Time
}

// A Commit in a Repository.
type Commit struct {
	OID           string
	Message       string
	URL           string
	Committer     GitActor
	Status        Status
	CommittedDate time.Time
	PushedDate    time.Time
}

// A Status represents a Commit status.
type Status struct {
	Contexts []StatusContext // The individual status contexts for this commit.
	State    string          // The combined commit status.
}

// A StatusContext represents an individual commit status context
type StatusContext struct {
	AvatarURL   string
	Context     string
	Description string
	State       string
	TargetURL   string
	CreatedAt   time.Time
	Creator     Actor
}

// PullRequest is a GitHub pull request.
type PullRequest struct {
	RepoWithOwner string `json:"-"`
	ID            string
	Title         string
	Body          string
	State         string
	URL           string
	Number        int
	Author        Actor
	Participants  []Actor
	Reviews       []Review
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// LoadPullRequests loads a list of PullRequests from Github.
func (c *Client) LoadPullRequests(ctx context.Context, prs ...*PullRequest) error {
	type repository struct {
		Owner string
		Name  string
		PRs   map[string]*PullRequest
	}

	labeled := map[string]*repository{}
	for _, pr := range prs {
		owner, repo, err := SplitRepositoryNameWithOwner(pr.RepoWithOwner)
		if err != nil {
			return err
		}

		repoLabel := owner + "_" + repo
		r, ok := labeled[repoLabel]
		if !ok {
			r = &repository{
				Owner: owner,
				Name:  repo,
				PRs:   map[string]*PullRequest{},
			}
			labeled[repoLabel] = r
		}

		prLabel := repoLabel + "_" + strconv.Itoa(pr.Number)
		r.PRs[prLabel] = pr
	}

	var q strings.Builder
	q.WriteString(`
		fragment actor on Actor { avatarUrl, login, url }
		fragment pr on PullRequest {
		  id, title, body, state, url, number, createdAt, updatedAt
		  author { ...actor }
		  participants(first: 100) { nodes { ...actor } }
		  reviews(first: 100) {
			nodes {
			  body, state, url, createdAt, submittedAt
			  author { ...actor }
			  commit {
				oid, message, committedDate, pushedDate, url
				committer {
				  avatarUrl, email, name
				  user { ...actor }
				}
				status {
				  state
				  contexts {
					avatarUrl, context, description, state, targetUrl, createdAt
					creator { ...actor }
				  }
				}
			  }
			}
		  }
		}
		query {
	`)

	for repoLabel, r := range labeled {
		q.WriteString(fmt.Sprintf("%s: repository(owner: %q, name: %q) {\n",
			repoLabel, r.Owner, r.Name))

		for prLabel, pr := range r.PRs {
			q.WriteString(fmt.Sprintf("%s: pullRequest(number: %d) { ...pr }\n",
				prLabel, pr.Number,
			))
		}

		q.WriteString("}\n")
	}

	q.WriteString("}")

	var results map[string]map[string]*struct {
		PullRequest
		Participants struct{ Nodes []Actor }
		Reviews      struct{ Nodes []Review }
	}

	err := c.requestGraphQL(ctx, "", q.String(), nil, &results)
	if err != nil {
		return err
	}

	for repoLabel, prs := range results {
		for prLabel, pr := range prs {
			pr.PullRequest.Participants = pr.Participants.Nodes
			pr.PullRequest.Reviews = pr.Reviews.Nodes
			*labeled[repoLabel].PRs[prLabel] = pr.PullRequest
		}
	}

	return nil
}
