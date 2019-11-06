package graphqlbackend

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/lsif"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type LSIFDumpsListOptions struct {
	Repository      graphql.ID
	Query           *string
	IsLatestForRepo *bool
	Limit           *int32
	NextURL         *string
}

// This method implements cursor-based forward pagination. The `after` parameter
// should be an `endCursor` value from a previous request. This value is the rel="next"
// URL in the Link header of the LSIF server response. This URL includes all of the
// query variables required to fetch the subsequent page of results. This state is not
// dependent on the limit, so we can overwrite this value if the user has changed its
// value since making the last request.

func (r *schemaResolver) LSIFDumps(args *struct {
	graphqlutil.ConnectionArgs
	Repository      graphql.ID
	Query           *string
	IsLatestForRepo *bool
	After           *string
}) (*lsifDumpConnectionResolver, error) {
	opt := LSIFDumpsListOptions{
		Repository:      args.Repository,
		Query:           args.Query,
		IsLatestForRepo: args.IsLatestForRepo,
	}
	if args.First != nil {
		if *args.First < 0 || *args.First > 5000 {
			return nil, errors.New("lsifDumps: requested pagination 'first' value outside allowed range (0 - 5000)")
		}

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

	return &lsifDumpConnectionResolver{opt: opt}, nil
}

type lsifDumpConnectionResolver struct {
	opt LSIFDumpsListOptions

	// cache results because they are used by multiple fields
	once       sync.Once
	dumps      []*types.LSIFDump
	repo       *RepositoryResolver
	totalCount int
	nextURL    string
	err        error
}

func (r *lsifDumpConnectionResolver) compute(ctx context.Context) ([]*types.LSIFDump, *RepositoryResolver, int, string, error) {
	r.once.Do(func() {
		repo, err := repositoryByID(ctx, r.opt.Repository)
		if err != nil {
			r.err = err
			return
		}

		var path string
		if r.opt.NextURL == nil {
			// first page of results
			path = fmt.Sprintf("/dumps/%s", url.PathEscape(repo.Name()))
		} else {
			// subsequent page of results
			path = *r.opt.NextURL
		}

		query := url.Values{}
		if r.opt.Query != nil {
			query.Set("query", *r.opt.Query)
		}
		if r.opt.IsLatestForRepo != nil && *r.opt.IsLatestForRepo {
			query.Set("visibleAtTip", "true")
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
		r.repo = repo
		r.totalCount = payload.TotalCount
		r.nextURL = lsif.ExtractNextURL(resp)
	})

	return r.dumps, r.repo, r.totalCount, r.nextURL, r.err
}

func (r *lsifDumpConnectionResolver) Nodes(ctx context.Context) ([]*lsifDumpResolver, error) {
	dumps, repo, _, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []*lsifDumpResolver
	for _, lsifDump := range dumps {
		l = append(l, &lsifDumpResolver{
			repo:     repo.repo,
			lsifDump: lsifDump,
		})
	}
	return l, nil
}

func (r *lsifDumpConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, _, count, _, err := r.compute(ctx)
	return int32(count), err
}

func (r *lsifDumpConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, _, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}
