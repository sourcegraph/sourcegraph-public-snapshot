package gerrit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func (c *client) GetChange(ctx context.Context, changeID string) (*Change, error) {
	reqURL := url.URL{Path: fmt.Sprintf("a/changes/%s", changeID)}
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var change Change
	resp, err := c.do(ctx, req, &change)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &change, nil
}

// AbandonChange abandons a Gerrit change.
func (c *client) AbandonChange(ctx context.Context, changeID string) (*Change, error) {
	reqURL := url.URL{Path: fmt.Sprintf("a/changes/%s/abandon", changeID)}
	req, err := http.NewRequest("POST", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var change Change
	resp, err := c.do(ctx, req, &change)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &change, nil
}

// SubmitChange submits a Gerrit change.
func (c *client) SubmitChange(ctx context.Context, changeID string) (*Change, error) {
	reqURL := url.URL{Path: fmt.Sprintf("a/changes/%s/submit", changeID)}
	req, err := http.NewRequest("POST", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var change Change
	resp, err := c.do(ctx, req, &change)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &change, nil
}

// WriteReviewComment writes a review comment on a Gerrit change.
func (c *client) WriteReviewComment(ctx context.Context, changeID string, comment ChangeReviewComment) error {
	reqURL := url.URL{Path: fmt.Sprintf("a/changes/%s/revisions/current/review", changeID)}
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
