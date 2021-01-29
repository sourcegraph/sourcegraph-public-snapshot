package resolvers

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

var _ graphqlbackend.InsightConnectionResolver = &insightConnectionResolver{}

type insightConnectionResolver struct {
	store        *store.Store
	settingStore *database.SettingStore

	// cache results because they are used by multiple fields
	once     sync.Once
	insights []*schema.Insight
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

func (r *insightConnectionResolver) compute(ctx context.Context) ([]*schema.Insight, int64, error) {
	r.once.Do(func() {
		// Get latest Global user settings.
		//
		// FUTURE: include user/org settings.
		subject := api.SettingsSubject{Site: true}
		globalSettingsRaw, err := r.settingStore.GetLatest(ctx, subject)
		if err != nil {
			r.err = err
			return
		}
		globalSettings, err := parseUserSettings(globalSettingsRaw)
		r.insights = globalSettings.Insights
	})
	return r.insights, r.next, r.err
}

func parseUserSettings(settings *api.Settings) (*schema.Settings, error) {
	if settings == nil {
		// Settings have never been saved for this subject; equivalent to `{}`.
		return &schema.Settings{}, nil
	}
	var v schema.Settings
	if err := jsonc.Unmarshal(settings.Contents, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
