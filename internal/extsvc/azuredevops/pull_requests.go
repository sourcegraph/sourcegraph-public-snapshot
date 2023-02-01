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

// CreatePullRequest creates a new PR with the specified properties, returns the newly created PR.
// NOTE: this API needs repository ID specified not repository Name in OrgProjectRepoArgs.
func (c *Client) CreatePullRequest(ctx context.Context, args OrgProjectRepoArgs, input CreatePullRequestInput) (PullRequest, error) {
	queryParams := make(url.Values)
	queryParams.Set("api-version", apiVersion)

	data, err := json.Marshal(&input)
	if err != nil {
		return PullRequest{}, errors.Wrap(err, "marshalling request")
	}

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests", args.Org, args.Project, args.RepoNameOrID), RawQuery: queryParams.Encode()}

	req, err := http.NewRequest("POST", urlRepositoriesByProjects.String(), bytes.NewBuffer(data))
	if err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

// AbandonPullRequest abandons (closes) the specified PR, returns the updated PR.
func (c *Client) AbandonPullRequest(ctx context.Context, args PullRequestCommonArgs) (PullRequest, error) {
	queryParams := make(url.Values)
	queryParams.Set("api-version", apiVersion)

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID), RawQuery: queryParams.Encode()}

	abandoned := PullRequestStatusAbandoned
	data, err := json.Marshal(PullRequestUpdateInput{Status: &abandoned})
	if err != nil {
		return PullRequest{}, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("PATCH", urlRepositoriesByProjects.String(), bytes.NewBuffer(data))
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
func (c *Client) GetPullRequest(ctx context.Context, args PullRequestCommonArgs) (PullRequest, error) {
	queryParams := make(url.Values)
	queryParams.Set("api-version", apiVersion)

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID), RawQuery: queryParams.Encode()}

	req, err := http.NewRequest("GET", urlRepositoriesByProjects.String(), nil)
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
func (c *Client) GetPullRequestStatuses(ctx context.Context, args PullRequestCommonArgs) ([]PullRequestBuildStatus, error) {
	queryParams := make(url.Values)
	queryParams.Set("api-version", apiVersion)

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s/statuses", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID), RawQuery: queryParams.Encode()}

	req, err := http.NewRequest("GET", urlRepositoriesByProjects.String(), nil)
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
func (c *Client) UpdatePullRequest(ctx context.Context, args PullRequestCommonArgs, input PullRequestUpdateInput) (PullRequest, error) {
	queryParams := make(url.Values)
	queryParams.Set("api-version", apiVersion)

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID), RawQuery: queryParams.Encode()}

	data, err := json.Marshal(input)
	if err != nil {
		return PullRequest{}, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("PATCH", urlRepositoriesByProjects.String(), bytes.NewBuffer(data))
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
func (c *Client) CreatePullRequestCommentThread(ctx context.Context, args PullRequestCommonArgs, input PullRequestCommentInput) (PullRequestCommentResponse, error) {
	queryParams := make(url.Values)
	queryParams.Set("api-version", apiVersion)

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s/threads", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID), RawQuery: queryParams.Encode()}

	data, err := json.Marshal(input)
	if err != nil {
		return PullRequestCommentResponse{}, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("POST", urlRepositoriesByProjects.String(), bytes.NewBuffer(data))
	if err != nil {
		return PullRequestCommentResponse{}, err
	}

	var pr PullRequestCommentResponse
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequestCommentResponse{}, err
	}

	return pr, nil
}

// CompletePullRequest abandons (closes) the specified PR, returns the updated PR. The PullRequestUpdateInput input just needs to specify the LastMergeSourceCommit.ID.
func (c *Client) CompletePullRequest(ctx context.Context, args PullRequestCommonArgs, input PullRequestCommitRef) (PullRequest, error) {
	queryParams := make(url.Values)
	queryParams.Set("api-version", apiVersion)

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/pullrequests/%s", args.Org, args.Project, args.RepoNameOrID, args.PullRequestID), RawQuery: queryParams.Encode()}

	completed := PullRequestStatusCompleted
	data, err := json.Marshal(PullRequestUpdateInput{Status: &completed, LastMergeSourceCommit: &input})
	if err != nil {
		return PullRequest{}, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("PATCH", urlRepositoriesByProjects.String(), bytes.NewBuffer(data))
	if err != nil {
		return PullRequest{}, err
	}

	var pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}
