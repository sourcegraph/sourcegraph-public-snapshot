package gitserver

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func client(ctx context.Context) (sourcegraph.ReposClient, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return sourcegraph.NewReposClient(cl.Conn), nil
}
