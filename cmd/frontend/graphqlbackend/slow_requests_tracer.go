package graphqlbackend

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// slowRequestRedisFIFOListPerPage sets the default count of returned request.
const slowRequestRedisFIFOListPerPage = 50

// slowRequestRedisFIFOList is a FIFO redis cache to store the slow requests.
var slowRequestRedisFIFOList = rcache.NewFIFOListDynamic(redispool.Cache, "slow-graphql-requests-list", func() int {
	return conf.Get().ObservabilityCaptureSlowGraphQLRequestsLimit
})

// captureSlowRequest stores in a redis cache slow GraphQL requests.
func captureSlowRequest(logger log.Logger, req *types.SlowRequest) {
	b, err := json.Marshal(req)
	if err != nil {
		logger.Warn("failed to marshal slowRequest", log.Error(err))
		return
	}
	if err := slowRequestRedisFIFOList.Insert(b); err != nil {
		logger.Warn("failed to capture slowRequest", log.Error(err))
	}
}

// getSlowRequestsAfter returns the last limit slow requests, starting at the request whose ID is set to after.
func getSlowRequestsAfter(ctx context.Context, list *rcache.FIFOList, after int, limit int) ([]*types.SlowRequest, error) {
	raws, err := list.Slice(ctx, after, after+limit-1)
	if err != nil {
		return nil, err
	}

	reqs := make([]*types.SlowRequest, len(raws))
	for i, raw := range raws {
		var req types.SlowRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, err
		}
		req.Index = strconv.Itoa(i + after)
		reqs[i] = &req
	}
	return reqs, nil
}

// SlowRequests returns a connection to fetch slow requests.
func (r *schemaResolver) SlowRequests(ctx context.Context, args *slowRequestsArgs) (*slowRequestConnectionResolver, error) {
	if conf.Get().ObservabilityCaptureSlowGraphQLRequestsLimit == 0 {
		return nil, errors.New("slow graphql requests capture is not enabled")
	}
	// 🚨 SECURITY: Only site admins may list outbound requests.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	after := "0"
	if args.After != nil {
		after = *args.After
	}
	return &slowRequestConnectionResolver{
		after:           after,
		perPage:         slowRequestRedisFIFOListPerPage,
		gitserverClient: r.gitserverClient,
		db:              r.db,
	}, nil
}

type slowRequestConnectionResolver struct {
	reqs []*types.SlowRequest

	after      string
	perPage    int
	totalCount int32

	err             error
	once            sync.Once
	gitserverClient gitserver.Client
	db              database.DB
}

type slowRequestsArgs struct {
	After *string
}

type slowRequestResolver struct {
	req *types.SlowRequest

	db              database.DB
	gitserverClient gitserver.Client
}

func (r *slowRequestConnectionResolver) fetch(ctx context.Context) ([]*types.SlowRequest, error) {
	r.once.Do(func() {
		n, err := strconv.Atoi(r.after)
		if err != nil {
			r.err = err
		}
		r.reqs, r.err = getSlowRequestsAfter(ctx, slowRequestRedisFIFOList, n, r.perPage)
		size, err := slowRequestRedisFIFOList.Size()
		if err != nil {
			r.err = errors.Append(r.err, err)
		} else {
			r.totalCount = int32(size)
		}
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
		resolvers = append(resolvers, &slowRequestResolver{
			req:             req,
			db:              r.db,
			gitserverClient: r.gitserverClient,
		})
	}
	return resolvers, nil
}

func (r *slowRequestConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, err := r.fetch(ctx)
	return r.totalCount, err
}

func (r *slowRequestConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
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
		return gqlutil.HasNextPage(false), nil
	} else {
		return gqlutil.NextPageCursor(reqs[len(reqs)-1].Index), nil
	}
}

// Index returns an opaque ID for that node.
func (r *slowRequestResolver) Index() string {
	return r.req.Index
}

// Start returns the start time of the slow request.
func (r *slowRequestResolver) Start() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.req.Start}
}

// Duration returns the recorded duration of the slow request.
func (r *slowRequestResolver) Duration() float64 {
	return r.req.Duration.Seconds()
}

func (r *slowRequestResolver) User(ctx context.Context) (*UserResolver, error) {
	if r.req.UserID == 0 {
		return nil, nil
	}
	return UserByID(ctx, r.db, MarshalUserID(r.req.UserID))
}

// Name returns the GraqhQL request name, if any. Blank if none.
func (r *slowRequestResolver) Name() string {
	return r.req.Name
}

// repoName guesses the name of the associated repository. Returns nil if none is found.
func guessRepoName(variables map[string]any) *string {
	if repoName, ok := variables["repoName"]; ok {
		if str, ok := repoName.(string); ok {
			return &str
		}
	}
	if repoName, ok := variables["repository"]; ok {
		if str, ok := repoName.(string); ok {
			return &str
		}
	}
	return nil
}

func (r *slowRequestResolver) getRepo(ctx context.Context) (*types.Repo, error) {
	if name := guessRepoName(r.req.Variables); name != nil {
		return r.db.Repos().GetByName(ctx, api.RepoName(*name))
	}
	return nil, nil
}

func (r *slowRequestResolver) Repository(ctx context.Context) (*RepositoryResolver, error) {
	repo, err := r.getRepo(ctx)
	if err != nil {
		return nil, err
	}
	if repo != nil {
		return NewRepositoryResolver(r.db, r.gitserverClient, repo), nil
	}
	return nil, nil
}

// Filepath guesses the name of the associated filepath if possible.
// Blank if none.
func (r *slowRequestResolver) Filepath() *string {
	if filepath, ok := r.req.Variables["filePath"]; ok {
		if str, ok := filepath.(string); ok {
			return &str
		}
	}
	if path, ok := r.req.Variables["path"]; ok {
		if str, ok := path.(string); ok {
			return &str
		}
	}
	return nil
}

// Query returns the GraphQL query performed by the slow request.
func (r *slowRequestResolver) Query() string {
	return r.req.Query
}

// Variables returns the GraphQL variables associated with the query
// performed by the request.
func (r *slowRequestResolver) Variables() string {
	raw, _ := json.Marshal(r.req.Variables)
	return string(raw)
}

// Errors returns a list of errors encountered when handling
// the slow request.
func (r *slowRequestResolver) Errors() []string {
	return r.req.Errors
}

// Source returns from where the GraphQL originated.
func (r *slowRequestResolver) Source() string {
	return r.req.Source
}
