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

// MockGetProject, if non-nil, will be called instead of Client.GetProject
var MockGetProject func(c *Client, ctx context.Context, op GetProjectOp) (*Project, error)

// MockListTree, if non-nil, will be called instead of Client.ListTree
var MockListTree func(c *Client, ctx context.Context, op ListTreeOp) ([]*Tree, error)
