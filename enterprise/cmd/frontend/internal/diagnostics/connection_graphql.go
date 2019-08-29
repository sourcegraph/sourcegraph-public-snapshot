package diagnostics

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type ConstConnection []graphqlbackend.Diagnostic

func (c ConstConnection) Nodes(context.Context) ([]graphqlbackend.Diagnostic, error) {
	return []graphqlbackend.Diagnostic(c), nil
}

func (c ConstConnection) TotalCount(context.Context) (int32, error) { return int32(len(c)), nil }

func (c ConstConnection) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
