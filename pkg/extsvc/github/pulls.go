package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
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
	User      *Actor `json:"User,omitempty"`
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
	Status        *Status `json:"Status,omitempty"`
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
	TimelineItems []TimelineItem
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AssignedEvent represents an 'assigned' event on a PullRequest.
type AssignedEvent struct {
	Actor     Actor
	Assignee  Actor
	CreatedAt time.Time
}

// ClosedEvent represents a 'closed' event on a PullRequest.
type ClosedEvent struct {
	Actor     Actor
	CreatedAt time.Time
	URL       string
}

// IssueComment represents a comment on an PullRequest that isn't
// a commit or review comment.
type IssueComment struct {
	Author              Actor
	Editor              *Actor
	AuthorAssociation   string
	Body                string
	URL                 string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	IncludesCreatedEdit bool
}

// RenamedTitleEvent represents a 'renamed' event on a given pull request.
type RenamedTitleEvent struct {
	Actor         Actor
	PreviousTitle string
	CurrentTitle  string
	CreatedAt     time.Time
}

// MergedEvent represents a 'merged' event on a given pull request.
type MergedEvent struct {
	Actor        Actor
	MergeRefName string
	URL          string
	Commit       Commit
	CreatedAt    time.Time
}

// PullRequestReview represents a review on a given pull request.
type PullRequestReview struct {
	Author              Actor
	AuthorAssociation   string
	Body                string
	State               string
	URL                 string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Commit              Commit
	IncludesCreatedEdit bool
}

// ReopenedEvent represents a 'reopened' event on a pull request.
type ReopenedEvent struct {
	Actor     Actor
	CreatedAt time.Time
}

// ReviewDismissedEvent represents a 'review_dismissed' event on a pull request.
type ReviewDismissedEvent struct {
	Actor            Actor
	Review           PullRequestReview
	DismissalMessage string
	CreatedAt        time.Time
}

// ReviewRequestRemovedEvent represents a 'review_request_removed' event on a
// pull request.
type ReviewRequestRemovedEvent struct {
	Actor             Actor
	RequestedReviewer Actor
	CreatedAt         time.Time
}

// ReviewRequestedRevent represents a 'review_requested' event on a
// pull request.
type ReviewRequestedEvent struct {
	Actor             Actor
	RequestedReviewer Actor
	CreatedAt         time.Time
}

// UnassignedEvent represents an 'unassigned' event on a pull request.
type UnassignedEvent struct {
	Actor     Actor
	Assignee  Actor
	CreatedAt time.Time
}

// TimelineItem is a union type of all supported pull request timeline items.
type TimelineItem struct {
	Type string
	Item interface{}
}

// UnmarshalJSON knows how to unmarshal a TimelineItem as produced
// by json.Marshal or as returned by the GitHub GraphQL API.
func (i *TimelineItem) UnmarshalJSON(data []byte) error {
	v := struct {
		Typename *string `json:"__typename"`
		Type     *string
		Item     json.RawMessage
	}{
		Typename: &i.Type,
		Type:     &i.Type,
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch i.Type {
	case "AssignedEvent":
		i.Item = new(AssignedEvent)
	case "ClosedEvent":
		i.Item = new(ClosedEvent)
	case "IssueComment":
		i.Item = new(IssueComment)
	case "RenamedTitleEvent":
		i.Item = new(RenamedTitleEvent)
	case "MergedEvent":
		i.Item = new(MergedEvent)
	case "PullRequestReview":
		i.Item = new(PullRequestReview)
	case "ReopenedEvent":
		i.Item = new(ReopenedEvent)
	case "ReviewDismissedEvent":
		i.Item = new(ReviewDismissedEvent)
	case "ReviewRequestRemovedEvent":
		i.Item = new(ReviewRequestRemovedEvent)
	case "ReviewRequestedEvent":
		i.Item = new(ReviewRequestedEvent)
	case "UnassignedEvent":
		i.Item = new(UnassignedEvent)
	default:
		return errors.Errorf("unknown timeline item type %q", i.Type)
	}

	if len(v.Item) > 0 {
		data = []byte(v.Item)
	}

	return json.Unmarshal(data, i.Item)
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
    fragment commit on Commit {
      oid, message, messageHeadline, committedDate, pushedDate, url
      committer {
        avatarUrl, email, name
        user { ...actor }
      }
    }
    fragment review on PullRequestReview {
      author { ...actor }
      authorAssociation
      body
      state
      url
      createdAt
      updatedAt
      commit { ...commit }
      includesCreatedEdit
    }
    fragment pr on PullRequest {
      id, title, body, state, url, number, createdAt, updatedAt
      author { ...actor }
      participants(first: 100) { nodes { ...actor } }
      timelineItems(
        first: 250
        itemTypes: [
          ASSIGNED_EVENT
          CLOSED_EVENT
          ISSUE_COMMENT
          RENAMED_TITLE_EVENT
          MERGED_EVENT
          PULL_REQUEST_REVIEW
          REOPENED_EVENT
          REVIEW_DISMISSED_EVENT
          REVIEW_REQUEST_REMOVED_EVENT
          REVIEW_REQUESTED_EVENT
          UNASSIGNED_EVENT
        ]
      ) {
        nodes {
          __typename
          ... on AssignedEvent {
            actor { ...actor }
            assignee { ...actor }
            createdAt
          }
          ... on ClosedEvent {
            actor { ...actor }
            createdAt
            url
          }
          ... on IssueComment {
            author { ...actor }
            authorAssociation
            body
            createdAt
            editor { ...actor }
            url
            updatedAt
            includesCreatedEdit
            publishedAt
          }
          ... on RenamedTitleEvent {
            actor { ...actor }
            previousTitle
            currentTitle
            createdAt
          }
          ... on MergedEvent {
            actor { ...actor }
            mergeRefName
            url
            commit { ...commit }
            createdAt
          }
          ... on PullRequestReview {
            ...review
          }
          ... on ReopenedEvent {
            actor { ...actor }
            createdAt
          }
          ... on ReviewDismissedEvent {
            actor { ...actor }
            review { ...review }
            dismissalMessage
            createdAt
          }
          ... on ReviewRequestRemovedEvent {
            actor { ...actor }
            requestedReviewer { ...actor }
            createdAt
          }
          ... on ReviewRequestedEvent {
            actor { ...actor }
            requestedReviewer { ...actor }
            createdAt
          }
          ... on UnassignedEvent {
            actor { ...actor }
            assignee { ...actor }
            createdAt
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
		Participants  struct{ Nodes []Actor }
		TimelineItems struct{ Nodes []TimelineItem }
	}

	err := c.requestGraphQL(ctx, "", q.String(), nil, &results)
	if err != nil {
		return err
	}

	for repoLabel, prs := range results {
		for prLabel, pr := range prs {
			pr.PullRequest.Participants = pr.Participants.Nodes
			pr.PullRequest.TimelineItems = pr.TimelineItems.Nodes
			*labeled[repoLabel].PRs[prLabel] = pr.PullRequest
		}
	}

	return nil
}
