package backend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/buildvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sqs/pbtypes"
)

var Meta sourcegraph.MetaServer = &meta{}

type meta struct{}

var _ sourcegraph.MetaServer = (*meta)(nil)

func (s *meta) Config(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.ServerConfig, error) {
	c := &sourcegraph.ServerConfig{
		Version: buildvar.Version,
		AppURL:  conf.AppURL(ctx).String(),
		IDKey:   idkey.FromContext(ctx).ID,
	}

	return c, nil
}
