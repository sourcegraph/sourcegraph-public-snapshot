package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/discovery"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/schema"
)

var _ graphqlbackend.InsightConnectionResolver = &insightConnectionResolver{}

type insightConnectionResolver struct {
	insightsStore   store.Interface
	workerBaseStore *basestore.Store
	settingStore    discovery.SettingStore

	// cache results because they are used by multiple fields
	once     sync.Once
	insights []*schema.Insight
	next     int64
	err      error

	denylist []api.RepoID
}

func (r *insightConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.InsightResolver, error) {
	nodes, denylist, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.InsightResolver, 0, len(nodes))
	for _, insight := range nodes {
		resolvers = append(resolvers, &insightResolver{
			insightsStore:   r.insightsStore,
			workerBaseStore: r.workerBaseStore,
			insight:         insight,
			denylist:        denylist,
		})
	}
	return resolvers, nil
}

func (r *insightConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	insights, _, _, err := r.compute(ctx)
	return int32(len(insights)), err
}

func (r *insightConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *insightConnectionResolver) compute(ctx context.Context) ([]*schema.Insight, []api.RepoID, int64, error) {
	r.once.Do(func() {
		var multi error
		insights, err := discovery.Discover(ctx, r.settingStore)
		r.insights = insights
		if r.err != nil {
			multi = multierror.Append(multi, r.err)
		}

		// 🚨 SECURITY: This is a double-negative repo permission enforcement. The list of authorized repos is generally expected to be very large, and nearly the full
		// set of repos installed on Sourcegraph. To make this faster, we query Postgres for a list of repos the current user cannot see, and then exclude those from the
		// time series results. 🚨
		denylist, err := FetchUnauthorizedRepos(ctx, r.workerBaseStore.Handle().DB())
		r.denylist = denylist
		if err != nil {
			multi = multierror.Append(multi, err)
		}
		r.err = multi
	})
	return r.insights, r.denylist, r.next, r.err
}

// InsightResolver is also defined here as it is covered by the same tests.

var _ graphqlbackend.InsightResolver = &insightResolver{}

type insightResolver struct {
	insightsStore   store.Interface
	workerBaseStore *basestore.Store
	insight         *schema.Insight
	denylist        []api.RepoID
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
			denylist:        r.denylist,
		})
	}
	return resolvers
}
