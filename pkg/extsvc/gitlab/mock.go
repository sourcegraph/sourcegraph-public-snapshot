package gitlab

import (
	"context"
)

// MockListProjects, if non-nil, will be called instead of every invocation of Client.ListProjects.
var MockListProjects func(c *Client, ctx context.Context, urlStr string) (proj []*Project, nextPageURL *string, err error)

// MockListUsers, if non-nil, will be called instead of Client.ListUsers
var MockListUsers func(c *Client, ctx context.Context, urlStr string) (users []*User, nextPageURL *string, err error)

// MockGetUser, if non-nil, will be called instead of Client.GetUser
var MockGetUser func(c *Client, ctx context.Context, id string) (*User, error)

// MockListEmails, if non-nil, will be called instead of Client.ListEmails
var MockListEmails func(ctx context.Context) ([]*UserEmail, error)
