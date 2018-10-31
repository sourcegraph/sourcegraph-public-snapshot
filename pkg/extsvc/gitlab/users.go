package gitlab

import (
	"context"
	"net/http"

	"github.com/peterhellberg/link"
)

type User struct {
	ID         int32      `json:"id"`
	Username   string     `json:"username"`
	Name       string     `json:"name"`
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
		return MockListUsers(ctx, urlStr)
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
