package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
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
	OID             string
	Message         string
	MessageHeadline string
	URL             string
	Committer       GitActor
	CommittedDate   time.Time
	PushedDate      time.Time
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
	HeadRefOid    string
	BaseRefOid    string
	HeadRefName   string
	BaseRefName   string
	Number        int64
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

// Key is a unique key identifying this event in the context of its pull request.
func (e AssignedEvent) Key() string {
	return fmt.Sprintf("%s:%s:%d", e.Actor.Login, e.Assignee.Login, e.CreatedAt.UnixNano())
}

// ClosedEvent represents a 'closed' event on a PullRequest.
type ClosedEvent struct {
	Actor     Actor
	CreatedAt time.Time
	URL       string
}

// Key is a unique key identifying this event in the context of its pull request.
func (e ClosedEvent) Key() string {
	return fmt.Sprintf("%s:%d", e.Actor.Login, e.CreatedAt.UnixNano())
}

// IssueComment represents a comment on an PullRequest that isn't
// a commit or review comment.
type IssueComment struct {
	DatabaseID          int64
	Author              Actor
	Editor              *Actor
	AuthorAssociation   string
	Body                string
	URL                 string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	IncludesCreatedEdit bool
}

// Key is a unique key identifying this event in the context of its pull request.
func (e IssueComment) Key() string {
	return strconv.FormatInt(e.DatabaseID, 10)
}

// RenamedTitleEvent represents a 'renamed' event on a given pull request.
type RenamedTitleEvent struct {
	Actor         Actor
	PreviousTitle string
	CurrentTitle  string
	CreatedAt     time.Time
}

// Key is a unique key identifying this event in the context of its pull request.
func (e RenamedTitleEvent) Key() string {
	return fmt.Sprintf("%s:%d", e.Actor.Login, e.CreatedAt.UnixNano())
}

// MergedEvent represents a 'merged' event on a given pull request.
type MergedEvent struct {
	Actor        Actor
	MergeRefName string
	URL          string
	Commit       Commit
	CreatedAt    time.Time
}

// Key is a unique key identifying this event in the context of its pull request.
func (e MergedEvent) Key() string {
	return fmt.Sprintf("%s:%d", e.Actor.Login, e.CreatedAt.UnixNano())
}

