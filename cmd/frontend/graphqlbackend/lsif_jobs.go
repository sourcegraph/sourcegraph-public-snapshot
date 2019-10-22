package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type LsifJobsListOptions struct {
	Status string
	Query  *string
	Limit  *int32
	Offset *int
}

func (r *schemaResolver) LsifJobs(args *struct {
	graphqlutil.ConnectionArgs
	Status string
	Query  *string
	After  *graphql.ID
}) (*lsifJobConnectionResolver, error) {
	opt := LsifJobsListOptions{
		Status: args.Status,
		Query:  args.Query,
	}
	if args.First != nil {
		opt.Limit = args.First
	}
	if args.After != nil {
		offset, err := unmarshalLsifJobsCursorGQLID(*args.After)
		if err != nil {
			return nil, err
		}
		opt.Offset = &offset
	}

	return &lsifJobConnectionResolver{opt: opt}, nil
}

type lsifJobConnectionResolver struct {
	opt LsifJobsListOptions

	// cache results because they are used by multiple fields
	once       sync.Once
	jobs       []*types.LsifJob
	totalCount int
	err        error
}

func (r *lsifJobConnectionResolver) compute(ctx context.Context) ([]*types.LsifJob, int, error) {
	r.once.Do(func() {
		query := url.Values{}
		if r.opt.Query != nil {
			query.Set("search", *r.opt.Query)
		}
		if r.opt.Limit != nil {
			query.Set("limit", fmt.Sprintf("%d", *r.opt.Limit))
		}
		if r.opt.Offset != nil {
			query.Set("offset", fmt.Sprintf("%d", *r.opt.Offset))
		}

		var payload struct {
			Jobs       []*types.LsifJob `json:"jobs"`
			TotalCount int              `json:"totalCount"`
		}

		err := lsifRequest(ctx, fmt.Sprintf("jobs/%s", r.opt.Status), query, &payload)
		if err != nil {
			r.err = err
			return
		}

		r.jobs = payload.Jobs
		r.totalCount = payload.TotalCount
	})

	return r.jobs, r.totalCount, r.err
}

func (r *lsifJobConnectionResolver) Nodes(ctx context.Context) ([]*lsifJobResolver, error) {
	jobs, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []*lsifJobResolver
	for _, lsifJob := range jobs {
		l = append(l, &lsifJobResolver{
			lsifJob: lsifJob,
		})
	}
	return l, nil
}

func (r *lsifJobConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, count, err := r.compute(ctx)
	return int32(count), err
}

func (r *lsifJobConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	jobs, count, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	seen := len(jobs)
	if r.opt.Offset != nil {
		seen += *r.opt.Offset
	}

	if seen < count {
		return graphqlutil.NextPageCursor(marshalLsifJobsCursorGQLID(seen)), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func marshalLsifJobsCursorGQLID(offset int) graphql.ID {
	return relay.MarshalID("LsifJobsCursor", offset)
}

func unmarshalLsifJobsCursorGQLID(id graphql.ID) (offset int, err error) {
	err = relay.UnmarshalSpec(id, &offset)
	return
}
