package resolvers

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

var _ graphqlbackend.InsightConnectionResolver = &insightConnectionResolver{}

type insightConnectionResolver struct {
	store        *store.Store
	settingStore *database.SettingStore

	// cache results because they are used by multiple fields
	once     sync.Once
	insights []graphqlbackend.InsightResolver
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
		resolvers = append(resolvers, &insightResolver{store: r.store, insight: insight})
	}
	return resolvers, nil
}

func (r *insightConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return 0, errors.New("not yet implemented")
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

func (r *insightConnectionResolver) compute(ctx context.Context) ([]graphqlbackend.InsightResolver, int64, error) {
	r.once.Do(func() {
		// TODO: populate r.insights, r.next, r.err
		// TODO: locate insights from user, org, global settings using r.settingStore.ListAll()
	})
	return r.insights, r.next, r.err
}
