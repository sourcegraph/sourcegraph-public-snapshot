package gerrit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (c *client) GetChange(ctx context.Context, changeID string) (*Change, error) {
	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID))
	if err != nil {
		return nil, err
	}
	reqURL := url.URL{Path: pathStr}
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var change Change
	resp, err := c.do(ctx, req, &change)
	if err != nil {
		// This is a fringe scenario where Gerrit has multiple changes with the same Change ID, we want
		// to pass back a unique error explicitly.
		if strings.Contains(err.Error(), "changes found for") {
			return nil, MultipleChangesError{ID: changeID}
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return &change, nil
}

// AbandonChange abandons a Gerrit change.
func (c *client) AbandonChange(ctx context.Context, changeID string) (*Change, error) {
	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID), "abandon")
	if err != nil {
		return nil, err
	}
	reqURL := url.URL{Path: pathStr}
	req, err := http.NewRequest("POST", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var change Change
	resp, err := c.do(ctx, req, &change)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &change, nil
}

// DeleteChange permanently deletes a Gerrit change.
func (c *client) DeleteChange(ctx context.Context, changeID string) error {
	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID))
	if err != nil {
		return err
	}
	reqURL := url.URL{Path: pathStr}
	req, err := http.NewRequest("DELETE", reqURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.do(ctx, req, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// SubmitChange submits a Gerrit change.
func (c *client) SubmitChange(ctx context.Context, changeID string) (*Change, error) {
	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID), "submit")
	if err != nil {
		return nil, err
	}
	reqURL := url.URL{Path: pathStr}
	req, err := http.NewRequest("POST", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var change Change
	resp, err := c.do(ctx, req, &change)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &change, nil
}

// RestoreChange restores a closed Gerrit change.
func (c *client) RestoreChange(ctx context.Context, changeID string) (*Change, error) {
	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID), "restore")
	if err != nil {
		return nil, err
	}
	reqURL := url.URL{Path: pathStr}
	req, err := http.NewRequest("POST", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var change Change
	resp, err := c.do(ctx, req, &change)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &change, nil
}

// SetReadyForReview sets the change status as ready for review.
func (c *client) SetReadyForReview(ctx context.Context, changeID string) error {
	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID), "ready")
	if err != nil {
		return err
	}
	reqURL := url.URL{Path: pathStr}
	req, err := http.NewRequest("POST", reqURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.do(ctx, req, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// SetWIP sets the change status as WIP (draft).
func (c *client) SetWIP(ctx context.Context, changeID string) error {
	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID), "wip")
	if err != nil {
		return err
	}
	reqURL := url.URL{Path: pathStr}
	req, err := http.NewRequest("POST", reqURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := c.do(ctx, req, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// WriteReviewComment writes a review comment on a Gerrit change.
func (c *client) WriteReviewComment(ctx context.Context, changeID string, comment ChangeReviewComment) error {
	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID), "revisions/current/review")
	if err != nil {
		return err
	}
	reqURL := url.URL{Path: pathStr}
	data, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", reqURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain; charset=UTF-8")

	resp, err := c.do(ctx, req, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetChangeReviews gets the list of reviewrs/reviews for the change.
func (c *client) GetChangeReviews(ctx context.Context, changeID string) (*[]Reviewer, error) {
	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID), "revisions/current/reviewers")
	if err != nil {
		return nil, err
	}
	reqURL := url.URL{Path: pathStr}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain; charset=UTF-8")

	var reviewers []Reviewer
	resp, err := c.do(ctx, req, &reviewers)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &reviewers, nil
}

// MoveChange moves a Gerrit change to a different destination branch.
func (c *client) MoveChange(ctx context.Context, changeID string, input MoveChangePayload) (*Change, error) {

	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID), "move")
	if err != nil {
		return nil, err
	}

	reqURL := url.URL{Path: pathStr}

	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", reqURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var change Change
	resp, err := c.do(ctx, req, &change)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return &change, nil
}

// SetCommitMessage changes the commit message of a Gerrit change.
func (c *client) SetCommitMessage(ctx context.Context, changeID string, input SetCommitMessagePayload) error {

	pathStr, err := url.JoinPath("a/changes", url.PathEscape(changeID), "message")
	if err != nil {
		return err
	}
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}

	reqURL := url.URL{Path: pathStr}
	req, err := http.NewRequest("PUT", reqURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(ctx, req, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
