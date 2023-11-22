package azuredevops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// AbandonPullRequest abandons (closes) the specified PR, returns the updated PR.
func (c *client) AbandonPullRequest(ctx context.Context, args PullRequestCommonArgs) (PullRequest, error) {
	reqURL := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID)}

	abandoned := PullRequestStatusAbandoned
	data, err := json.Marshal(PullRequestUpdateInput{Status: &abandoned})
	if err != nil {
		return PullRequest{}, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("PATCH", reqURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

// CreatePullRequest creates a new PR with the specified properties, returns the newly created PR.
// NOTE: this API needs repository ID specified not repository Name in OrgProjectRepoArgs.
func (c *client) CreatePullRequest(ctx context.Context, args OrgProjectRepoArgs, input CreatePullRequestInput) (PullRequest, error) {
	data, err := json.Marshal(&input)
	if err != nil {
		return PullRequest{}, errors.Wrap(err, "marshalling request")
	}

	reqURL := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests", args.Org, args.Project, args.RepoNameOrID)}

	req, err := http.NewRequest("POST", reqURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

// GetPullRequest gets the specified PR.
func (c *client) GetPullRequest(ctx context.Context, args PullRequestCommonArgs) (PullRequest, error) {
	reqURL := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID)}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

// GetPullRequestStatuses returns the build statuses associated with the specified PR.
func (c *client) GetPullRequestStatuses(ctx context.Context, args PullRequestCommonArgs) ([]PullRequestBuildStatus, error) {
	reqURL := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s/statuses", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID)}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var pr PullRequestStatuses
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return nil, err
	}

	return pr.Value, nil
}

// UpdatePullRequest updates the specified PR, returns the updated PR.
//
// Warning: If you are setting the TargetRefName in the PullRequestUpdateInput, it will be the only thing to get updated (bug in the ADO API).
func (c *client) UpdatePullRequest(ctx context.Context, args PullRequestCommonArgs, input PullRequestUpdateInput) (PullRequest, error) {
	reqURL := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID)}

	data, err := json.Marshal(input)
	if err != nil {
		return PullRequest{}, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("PATCH", reqURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

// CreatePullRequestCommentThread creates a new comment Thread specified PR, returns the updated PR.
func (c *client) CreatePullRequestCommentThread(ctx context.Context, args PullRequestCommonArgs, input PullRequestCommentInput) (PullRequestCommentResponse, error) {
	reqURL := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s/threads", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID)}

	data, err := json.Marshal(input)
	if err != nil {
		return PullRequestCommentResponse{}, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("POST", reqURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return PullRequestCommentResponse{}, err
	}

	var pr PullRequestCommentResponse
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequestCommentResponse{}, err
	}

	return pr, nil
}

// CompletePullRequest completes(merges) the specified PR, returns the updated PR.
func (c *client) CompletePullRequest(ctx context.Context, args PullRequestCommonArgs, input PullRequestCompleteInput) (PullRequest, error) {
	reqURL := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID)}
	completed := PullRequestStatusCompleted
	pri := PullRequestUpdateInput{
		Status:                &completed,
		LastMergeSourceCommit: &PullRequestCommit{CommitID: input.CommitID},
		CompletionOptions:     &PullRequestCompletionOptions{DeleteSourceBranch: input.DeleteSourceBranch},
	}
	if input.MergeStrategy != nil {
		pri.CompletionOptions.MergeStrategy = *input.MergeStrategy
	}

	data, err := json.Marshal(pri)
	if err != nil {
		return PullRequest{}, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("PATCH", reqURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}
