package graphqlbackend

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var slowRequestRedisTTL = int(24 * time.Hour / time.Second)
var slowRequestRedisStore = rcache.NewWithTTL("slow-graphql-requests", slowRequestRedisTTL)
var slowRequestRedisRecentList = rcache.NewRecentList("slow-graphql-requests-list", 5000)
var myOnce sync.Once

// captureSlowRequest stores in a redis cache slow GraphQL requests.
func captureSlowRequest(ctx context.Context, logger log.Logger, req *types.SlowRequest) {
	myOnce.Do(func() {
		conf.Watch(func() {
			limit := conf.Get().ObservabilityCaptureSlowGraphQLRequestsLimit
			if limit == 0 {
				limit = 5000
			}
			slowRequestRedisRecentList = rcache.NewRecentList("slow-graphql-requests-list", limit)
		})
	})

	b, err := json.Marshal(req)
	if err != nil {
		logger.Warn("failed to marshal slowRequest", log.Error(err))
		return
	}
	slowRequestRedisRecentList.Insert(b)
}

// getSlowRequestsAfter returns the last limit slow requests, starting at the request whose ID is set to after.
func getSlowRequestsAfter(ctx context.Context, list *rcache.RecentList, after int, limit int) ([]*types.SlowRequest, error) {
	raws, err := list.Slice(ctx, after, after+limit-1)
	if err != nil {
		return nil, err
	}

	reqs := make([]*types.SlowRequest, 0, len(raws))
	for i, raw := range raws {
		var req types.SlowRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, err
		}
		req.ID = strconv.Itoa(i + after)
		reqs = append(reqs, &req)
	}
	return reqs, nil
}

type slowRequestsArgs struct {
	After *graphql.ID
	Query *string
}

type slowRequestResolver struct {
	req *types.SlowRequest
}

type slowRequestConnectionResolver struct {
	after string

	err     error
	once    sync.Once
	reqs    []*types.SlowRequest
	perPage int
}

func (r *schemaResolver) SlowRequests(ctx context.Context, args *slowRequestsArgs) (*slowRequestConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may list outbound requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	var after string
	if args.After != nil {
		err := relay.UnmarshalSpec(*args.After, &after)
		if err != nil {
			return nil, err
		}
	} else {
		after = "0"
	}
	return &slowRequestConnectionResolver{
		after:   after,
		perPage: 50,
	}, nil
}

func (r *slowRequestConnectionResolver) fetch(ctx context.Context) ([]*types.SlowRequest, error) {
	r.once.Do(func() {
		n, err := strconv.Atoi(r.after)
		if err != nil {
			r.err = err
		}
		r.reqs, r.err = getSlowRequestsAfter(ctx, slowRequestRedisRecentList, n, r.perPage)
		println("wfpwffeeeeee--------", len(r.reqs))
	})
	return r.reqs, r.err
}

func (r *slowRequestConnectionResolver) Nodes(ctx context.Context) ([]*slowRequestResolver, error) {
	reqs, err := r.fetch(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*slowRequestResolver, 0, len(reqs))
	for _, req := range r.reqs {
		resolvers = append(resolvers, &slowRequestResolver{req: req})
	}
	return resolvers, nil
}

func (r *slowRequestConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	n := slowRequestRedisRecentList.Size()
	return int32(n), nil
}

func (r *slowRequestConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	reqs, err := r.fetch(ctx)
	if err != nil {
		return nil, err
	}

	n, err := strconv.Atoi(r.after)
	if err != nil {
		return nil, err
	}
	total, err := r.TotalCount(ctx)
	if err != nil {
		return nil, err
	}
	if int32(n+r.perPage) >= total {
		return graphqlutil.HasNextPage(false), nil
	} else {
		return graphqlutil.NextPageCursor(string(relay.MarshalID("SlowRequest", reqs[len(reqs)-1].ID))), nil
	}
}

func (r *slowRequestResolver) ID() graphql.ID {
	return relay.MarshalID("SlowRequest", r.req.ID)
}

func (r *slowRequestResolver) Start() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.req.Start}
}

func (r *slowRequestResolver) Duration() float64 {
	return r.req.Duration.Seconds()
}

func (r *slowRequestResolver) UserId() string {
	return strconv.Itoa(int(r.req.UserID))
}

func (r *slowRequestResolver) Name() string {
	return r.req.Name
}

func (r *slowRequestResolver) RepoName() string {
	if repoName, ok := r.req.Variables["repoName"]; ok {
		if str, ok := repoName.(string); ok {
			return str
		}
	}
	if repoName, ok := r.req.Variables["repository"]; ok {
		if str, ok := repoName.(string); ok {
			return str
		}
	}
	return ""
}

func (r *slowRequestResolver) Filepath() string {
	if filepath, ok := r.req.Variables["filePath"]; ok {
		if str, ok := filepath.(string); ok {
			return str
		}
	}
	if path, ok := r.req.Variables["path"]; ok {
		if str, ok := path.(string); ok {
			return str
		}
	}
	return ""
}

func (r *slowRequestResolver) Query() string {
	return r.req.Query
}

func (r *slowRequestResolver) Variables() string {
	raw, _ := json.Marshal(r.req.Variables)
	return string(raw)
}

func (r *slowRequestResolver) Errors() []string {
	return r.req.Errors
}

func (r *slowRequestResolver) Source() string {
	return r.req.Source
}
