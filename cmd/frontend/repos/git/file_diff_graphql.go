package git

import (
	"context"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type FileDiffConnection []graphqlbackend.FileDiff

func (c FileDiffConnection) Nodes(context.Context) ([]graphqlbackend.FileDiff, error) {
	return []graphqlbackend.FileDiff(c), nil
}

func (c FileDiffConnection) TotalCount(context.Context) (*int32, error) {
	n := int32(len(c))
	return &n, nil
}

func (c FileDiffConnection) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (c FileDiffConnection) DiffStat(context.Context) (graphqlbackend.IDiffStat, error) {
	var stat graphqlbackend.DiffStat
	for _, fileDiff := range c {
		s := fileDiff.Stat()
		stat.Added_ += s.Added()
		stat.Changed_ += s.Changed()
		stat.Deleted_ += s.Deleted()
	}
	return stat, nil
}

func (c FileDiffConnection) RawDiff(context.Context) (string, error) {
	diffs := make([]*diff.FileDiff, len(c))
	for i, d := range c {
		diffs[i] = d.Raw()
	}
	b, err := diff.PrintMultiFileDiff(diffs)
	return string(b), err
}
