package bitbucketcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PullRequestInput struct {
	Title        string
	Description  string
	SourceBranch string

	// The following fields are optional.
	//
	// If SourceRepo is provided, only FullName is actually used.
	SourceRepo        *Repo
	DestinationBranch *string
}

var _ json.Marshaler = &PullRequestInput{}

// CreatePullRequest opens a new pull request.
//
// Invoking CreatePullRequest with the same repo and options will succeed: the
// same PR will be returned each time, and will be updated accordingly on
// Bitbucket with any changed information in the options.
func (c *Client) CreatePullRequest(ctx context.Context, repo *Repo, opts PullRequestInput) (*PullRequest, error) {
	data, err := json.Marshal(&opts)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/2.0/repositories/%s/pullrequests", repo.FullName), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var pr PullRequest
	if err := c.do(ctx, req, &pr); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &pr, nil
}

// DeclinePullRequest declines (closes without merging) a pull request.
//
// Invoking DeclinePullRequest on an already declined PR will error.
func (c *Client) DeclinePullRequest(ctx context.Context, repo *Repo, id int64) (*PullRequest, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d/decline", repo.FullName, id), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var pr PullRequest
	if err := c.do(ctx, req, &pr); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &pr, nil
}

// GetPullRequest retrieves a single pull request.
func (c *Client) GetPullRequest(ctx context.Context, repo *Repo, id int64) (*PullRequest, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d", repo.FullName, id), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var pr PullRequest
	if err := c.do(ctx, req, &pr); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &pr, nil
}

