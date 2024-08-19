package bitbucket

import (
	"net/http"

	"github.com/dghubble/sling"
)

const bitbucketAPI = "https://bitbucket.org/api/2.0/"

// User is a Bitbucket user.
type User struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Website     string `json:"website"`
	Location    string `json:"location"`
	Type        string `json:"type"` // user, team
}

// client is a Bitbucket client for obtaining a User.
type client struct {
	sling *sling.Sling
}

// newClient returns a new Bitbucket client.
func newClient(httpClient *http.Client) *client {
	base := sling.New().Client(httpClient).Base(bitbucketAPI)
	return &client{
		sling: base,
	}
}

// CurrentUser gets the current user's profile information.
// https://confluence.atlassian.com/bitbucket/users-endpoint-423626336.html
func (c *client) CurrentUser() (*User, *http.Response, error) {
	user := new(User)
	resp, err := c.sling.New().Get("user").ReceiveSuccess(user)
	return user, resp, err
}
