package github

import (
	"context"
	"fmt"
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
	ID           string
	Title        string
	Body         string
	State        string
	URL          string
	Number       int
	Author       Actor
	Participants []Actor
	Reviews      []Review
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GetPullRequest gets a PullRequest of the given repo with the given id.
func (c *Client) GetPullRequest(ctx context.Context, repoWithOwner string, number int) (*PullRequest, error) {
	owner, repo, err := SplitRepositoryNameWithOwner(repoWithOwner)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		fragment actor on Actor { avatarUrl, login, url }
		query {
		  repository(owner: %q, name: %q) {
			pullRequest(number: %d) {
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
		  }
		}
	`, owner, repo, number)

	var pr PullRequest
	var r struct {
		Repository struct {
			PullRequest struct {
				*PullRequest
				Participants struct{ Nodes *[]Actor }
				Reviews      struct{ Nodes *[]Review }
			}
		}
	}

	r.Repository.PullRequest.PullRequest = &pr
	r.Repository.PullRequest.Participants.Nodes = &pr.Participants
	r.Repository.PullRequest.Reviews.Nodes = &pr.Reviews

	err = c.requestGraphQL(ctx, "", query, nil, &r)
	if err != nil {
		return nil, err
	}

	return &pr, nil
}
