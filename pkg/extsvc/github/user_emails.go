package github

import (
	"context"
)

type UserEmail struct {
	Email      string `json:"email,omitempty"`
	Primary    bool   `json:"primary,omitempty"`
	Verified   bool   `json:"verified,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}

var MockGetAuthenticatedUserEmails func(ctx context.Context, token string) ([]*UserEmail, error)

// GetAuthenticatedUserEmails returns the first 100 emails associated with the currently
// authenticated user.
func (c *Client) GetAuthenticatedUserEmails(ctx context.Context, token string) ([]*UserEmail, error) {
	if MockGetAuthenticatedUserEmails != nil {
		return MockGetAuthenticatedUserEmails(ctx, token)
	}

	var emails []*UserEmail
	err := c.requestGet(ctx, token, "/user/emails?per_page=100", &emails)
	if err != nil {
		return nil, err
	}
	return emails, nil
}
