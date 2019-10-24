package graphqlbackend

import (
	"context"
	"net/url"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type LSIFDumpsListOptions struct {
	Repository string
}

func (r *schemaResolver) LSIFDumps(args *struct {
	graphqlutil.ConnectionArgs
	Repository string
}) (*lsifDumpConnectionResolver, error) {
	opt := LSIFDumpsListOptions{
		Repository: args.Repository,
	}

	return &lsifDumpConnectionResolver{opt: opt}, nil
}

type lsifDumpConnectionResolver struct {
	opt LSIFDumpsListOptions

	// cache results because they are used by multiple fields
	once       sync.Once
	dumps      []*types.LSIFDump
	totalCount int
	err        error
}

func (r *lsifDumpConnectionResolver) compute(ctx context.Context) ([]*types.LSIFDump, int, error) {
	r.once.Do(func() {
		query := url.Values{}
		query.Add("repository", r.opt.Repository)

		payload := []*types.LSIFDump{}
		err := lsifRequest(ctx, "dumps", query, &payload)
		if err != nil {
			r.err = err
			return
		}

		r.dumps = payload
		r.totalCount = len(r.dumps)
	})

	return r.dumps, r.totalCount, r.err
}

func (r *lsifDumpConnectionResolver) Nodes(ctx context.Context) ([]*lsifDumpResolver, error) {
	dumps, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []*lsifDumpResolver
	for _, lsifDump := range dumps {
		l = append(l, &lsifDumpResolver{
			lsifDump: lsifDump,
		})
	}
	return l, nil
}

func (r *lsifDumpConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, count, err := r.compute(ctx)
	return int32(count), err
}

func (r *lsifDumpConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	return graphqlutil.HasNextPage(false), nil
}

func marshalLSIFDumpsCursorGQLID(offset int) graphql.ID {
	return relay.MarshalID("LSIFDumpsCursor", offset)
}

func unmarshalLSIFDumpsCursorGQLID(id graphql.ID) (offset int, err error) {
	err = relay.UnmarshalSpec(id, &offset)
	return
}
