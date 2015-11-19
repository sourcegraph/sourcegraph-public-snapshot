package gitserver

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb"
)

func client(ctx context.Context) gitpb.GitTransportClient {
	cl := sourcegraph.NewClientFromContext(ctx)
	return gitpb.NewGitTransportClient(cl.Conn)
}
