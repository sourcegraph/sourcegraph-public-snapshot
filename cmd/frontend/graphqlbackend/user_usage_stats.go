package graphqlbackend

import (
	"context"
	"errors"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestats2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func (r *UserResolver) UsageStatistics(ctx context.Context) (*userUsageStatisticsResolver, error) {
	if envvar.SourcegraphDotComMode() {
		return nil, errors.New("usage statistics are not available on sourcegraph.com")
	}

	stats, err := usagestats2.GetByUserID(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	return &userUsageStatisticsResolver{stats}, nil
}

type userUsageStatisticsResolver struct {
	userUsageStatistics *types.UserUsageStatistics
}

func (s *userUsageStatisticsResolver) PageViews() int32 { return s.userUsageStatistics.PageViews }

func (s *userUsageStatisticsResolver) SearchQueries() int32 {
	return s.userUsageStatistics.SearchQueries
}

func (s *userUsageStatisticsResolver) CodeIntelligenceActions() int32 {
	return s.userUsageStatistics.CodeIntelligenceActions
}

func (s *userUsageStatisticsResolver) FindReferencesActions() int32 {
	return s.userUsageStatistics.FindReferencesActions
}

func (s *userUsageStatisticsResolver) LastActiveTime() *string {
	if s.userUsageStatistics.LastActiveTime != nil {
		t := s.userUsageStatistics.LastActiveTime.Format(time.RFC3339)
		return &t
	}
	return nil
}

func (s *userUsageStatisticsResolver) LastActiveCodeHostIntegrationTime() *string {
	if s.userUsageStatistics.LastCodeHostIntegrationTime != nil {
		t := s.userUsageStatistics.LastCodeHostIntegrationTime.Format(time.RFC3339)
		return &t
	}
	return nil
}

func (*schemaResolver) LogUserEvent(ctx context.Context, args *struct {
	Event        string
	UserCookieID string
}) (*EmptyResponse, error) {
	return nil, nil
}

func (*schemaResolver) LogEvent(ctx context.Context, args *struct {
	Event        string
	UserCookieID string
	URL          string
	Source       string
	Argument     *string
}) (*EmptyResponse, error) {
	if !conf.EventLoggingEnabled() {
		return nil, nil
	}
	actor := actor.FromContext(ctx)
	return nil, usagestats2.LogEvent(ctx, usagestats2.Event{
		EventName:    args.Event,
		URL:          args.URL,
		UserID:       actor.UID,
		UserCookieID: args.UserCookieID,
		Source:       args.Source,
		Argument:     args.Argument,
	})
}
