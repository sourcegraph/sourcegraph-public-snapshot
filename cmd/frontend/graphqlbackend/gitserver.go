package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/auth"
)

const gitserverIDKind = "Gitserver"

func marshalGitserverID(id string) graphql.ID { return relay.MarshalID(gitserverIDKind, id) }

func UnmarshalGitserverID(id graphql.ID) (gitserverID int32, err error) {
	err = relay.UnmarshalSpec(id, &gitserverID)
	return
}

func (r *schemaResolver) gitserverByID(ctx context.Context, id graphql.ID) (*gitserverResolver, error) {
	// 🚨 SECURITY: Only site admins can query role permissions or all permissions.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &gitserverResolver{}, nil
}

func (r *schemaResolver) Gitservers(ctx context.Context) ([]*gitserverResolver, error) {
	infos, err := r.gitserverClient.SystemInfo(ctx)
	if err != nil {
		return nil, err
	}

	var resolvers = make([]*gitserverResolver, 0, len(infos))
	for _, info := range infos {
		resolvers = append(resolvers, &gitserverResolver{
			address:             info.Address,
			freeDiskSpaceBytes:  info.FreeSpace,
			totalDiskSpaceBytes: info.TotalSpace,
		})
	}
	return resolvers, nil
}

type gitserverResolver struct {
	address             string
	freeDiskSpaceBytes  uint64
	totalDiskSpaceBytes uint64
}

func (g *gitserverResolver) ID() graphql.ID {
	return marshalGitserverID(g.address)
}

func (g *gitserverResolver) Address() string {
	return g.address
}

func (g *gitserverResolver) FreeDiskSpaceBytes() BigInt {
	return BigInt(g.freeDiskSpaceBytes)
}

func (g *gitserverResolver) TotalDiskSpaceBytes() BigInt {
	return BigInt(g.totalDiskSpaceBytes)
}
