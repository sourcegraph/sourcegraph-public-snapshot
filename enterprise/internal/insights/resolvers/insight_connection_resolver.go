package resolvers

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.InsightConnectionResolver = &insightConnectionResolver{}

type insightConnectionResolver struct {
	insightsStore        store.Interface
	workerBaseStore      *basestore.Store
	orgStore             database.OrgStore
	insightMetadataStore store.InsightMetadataStore

	// arguments from query
	ids []string

	// cache results because they are used by multiple fields
	once     sync.Once
	insights []types.Insight
	next     int64
	err      error
}

func (r *insightConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.InsightResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.InsightResolver, 0, len(nodes))
	for _, insight := range nodes {
		resolvers = append(resolvers, &insightResolver{
			insightsStore:   r.insightsStore,
			workerBaseStore: r.workerBaseStore,
			insight:         insight,
			metadataStore:   r.insightMetadataStore,
		})
	}
	return resolvers, nil
}

func (r *insightConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	results, _, err := r.compute(ctx)
	return int32(len(results)), err
}

func (r *insightConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *insightConnectionResolver) compute(ctx context.Context) ([]types.Insight, int64, error) {
	r.once.Do(func() {
		args := store.InsightQueryArgs{UniqueIDs: r.ids}
		var err error
		args.UserID, args.OrgID, err = getUserPermissions(ctx, r.orgStore)
		if err != nil {
			r.err = errors.Wrap(err, "getUserPermissions")
			return
		}

		mapped, err := r.insightMetadataStore.GetMapped(ctx, args)
		if err != nil {
			r.err = err
			return
		}
		// currently insight metadata is partially stored in user settings. Series will be joined with their appropriate
		// metadata by sorting based on query text, and joining in the frontend. This is largely a temporary solution
		// until insights has a full graphql api
		for _, insight := range mapped {
			sort.Slice(insight.Series, func(i, j int) bool {
				return strings.ToUpper(insight.Series[i].Query) < strings.ToUpper(insight.Series[j].Query)
			})
		}

		r.insights = mapped
	})
	return r.insights, r.next, r.err
}

// InsightResolver is also defined here as it is covered by the same tests.

var _ graphqlbackend.InsightResolver = &insightResolver{}

type insightResolver struct {
	insightsStore   store.Interface
	workerBaseStore *basestore.Store
	metadataStore   store.InsightMetadataStore
	insight         types.Insight
}

func (r *insightResolver) ID() string {
	return r.insight.UniqueID
}

func (r *insightResolver) Title() string { return r.insight.Title }

func (r *insightResolver) Description() string { return r.insight.Description }

func (r *insightResolver) Series() []graphqlbackend.InsightSeriesResolver {
	series := r.insight.Series
	resolvers := make([]graphqlbackend.InsightSeriesResolver, 0, len(series))
	for _, series := range series {
		resolvers = append(resolvers, &insightSeriesResolver{
			insightsStore:   r.insightsStore,
			workerBaseStore: r.workerBaseStore,
			series:          series,
			metadataStore:   r.metadataStore,
		})
	}
	return resolvers
}
