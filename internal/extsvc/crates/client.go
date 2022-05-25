package crates

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

type Client struct {
	cli httpcli.Doer
}

type File struct {
}

func NewClient(cli httpcli.Doer) *Client {
	return &Client{cli: cli}
}

func (c *Client) Version(ctx context.Context, name string, version string) (*File, error) {
	return nil, nil
}
