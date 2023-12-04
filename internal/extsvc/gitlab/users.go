package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

// AuthUser represents a GitLab user for authentication it's slightly different from User
// as this has the CreatedAt field. This object is used for handling authenticating users,
// so that we can check the creation time of the account.
type AuthUser struct {
	ID         int32      `json:"id"`
	Name       string     `json:"name"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	State      string     `json:"state"`
	AvatarURL  string     `json:"avatar_url"`
	WebURL     string     `json:"web_url"`
	Identities []Identity `json:"identities"`
	CreatedAt  time.Time  `json:"created_at,omitempty"`
}

type Identity struct {
	Provider  string `json:"provider"`
	ExternUID string `json:"extern_uid"`
}

func (c *Client) ListUsers(ctx context.Context, urlStr string) (users []*AuthUser, nextPageURL *string, err error) {
	if MockListUsers != nil {
		return MockListUsers(c, ctx, urlStr)
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, nil, err
	}
	respHeader, _, err := c.do(ctx, req, &users)
	if err != nil {
		return nil, nil, err
	}

	// Get URL to next page. See https://docs.gitlab.com/ee/api/README.html#pagination-link-header.
	if l := link.Parse(respHeader.Get("Link"))["next"]; l != nil {
		nextPageURL = &l.URI
	}

	return users, nextPageURL, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*AuthUser, error) {
	if MockGetUser != nil {
		return MockGetUser(c, ctx, id)
	}

	var urlStr string
	if id == "" {
		urlStr = "user"
	} else {
		urlStr = fmt.Sprintf("users/%s", id)
	}

	var usr AuthUser
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	if _, _, err := c.do(ctx, req, &usr); err != nil {
		return nil, err
	}
	return &usr, nil
}
