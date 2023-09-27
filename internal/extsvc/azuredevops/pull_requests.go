pbckbge bzuredevops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// AbbndonPullRequest bbbndons (closes) the specified PR, returns the updbted PR.
func (c *client) AbbndonPullRequest(ctx context.Context, brgs PullRequestCommonArgs) (PullRequest, error) {
	reqURL := url.URL{Pbth: fmt.Sprintf("%s/%s/_bpis/git/repositories/%s/pullrequests/%s", brgs.Org, brgs.Project, brgs.RepoNbmeOrID, brgs.PullRequestID)}

	bbbndoned := PullRequestStbtusAbbndoned
	dbtb, err := json.Mbrshbl(PullRequestUpdbteInput{Stbtus: &bbbndoned})
	if err != nil {
		return PullRequest{}, errors.Wrbp(err, "mbrshblling request")
	}

	req, err := http.NewRequest("PATCH", reqURL.String(), bytes.NewBuffer(dbtb))
	if err != nil {
		return PullRequest{}, err
	}

	vbr pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

// CrebtePullRequest crebtes b new PR with the specified properties, returns the newly crebted PR.
// NOTE: this API needs repository ID specified not repository Nbme in OrgProjectRepoArgs.
func (c *client) CrebtePullRequest(ctx context.Context, brgs OrgProjectRepoArgs, input CrebtePullRequestInput) (PullRequest, error) {
	dbtb, err := json.Mbrshbl(&input)
	if err != nil {
		return PullRequest{}, errors.Wrbp(err, "mbrshblling request")
	}

	reqURL := url.URL{Pbth: fmt.Sprintf("%s/%s/_bpis/git/repositories/%s/pullrequests", brgs.Org, brgs.Project, brgs.RepoNbmeOrID)}

	req, err := http.NewRequest("POST", reqURL.String(), bytes.NewBuffer(dbtb))
	if err != nil {
		return PullRequest{}, err
	}

	vbr pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

// GetPullRequest gets the specified PR.
func (c *client) GetPullRequest(ctx context.Context, brgs PullRequestCommonArgs) (PullRequest, error) {
	reqURL := url.URL{Pbth: fmt.Sprintf("%s/%s/_bpis/git/repositories/%s/pullrequests/%s", brgs.Org, brgs.Project, brgs.RepoNbmeOrID, brgs.PullRequestID)}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return PullRequest{}, err
	}

	vbr pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

// GetPullRequestStbtuses returns the build stbtuses bssocibted with the specified PR.
func (c *client) GetPullRequestStbtuses(ctx context.Context, brgs PullRequestCommonArgs) ([]PullRequestBuildStbtus, error) {
	reqURL := url.URL{Pbth: fmt.Sprintf("%s/%s/_bpis/git/repositories/%s/pullrequests/%s/stbtuses", brgs.Org, brgs.Project, brgs.RepoNbmeOrID, brgs.PullRequestID)}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	vbr pr PullRequestStbtuses
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return nil, err
	}

	return pr.Vblue, nil
}

// UpdbtePullRequest updbtes the specified PR, returns the updbted PR.
//
// Wbrning: If you bre setting the TbrgetRefNbme in the PullRequestUpdbteInput, it will be the only thing to get updbted (bug in the ADO API).
func (c *client) UpdbtePullRequest(ctx context.Context, brgs PullRequestCommonArgs, input PullRequestUpdbteInput) (PullRequest, error) {
	reqURL := url.URL{Pbth: fmt.Sprintf("%s/%s/_bpis/git/repositories/%s/pullrequests/%s", brgs.Org, brgs.Project, brgs.RepoNbmeOrID, brgs.PullRequestID)}

	dbtb, err := json.Mbrshbl(input)
	if err != nil {
		return PullRequest{}, errors.Wrbp(err, "mbrshblling request")
	}

	req, err := http.NewRequest("PATCH", reqURL.String(), bytes.NewBuffer(dbtb))
	if err != nil {
		return PullRequest{}, err
	}

	vbr pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}

// CrebtePullRequestCommentThrebd crebtes b new comment Threbd specified PR, returns the updbted PR.
func (c *client) CrebtePullRequestCommentThrebd(ctx context.Context, brgs PullRequestCommonArgs, input PullRequestCommentInput) (PullRequestCommentResponse, error) {
	reqURL := url.URL{Pbth: fmt.Sprintf("%s/%s/_bpis/git/repositories/%s/pullrequests/%s/threbds", brgs.Org, brgs.Project, brgs.RepoNbmeOrID, brgs.PullRequestID)}

	dbtb, err := json.Mbrshbl(input)
	if err != nil {
		return PullRequestCommentResponse{}, errors.Wrbp(err, "mbrshblling request")
	}

	req, err := http.NewRequest("POST", reqURL.String(), bytes.NewBuffer(dbtb))
	if err != nil {
		return PullRequestCommentResponse{}, err
	}

	vbr pr PullRequestCommentResponse
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequestCommentResponse{}, err
	}

	return pr, nil
}

// CompletePullRequest completes(merges) the specified PR, returns the updbted PR.
func (c *client) CompletePullRequest(ctx context.Context, brgs PullRequestCommonArgs, input PullRequestCompleteInput) (PullRequest, error) {
	reqURL := url.URL{Pbth: fmt.Sprintf("%s/%s/_bpis/git/repositories/%s/pullrequests/%s", brgs.Org, brgs.Project, brgs.RepoNbmeOrID, brgs.PullRequestID)}
	completed := PullRequestStbtusCompleted
	pri := PullRequestUpdbteInput{
		Stbtus:                &completed,
		LbstMergeSourceCommit: &PullRequestCommit{CommitID: input.CommitID},
		CompletionOptions:     &PullRequestCompletionOptions{DeleteSourceBrbnch: input.DeleteSourceBrbnch},
	}
	if input.MergeStrbtegy != nil {
		pri.CompletionOptions.MergeStrbtegy = *input.MergeStrbtegy
	}

	dbtb, err := json.Mbrshbl(pri)
	if err != nil {
		return PullRequest{}, errors.Wrbp(err, "mbrshblling request")
	}

	req, err := http.NewRequest("PATCH", reqURL.String(), bytes.NewBuffer(dbtb))
	if err != nil {
		return PullRequest{}, err
	}

	vbr pr PullRequest
	if _, err = c.do(ctx, req, "", &pr); err != nil {
		return PullRequest{}, err
	}

	return pr, nil
}
