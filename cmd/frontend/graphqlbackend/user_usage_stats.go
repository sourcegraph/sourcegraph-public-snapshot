package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestats2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/pubsub/pubsubutil"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// pubSubDotComEventsTopicID is the topic ID of the topic that forwards messages to Sourcegraph.com events' pub/sub subscribers.
var pubSubDotComEventsTopicID = env.Get("PUBSUB_DOTCOM_EVENTS_TOPIC_ID", "", "Pub/sub dotcom events topic ID is the pub/sub topic id where Sourcegraph.com events are published.")

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

	// On Sourcegraph.com, log events to BigQuery instead of the internal Postgres table.
	if envvar.SourcegraphDotComMode() {
		if pubSubDotComEventsTopicID == "" {
			return nil, nil
		}
		var argument string
		if args.Argument != nil {
			argument = *args.Argument
		}
		event, err := json.Marshal(bigQueryEvent{
			EventName:       args.Event,
			UserID:          int(actor.UID),
			AnonymousUserID: args.UserCookieID,
			URL:             args.URL,
			Source:          args.Source,
			Argument:        argument,
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
			Version:         version.Version(),
		})
		if err != nil {
			return nil, err
		}
		return nil, pubsubutil.Publish(pubSubDotComEventsTopicID, string(event))
	}

	return nil, usagestats2.LogEvent(
		ctx,
		args.Event,
		args.URL,
		actor.UID,
		args.UserCookieID,
		args.Source,
		args.Argument,
	)
}

type bigQueryEvent struct {
	EventName       string `json:"name"`
	AnonymousUserID string `json:"anonymous_user_id"`
	UserID          int    `json:"user_id"`
	URL             string `json:"url"`
	Source          string `json:"source"`
	Argument        string `json:"argument,omitempty"`
	Timestamp       string `json:"timestamp"`
	Version         string `json:"version"`
}
