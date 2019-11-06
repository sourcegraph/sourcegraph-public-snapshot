package graphqlbackend

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/lsif"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type LSIFJobsListOptions struct {
	State   string
	Query   *string
	Limit   *int32
	NextURL *string
}

// This method implements cursor-based forward pagination. The `after` parameter
// should be an `endCursor` value from a previous request. This value is the rel="next"
// URL in the Link header of the LSIF server response. This URL includes all of the
// query variables required to fetch the subsequent page of results. This state is not
// dependent on the limit, so we can overwrite this value if the user has changed its
// value since making the last request.

func (r *schemaResolver) LSIFJobs(args *struct {
	graphqlutil.ConnectionArgs
	State string
	Query *string
	After *string
}) (*lsifJobConnectionResolver, error) {
	opt := LSIFJobsListOptions{
		State: args.State,
		Query: args.Query,
	}
	if args.First != nil {
		opt.Limit = args.First
	}
	if args.After != nil {
		decoded, err := base64.StdEncoding.DecodeString(*args.After)
		if err != nil {
			return nil, err
		}
		nextURL := string(decoded)
		opt.NextURL = &nextURL
	}

	return &lsifJobConnectionResolver{opt: opt}, nil
}

type lsifJobConnectionResolver struct {
	opt LSIFJobsListOptions

	// cache results because they are used by multiple fields
	once       sync.Once
	jobs       []*types.LSIFJob
	totalCount *int
	nextURL    string
	err        error
}

func (r *lsifJobConnectionResolver) compute(ctx context.Context) ([]*types.LSIFJob, *int, string, error) {
	r.once.Do(func() {
		var path string
		if r.opt.NextURL == nil {
			// first page of results
			path = fmt.Sprintf("/jobs/%s", strings.ToLower(r.opt.State))
		} else {
			// subsequent page of results
			path = *r.opt.NextURL
		}

		query := url.Values{}
		if r.opt.Query != nil {
			query.Set("query", *r.opt.Query)
		}
		if r.opt.Limit != nil {
			query.Set("limit", strconv.FormatInt(int64(*r.opt.Limit), 10))
		}

		resp, err := lsif.BuildAndTraceRequest(ctx, path, query)
		if err != nil {
			r.err = err
			return
		}

		payload := struct {
			Jobs       []*types.LSIFJob `json:"jobs"`
			TotalCount *int             `json:"totalCount"`
		}{
			Jobs: []*types.LSIFJob{},
		}

		if err := lsif.UnmarshalPayload(resp, &payload); err != nil {
			r.err = err
			return
		}

		r.jobs = payload.Jobs
		r.totalCount = payload.TotalCount
		r.nextURL = lsif.ExtractNextURL(resp)
	})

	return r.jobs, r.totalCount, r.nextURL, r.err
}

func (r *lsifJobConnectionResolver) Nodes(ctx context.Context) ([]*lsifJobResolver, error) {
	jobs, _, _, err := r.compute(ctx)
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

func (r *lsifJobConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	_, count, _, err := r.compute(ctx)
	if count == nil || err != nil {
		return nil, err
	}

	c := int32(*count)
	return &c, nil
}

func (r *lsifJobConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}
