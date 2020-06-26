package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

type MergeRequestState string

const (
	MergeRequestStateOpened MergeRequestState = "opened"
	MergeRequestStateClosed MergeRequestState = "closed"
	MergeRequestStateLocked MergeRequestState = "locked"
	MergeRequestStateMerged MergeRequestState = "merged"
)

type MergeRequest struct {
	ID           int64             `json:"id"`
	IID          int64             `json:"iid"`
	ProjectID    int64             `json:"project_id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	State        MergeRequestState `json:"state"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	MergedAt     *time.Time        `json:"merged_at"`
	ClosedAt     *time.Time        `json:"closed_at"`
	Labels       []string          `json:"labels"`
	Pipeline     *Pipeline         `json:"pipeline"`
	SourceBranch string            `json:"source_branch"`
	TargetBranch string            `json:"target_branch"`
	WebURL       string            `json:"web_url"`

	DiffRefs struct {
		BaseSHA  string `json:"base_sha"`
		HeadSHA  string `json:"head_sha"`
		StartSHA string `json:"start_sha"`
	} `json:"diff_refs"`

	// TODO: other fields at
	// https://docs.gitlab.com/ee/api/merge_requests.html#create-mr as needed.
}

type Pipeline struct {
	ID     int64          `json:"id"`
	SHA    string         `json:"sha"`
	Ref    string         `json:"ref"`
	Status PipelineStatus `json:"status"`
	WebURL string         `json:"web_url"`
}

type PipelineStatus string

const (
	PipelineStatusRunning  PipelineStatus = "running"
	PipelineStatusPending  PipelineStatus = "pending"
	PipelineStatusSuccess  PipelineStatus = "success"
	PipelineStatusFailed   PipelineStatus = "failed"
	PipelineStatusCanceled PipelineStatus = "canceled"
	PipelineStatusSkipped  PipelineStatus = "skipped"
	PipelineStatusCreated  PipelineStatus = "created"
	PipelineStatusManual   PipelineStatus = "manual"
)

var (
	ErrMergeRequestAlreadyExists = errors.New("merge request already exists")
	ErrMergeRequestNotFound      = errors.New("merge request not found")
	ErrTooManyMergeRequests      = errors.New("retrieved too many merge requests")
)

type CreateMergeRequestOpts struct {
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	// TODO: other fields at
	// https://docs.gitlab.com/ee/api/merge_requests.html#create-mr as needed.
}

func (c *Client) CreateMergeRequest(ctx context.Context, project *Project, opts CreateMergeRequestOpts) (*MergeRequest, error) {
	data, err := json.Marshal(opts)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling options")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("projects/%d/merge_requests", project.ID), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "creating request to create a merge request")
	}

	resp := &MergeRequest{}
	if _, code, err := c.do(ctx, req, resp); err != nil {
		if code == http.StatusConflict {
			return nil, ErrMergeRequestAlreadyExists
		}

		return nil, errors.Wrap(err, "sending request to create a merge request")
	}

	return resp, nil
}

func (c *Client) GetMergeRequest(ctx context.Context, project *Project, iid int64) (*MergeRequest, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("projects/%d/merge_requests/%d", project.ID, iid), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request to get a merge request")
	}

	resp := &MergeRequest{}
	if _, _, err := c.do(ctx, req, resp); err != nil {
		return nil, errors.Wrap(err, "sending request to get a merge request")
	}

	return resp, nil
}

func (c *Client) GetOpenMergeRequestByRefs(ctx context.Context, project *Project, source, target string) (*MergeRequest, error) {
	values := make(url.Values)
	// Since we're only expecting one merge request, we don't need to enumerate
	// the full list of merge requests if more than one matches: just the
	// existence of a second merge request is sufficient for us to return an
	// error from this function.
	values.Add("per_page", "2")
	values.Add("state", "opened")
	values.Add("source_branch", source)
	values.Add("target_branch", target)
	// The list endpoint doesn't return the full set of fields that we get from
	// the create and get single endpoints, and we need some of those fields
	// (specifically, diff_refs), so we'll just get the minimal set of fields
	// necessary from the list endpoint and then call the get endpoint to flesh
	// out the response.
	values.Add("view", "simple")
	u := &url.URL{
		Path: fmt.Sprintf("projects/%d/merge_requests", project.ID), RawQuery: values.Encode(),
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request to get merge request by refs")
	}

	resp := []*MergeRequest{}
	if _, _, err := c.do(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "sending request to get merge request by refs")
	}

	if len(resp) > 1 {
		return nil, ErrTooManyMergeRequests
	} else if len(resp) == 0 {
		return nil, ErrMergeRequestNotFound
	}

	return c.GetMergeRequest(ctx, project, resp[0].IID)
}

type UpdateMergeRequestOpts struct {
	TargetBranch string                       `json:"target_branch"`
	Title        string                       `json:"title"`
	Description  string                       `json:"description,omitempty"`
	StateEvent   UpdateMergeRequestStateEvent `json:"state_event,omitempty"`
	// TODO: other fields at
	// https://docs.gitlab.com/ee/api/merge_requests.html#update-mr as needed.
}

type UpdateMergeRequestStateEvent string

const (
	UpdateMergeRequestStateEventClose     UpdateMergeRequestStateEvent = "close"
	UpdateMergeRequestStateEventReopen    UpdateMergeRequestStateEvent = "reopen"
	UpdateMergeRequestStateEventUnchanged UpdateMergeRequestStateEvent = ""
)

func (c *Client) UpdateMergeRequest(ctx context.Context, project *Project, mr *MergeRequest, opts UpdateMergeRequestOpts) (*MergeRequest, error) {
	data, err := json.Marshal(opts)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling options")
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("projects/%d/merge_requests/%d", project.ID, mr.IID), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "creating request to update a merge request")
	}

	resp := &MergeRequest{}
	if _, _, err := c.do(ctx, req, resp); err != nil {
		return nil, errors.Wrap(err, "sending request to update a merge request")
	}

	return resp, nil
}
