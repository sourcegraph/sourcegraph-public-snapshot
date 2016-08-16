package gitserver

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func client(ctx context.Context) (sourcegraph.ReposClient, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return sourcegraph.NewReposClient(cl.Conn), nil
}
