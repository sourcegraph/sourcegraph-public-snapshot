package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

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
	return graphql.ID(g.address)
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
