package git

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type GitCommitConnection []*graphqlbackend.GitCommitResolver

func (c GitCommitConnection) Nodes(context.Context) ([]*graphqlbackend.GitCommitResolver, error) {
	return []*graphqlbackend.GitCommitResolver(c), nil
}

func (c GitCommitConnection) TotalCount(context.Context) (*int32, error) {
	n := int32(len(c))
	return &n, nil
}

func (c GitCommitConnection) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
