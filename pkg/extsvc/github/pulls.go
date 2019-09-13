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

// PullRequest is a GitHub pull request.
type PullRequest struct {
	ID           string
	Title        string
	Body         string
	URL          string
	Number       int
	Author       Actor
	Participants []Actor
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GetPullRequest gets a PullRequest of the given repo with the given id.
func (c *Client) GetPullRequest(ctx context.Context, repoWithOwner string, number int) (*PullRequest, error) {
	owner, repo, err := SplitRepositoryNameWithOwner(repoWithOwner)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`query {
	  repository(owner: %q, name: %q) {
		pullRequest(number: %d) {
		  id
		  title
		  body
		  state
		  url
		  number
		  author {
		    avatarUrl
			login
			url
		  }
		  participants(first: 100) {
			nodes {
			  avatarUrl
			  login
			  url
			}
		  }
		  createdAt
		  updatedAt
		}
	  }
	}`, owner, repo, number)

	var pr PullRequest
	var r struct {
		Repository struct {
			PullRequest struct {
				*PullRequest
				Participants struct {
					Nodes *[]Actor
				}
			}
		}
	}

	r.Repository.PullRequest.PullRequest = &pr
	r.Repository.PullRequest.Participants.Nodes = &pr.Participants

	err = c.requestGraphQL(ctx, "", query, nil, &r)
	if err != nil {
		return nil, err
	}

	return &pr, nil
}
