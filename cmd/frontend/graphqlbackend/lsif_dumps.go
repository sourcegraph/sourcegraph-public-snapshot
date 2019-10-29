package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/lsif"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type LSIFDumpsListOptions struct {
	Repository string
	Query      *string
	Limit      *int32
	NextURL    *string
}

// This method implements cursor-based forward pagination. The `after` parameter
// should be an `endCursor` value from a previous request. This value is the rel="next"
// URL in the Link header of the LSIF server response. This URL includes all of the
// query variables required to fetch the subsequent page of results. This state is not
// dependent on the limit, so we can overwrite this value if the user has changed its
// value since making the last request.

func (r *schemaResolver) LSIFDumps(args *struct {
	graphqlutil.ConnectionArgs
	Repository string
	Query      *string
	After      *graphql.ID
}) (*lsifDumpConnectionResolver, error) {
	opt := LSIFDumpsListOptions{
		Repository: args.Repository,
		Query:      args.Query,
	}
	if args.First != nil {
		opt.Limit = args.First
	}
	if args.After != nil {
		nextURL, err := unmarshalLSIFDumpsCursorGQLID(*args.After)
		if err != nil {
			return nil, err
		}
		opt.NextURL = &nextURL
	}

	return &lsifDumpConnectionResolver{opt: opt}, nil
}

type lsifDumpConnectionResolver struct {
	opt LSIFDumpsListOptions

	// cache results because they are used by multiple fields
	once       sync.Once
	dumps      []*types.LSIFDump
	totalCount int
	nextURL    string
	err        error
}

func (r *lsifDumpConnectionResolver) compute(ctx context.Context) ([]*types.LSIFDump, int, string, error) {
	r.once.Do(func() {
		var path string
		if r.opt.NextURL == nil {
			// first page of results
			path = fmt.Sprintf("/dumps/%s", url.QueryEscape(r.opt.Repository))
		} else {
			// subsequent page of results
			path = *r.opt.NextURL
		}

		query := url.Values{}
		if r.opt.Query != nil {
			query.Set("query", *r.opt.Query)
		}
		if r.opt.Limit != nil {
			query.Set("limit", fmt.Sprintf("%d", *r.opt.Limit))
		}

		resp, err := lsif.BuildAndTraceRequest(ctx, path, query)
		if err != nil {
			r.err = err
			return
		}

		payload := struct {
			Dumps      []*types.LSIFDump `json:"dumps"`
			TotalCount int               `json:"totalCount"`
		}{
			Dumps: []*types.LSIFDump{},
		}

		if err := lsif.UnmarshalPayload(resp, &payload); err != nil {
			r.err = err
			return
		}

		r.dumps = payload.Dumps
		r.totalCount = payload.TotalCount
		r.nextURL = lsif.ExtractNextURL(resp)
	})

	return r.dumps, r.totalCount, r.nextURL, r.err
}

func (r *lsifDumpConnectionResolver) Nodes(ctx context.Context) ([]*lsifDumpResolver, error) {
	dumps, _, _, err := r.compute(ctx)
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
	_, count, _, err := r.compute(ctx)
	return int32(count), err
}

func (r *lsifDumpConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(marshalLSIFDumpsCursorGQLID(nextURL)), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func marshalLSIFDumpsCursorGQLID(nextURL string) graphql.ID {
	return relay.MarshalID("LSIFDumpsCursor", nextURL)
}

func unmarshalLSIFDumpsCursorGQLID(id graphql.ID) (nextURL string, err error) {
	err = relay.UnmarshalSpec(id, &nextURL)
	return
}
