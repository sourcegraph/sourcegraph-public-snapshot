package backend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

var Meta = &meta{}

type meta struct{}

func (s *meta) Config(ctx context.Context) (*sourcegraph.ServerConfig, error) {
	c := &sourcegraph.ServerConfig{
		Version: buildvar.Version,
		AppURL:  conf.AppURL.String(),
	}

	return c, nil
}

type MockMeta struct {
	Config func(v0 context.Context) (*sourcegraph.ServerConfig, error)
}
