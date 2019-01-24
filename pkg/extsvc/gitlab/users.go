package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"github.com/peterhellberg/link"
)

type User struct {
	ID         int32      `json:"id"`
	Name       string     `json:"name"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	State      string     `json:"state"`
	AvatarURL  string     `json:"avatar_url"`
	WebURL     string     `json:"web_url"`
	Identities []Identity `json:"identities"`
}

type Identity struct {
	Provider  string `json:"provider"`
	ExternUID string `json:"extern_uid"`
}

func (c *Client) ListUsers(ctx context.Context, urlStr string) (users []*User, nextPageURL *string, err error) {
	if MockListUsers != nil {
		return MockListUsers(c, ctx, urlStr)
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, nil, err
	}
	respHeader, err := c.do(ctx, req, &users)
	if err != nil {
		return nil, nil, err
	}

	// Get URL to next page. See https://docs.gitlab.com/ee/api/README.html#pagination-link-header.
	if l := link.Parse(respHeader.Get("Link"))["next"]; l != nil {
		nextPageURL = &l.URI
	}

	return users, nextPageURL, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	if MockGetUser != nil {
		return MockGetUser(c, ctx, id)
	}

	var urlStr string
	if id == "" {
		urlStr = "user"
	} else {
		urlStr = fmt.Sprintf("users/%s", id)
	}

	var usr User
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	if _, err := c.do(ctx, req, &usr); err != nil {
		return nil, err
	}
	return &usr, nil
}

type UserEmail struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

// ListEmails lists the currently authenticated user's emails (the client must be authenticated).
func (c *Client) ListEmails(ctx context.Context) ([]*UserEmail, error) {
	if MockListEmails != nil {
		return MockListEmails(ctx)
	}

	var emails []*UserEmail
	req, err := http.NewRequest("GET", "user/emails", nil)
	if err != nil {
		return nil, err
	}
	if _, err := c.do(ctx, req, &emails); err != nil {
		return nil, err
	}
	return emails, nil
}