// PullRequestReview represents a review on a given pull request.
type PullRequestReview struct {
	DatabaseID          int64
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

// Key is a unique key identifying this event in the context of its pull request.
func (e PullRequestReview) Key() string {
	return strconv.FormatInt(e.DatabaseID, 10)
}

// PullRequestReviewThread represents a thread of review comments on a given pull request.
// Since webhooks only send pull request review comment payloads, we normalize
// each thread we receive via GraphQL, and don't store this event as the metadata
// of a ChangesetEvent, instead storing each contained comment as a separate ChangesetEvent.
// That's why this type doesn't have a Key method like the others.
type PullRequestReviewThread struct {
	Comments []*PullRequestReviewComment
}

// PullRequestReviewComment represents a review comment on a given pull request.
type PullRequestReviewComment struct {
	DatabaseID          int64
	Author              Actor
	AuthorAssociation   string
	Editor              Actor
	Commit              Commit
	Body                string
	State               string
	URL                 string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	IncludesCreatedEdit bool
}

// Key is a unique key identifying this event in the context of its pull request.
func (e PullRequestReviewComment) Key() string {
	return strconv.FormatInt(e.DatabaseID, 10)
}

// ReopenedEvent represents a 'reopened' event on a pull request.
type ReopenedEvent struct {
	Actor     Actor
	CreatedAt time.Time
}

// Key is a unique key identifying this event in the context of its pull request.
func (e ReopenedEvent) Key() string {
	return fmt.Sprintf("%s:%d", e.Actor.Login, e.CreatedAt.UnixNano())
}

// ReviewDismissedEvent represents a 'review_dismissed' event on a pull request.
type ReviewDismissedEvent struct {
	Actor            Actor
	Review           PullRequestReview
	DismissalMessage string
	CreatedAt        time.Time
}

// Key is a unique key identifying this event in the context of its pull request.
func (e ReviewDismissedEvent) Key() string {
	return fmt.Sprintf(
		"%s:%d:%d",
		e.Actor.Login,
		e.Review.DatabaseID,
		e.CreatedAt.UnixNano(),
	)
}

// ReviewRequestRemovedEvent represents a 'review_request_removed' event on a
// pull request.
type ReviewRequestRemovedEvent struct {
	Actor             Actor
	RequestedReviewer Actor
	CreatedAt         time.Time
}

// Key is a unique key identifying this event in the context of its pull request.
func (e ReviewRequestRemovedEvent) Key() string {
	return fmt.Sprintf(
		"%s:%s:%d",
		e.Actor.Login,
		e.RequestedReviewer.Login,
		e.CreatedAt.UnixNano(),
	)
}

// ReviewRequestedRevent represents a 'review_requested' event on a
// pull request.
type ReviewRequestedEvent struct {
	Actor             Actor
	RequestedReviewer Actor
	CreatedAt         time.Time
}

// Key is a unique key identifying this event in the context of its pull request.
func (e ReviewRequestedEvent) Key() string {
	return fmt.Sprintf(
		"%s:%s:%d",
		e.Actor.Login,
		e.RequestedReviewer.Login,
		e.CreatedAt.UnixNano(),
	)
}

// UnassignedEvent represents an 'unassigned' event on a pull request.
type UnassignedEvent struct {
	Actor     Actor
	Assignee  Actor
	CreatedAt time.Time
}

// Key is a unique key identifying this event in the context of its pull request.
func (e UnassignedEvent) Key() string {
	return fmt.Sprintf("%s:%s:%d", e.Actor.Login, e.Assignee.Login, e.CreatedAt.UnixNano())
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
	case "PullRequestReviewComment":
		i.Item = new(PullRequestReviewComment)
	case "PullRequestReviewThread":
		i.Item = new(PullRequestReviewThread)
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

type CreatePullRequestInput struct {
	// The Node ID of the repository.
	RepositoryID string `json:"repositoryId"`
	// The name of the branch you want your changes pulled into. This should be
	// an existing branch on the current repository.
	BaseRefName string `json:"baseRefName"`
	// The name of the branch where your changes are implemented.
	HeadRefName string `json:"headRefName"`
	// The title of the pull request.
	Title string `json:"title"`
	// The body of the pull request (optional).
	Body string `json:"body"`
}

// CreatePullRequest creates a PullRequest on Github.
func (c *Client) CreatePullRequest(ctx context.Context, in *CreatePullRequestInput) (*PullRequest, error) {
	var q strings.Builder
	q.WriteString(pullRequestFragments)
	q.WriteString(`mutation	CreatePullRequest($input:CreatePullRequestInput!) {
  createPullRequest(input:$input) {
    pullRequest {
      ... pr
    }
  }
}`)

	var result struct {
		CreatePullRequest struct {
			PullRequest struct {
				PullRequest
				Participants  struct{ Nodes []Actor }
				TimelineItems struct{ Nodes []TimelineItem }
			} `json:"pullRequest"`
		} `json:"createPullRequest"`
	}

	input := map[string]interface{}{"input": in}
	err := c.requestGraphQL(ctx, "", q.String(), input, &result)
	if err != nil {
		if gqlErrs, ok := err.(graphqlErrors); ok && len(gqlErrs) == 1 {
			e := gqlErrs[0]
			if strings.Contains(e.Message, "A pull request already exists for") {
				return nil, ErrPullRequestAlreadyExists
			}
		}
		return nil, err
	}

	pr := &result.CreatePullRequest.PullRequest.PullRequest
	pr.TimelineItems = result.CreatePullRequest.PullRequest.TimelineItems.Nodes
	pr.Participants = result.CreatePullRequest.PullRequest.Participants.Nodes
	return pr, nil
}

// ClosePullRequest closes the PullRequest on Github.
func (c *Client) ClosePullRequest(ctx context.Context, pr *PullRequest) error {
	var q strings.Builder
	q.WriteString(pullRequestFragments)
	q.WriteString(`mutation	ClosePullRequest($input:ClosePullRequestInput!) {
  closePullRequest(input:$input) {
    pullRequest {
      ... pr
    }
  }
}`)

	var result struct {
		ClosePullRequest struct {
			PullRequest struct {
				PullRequest
				Participants  struct{ Nodes []Actor }
				TimelineItems struct{ Nodes []TimelineItem }
			} `json:"pullRequest"`
		} `json:"closePullRequest"`
	}

	input := map[string]interface{}{"input": struct {
		ID string `json:"pullRequestId"`
	}{ID: pr.ID}}
	err := c.requestGraphQL(ctx, "", q.String(), input, &result)
	if err != nil {
		return err
	}

	*pr = result.ClosePullRequest.PullRequest.PullRequest
	pr.TimelineItems = result.ClosePullRequest.PullRequest.TimelineItems.Nodes
	pr.Participants = result.ClosePullRequest.PullRequest.Participants.Nodes

	return nil
}

// LoadPullRequests loads a list of PullRequests from Github.
func (c *Client) LoadPullRequests(ctx context.Context, prs ...*PullRequest) error {
	const batchSize = 15
	// We load prs in batches to avoid hitting Github's GraphQL node limit
	for i := 0; i < len(prs); i += batchSize {
		j := i + batchSize
		if j > len(prs) {
			j = len(prs)
		}
		if err := c.loadPullRequests(ctx, prs[i:j]...); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) loadPullRequests(ctx context.Context, prs ...*PullRequest) error {
	type repository struct {
		Owner string
		Name  string
		PRs   map[string]*PullRequest
	}

	labeled := map[string]*repository{}
	for i, pr := range prs {
		owner, repo, err := SplitRepositoryNameWithOwner(pr.RepoWithOwner)
		if err != nil {
			return err
		}

		repoLabel := fmt.Sprintf("repo_%d", i)
		r, ok := labeled[repoLabel]
		if !ok {
			r = &repository{
				Owner: owner,
				Name:  repo,
				PRs:   map[string]*PullRequest{},
			}
			labeled[repoLabel] = r
		}

		prLabel := repoLabel + "_" + strconv.FormatInt(pr.Number, 10)
		r.PRs[prLabel] = pr
	}

	var q strings.Builder
	q.WriteString(pullRequestFragments)
	q.WriteString("query {\n")

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

// GetOpenPullRequestByRefs fetches the the pull request associated with the supplied
// refs. GitHub only allows one open PR by ref at a time.
// If nothing is found an error is returned.
func (c *Client) GetOpenPullRequestByRefs(ctx context.Context, owner, name, baseRef, headRef string) (*PullRequest, error) {
	var q strings.Builder
	q.WriteString(pullRequestFragments)
	q.WriteString("query {\n")
	q.WriteString(fmt.Sprintf("repository(owner: %q, name: %q) {\n",
		owner, name))
	q.WriteString(fmt.Sprintf("pullRequests(baseRefName: %q, headRefName: %q, first: 1, states: OPEN) { \n",
		git.AbbreviateRef(baseRef), git.AbbreviateRef(headRef),
	))
	q.WriteString(fmt.Sprintf("nodes{ ... pr }\n"))
	q.WriteString("}\n")
	q.WriteString("}\n")
	q.WriteString("}")

	var results struct {
		Repository struct {
			PullRequests struct {
				Nodes []*struct {
					PullRequest
					Participants  struct{ Nodes []Actor }
					TimelineItems struct{ Nodes []TimelineItem }
				}
			}
		}
	}

	err := c.requestGraphQL(ctx, "", q.String(), nil, &results)
	if err != nil {
		return nil, err
	}
	if len(results.Repository.PullRequests.Nodes) != 1 {
		return nil, fmt.Errorf("expected 1 pull request, got %d instead", len(results.Repository.PullRequests.Nodes))
	}

	pr := results.Repository.PullRequests.Nodes[0].PullRequest
	pr.Participants = results.Repository.PullRequests.Nodes[0].Participants.Nodes
	pr.TimelineItems = results.Repository.PullRequests.Nodes[0].TimelineItems.Nodes

	return &pr, nil
}

const pullRequestFragments = `
fragment actor on Actor { avatarUrl, login, url }
fragment commit on Commit {
	oid, message, messageHeadline, committedDate, pushedDate, url
	committer {
	avatarUrl, email, name
	user { ...actor }
	}
}
fragment review on PullRequestReview {
	databaseId
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
	headRefOid, baseRefOid, headRefName, baseRefName
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
		PULL_REQUEST_REVIEW_THREAD
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
		databaseId
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
		... on PullRequestReviewThread {
		comments(last: 100) {
			nodes {
			databaseId
			author { ...actor }
			authorAssociation
			editor { ...actor }
			commit { ...commit }
			body
			state
			url
			createdAt
			updatedAt
			includesCreatedEdit
			}
		}
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
`
