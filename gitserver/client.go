package gitserver

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func client(ctx context.Context) (gitpb.GitTransportClient, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return gitpb.NewGitTransportClient(cl.Conn), nil
}
