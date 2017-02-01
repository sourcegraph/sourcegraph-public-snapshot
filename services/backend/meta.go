package backend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var Meta = &meta{}

type meta struct{}

func (s *meta) Config(ctx context.Context) (*sourcegraph.ServerConfig, error) {
	c := &sourcegraph.ServerConfig{
		Version: env.Version,
		AppURL:  conf.AppURL.String(),
	}

	return c, nil
}

type MockMeta struct {
	Config func(v0 context.Context) (*sourcegraph.ServerConfig, error)
}
