package bitbucketcloud

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CurrentUser returns the user associated with the authenticator in use.
func (c *client) CurrentUser(ctx context.Context) (*User, error) {
	req, err := http.NewRequest("GET", "/2.0/user", nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var user User
	if err := c.do(ctx, req, &user); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &user, nil
}

type User struct {
	Account
	IsStaff   bool   `json:"is_staff"`
	AccountID string `json:"account_id"`
}

type UserEmail struct {
	Email       string `json:"email"`
	IsConfirmed bool   `json:"is_confirmed"`
	IsPrimary   bool   `json:"is_primary"`
}

func (c *client) CurrentUserEmails(ctx context.Context, pageToken *PageToken) (emails []*UserEmail, next *PageToken, err error) {
	if pageToken.HasMore() {
		next, err = c.reqPage(ctx, pageToken.Next, &emails)
		return
	}

	next, err = c.page(ctx, "/user/emails", nil, pageToken, &emails)
	return
}