// GetPullRequestStatuses retrieves the statuses for a pull request.
//
// Each item in the result set is a *PullRequestStatus.
func (c *Client) GetPullRequestStatuses(repo *Repo, id int64) (*ResultSet, error) {
	u, err := url.Parse(fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d/statuses", repo.FullName, id))
	if err != nil {
		return nil, errors.Wrap(err, "parsing URL")
	}

	return newResultSet(c, u, func(ctx context.Context, req *http.Request) (*PageToken, []interface{}, error) {
		var page struct {
			*PageToken
			Values []*PullRequestStatus `json:"values"`
		}

		if err := c.do(ctx, req, &page); err != nil {
			return nil, nil, err
		}

		values := []interface{}{}
		for _, value := range page.Values {
			values = append(values, value)
		}

		return page.PageToken, values, nil
	}), nil
}

// UpdatePullRequest updates a pull request.
func (c *Client) UpdatePullRequest(ctx context.Context, repo *Repo, id int64, opts PullRequestInput) (*PullRequest, error) {
	data, err := json.Marshal(&opts)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d", repo.FullName, id), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var updated PullRequest
	if err := c.do(ctx, req, &updated); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &updated, nil
}

// CreatePullRequestComment adds a comment to a pull request.
//
// The comment content is expected to be valid Markdown.
func (c *Client) CreatePullRequestComment(ctx context.Context, repo *Repo, id int64, input CommentInput) (*Comment, error) {
	data, err := json.Marshal(&input)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d/comments", repo.FullName, id), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var comment Comment
	if err := c.do(ctx, req, &comment); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &comment, nil
}

type MergePullRequestOpts struct {
	Message           *string        `json:"message,omitempty"`
	CloseSourceBranch *bool          `json:"close_source_branch,omitempty"`
	MergeStrategy     *MergeStrategy `json:"merge_strategy,omitempty"`
}

// MergePullRequest merges the given pull request.
func (c *Client) MergePullRequest(ctx context.Context, repo *Repo, id int64, opts MergePullRequestOpts) (*PullRequest, error) {
	data, err := json.Marshal(&opts)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/2.0/repositories/%s/pullrequests/%d/merge", repo.FullName, id), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var pr PullRequest
	if err := c.do(ctx, req, &pr); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &pr, nil
}

// PullRequest represents a single pull request, as returned by the API.
type PullRequest struct {
	Links             Links                     `json:"links"`
	ID                int64                     `json:"id"`
	Title             string                    `json:"title"`
	Rendered          RenderedPullRequestMarkup `json:"rendered"`
	Summary           RenderedMarkup            `json:"summary"`
	State             PullRequestState          `json:"state"`
	Author            Account                   `json:"author"`
	Source            PullRequestEndpoint       `json:"source"`
	Destination       PullRequestEndpoint       `json:"destination"`
	MergeCommit       *PullRequestCommit        `json:"merge_commit,omitempty"`
	CommentCount      int64                     `json:"comment_count"`
	TaskCount         int64                     `json:"task_count"`
	CloseSourceBranch bool                      `json:"close_source_branch"`
	ClosedBy          *Account                  `json:"account,omitempty"`
	Reason            *string                   `json:"reason,omitempty"`
	CreatedOn         time.Time                 `json:"created_on"`
	UpdatedOn         time.Time                 `json:"updated_on"`
	Reviewers         []Account                 `json:"reviewers"`
	Participants      []Participant             `json:"participants"`
}

type PullRequestBranch struct {
	Name                 string          `json:"name"`
	MergeStrategies      []MergeStrategy `json:"merge_strategies"`
	DefaultMergeStrategy MergeStrategy   `json:"default_merge_strategy"`
}

type PullRequestCommit struct {
	Hash string `json:"hash"`
}

type PullRequestEndpoint struct {
	Repo   Repo              `json:"repository"`
	Branch PullRequestBranch `json:"branch"`
	Commit PullRequestCommit `json:"commit"`
}

type RenderedPullRequestMarkup struct {
	Title       RenderedMarkup `json:"title"`
	Description RenderedMarkup `json:"description"`
	Reason      RenderedMarkup `json:"reason"`
}

type PullRequestStatus struct {
	Links       Links                  `json:"links"`
	UUID        string                 `json:"uuid"`
	Key         string                 `json:"key"`
	RefName     string                 `json:"refname"`
	URL         string                 `json:"url"`
	State       PullRequestStatusState `json:"state"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	CreatedOn   time.Time              `json:"created_on"`
	UpdatedOn   time.Time              `json:"updated_on"`
}

type MergeStrategy string
type PullRequestState string
type PullRequestStatusState string

const (
	MergeStrategyMergeCommit MergeStrategy = "merge_commit"
	MergeStrategySquash      MergeStrategy = "squash"
	MergeStrategyFastForward MergeStrategy = "fast_forward"

	PullRequestStateMerged     PullRequestState = "MERGED"
	PullRequestStateSuperseded PullRequestState = "SUPERSEDED"
	PullRequestStateOpen       PullRequestState = "OPEN"
	PullRequestStateDeclined   PullRequestState = "DECLINED"

	PullRequestStatusStateSuccessful PullRequestStatusState = "SUCCESSFUL"
	PullRequestStatusStateFailed     PullRequestStatusState = "FAILED"
	PullRequestStatusStateInProgress PullRequestStatusState = "INPROGRESS"
	PullRequestStatusStateStopped    PullRequestStatusState = "STOPPED"
)

func (opts *PullRequestInput) MarshalJSON() ([]byte, error) {
	type branch struct {
		Name string `json:"name"`
	}

	type repository struct {
		FullName string `json:"full_name"`
	}

	type source struct {
		Branch     branch      `json:"branch"`
		Repository *repository `json:"repository,omitempty"`
	}

	type request struct {
		Title       string  `json:"title"`
		Description string  `json:"description,omitempty"`
		Source      source  `json:"source"`
		Destination *source `json:"destination,omitempty"`
	}

	req := request{
		Title:       opts.Title,
		Description: opts.Description,
		Source: source{
			Branch: branch{Name: opts.SourceBranch},
		},
	}
	if opts.SourceRepo != nil {
		req.Source.Repository = &repository{
			FullName: opts.SourceRepo.FullName,
		}
	}
	if opts.DestinationBranch != nil {
		req.Destination = &source{
			Branch: branch{Name: *opts.DestinationBranch},
		}
	}

	return json.Marshal(&req)
}

type CommentInput struct {
	Content string
}

var _ json.Marshaler = &CommentInput{}

func (ci *CommentInput) MarshalJSON() ([]byte, error) {
	type content struct {
		Raw string `json:"raw"`
	}
	type comment struct {
		Content content `json:"content"`
	}

	return json.Marshal(&comment{
		Content: content{
			Raw: ci.Content,
		},
	})
}
