package graphqlbackend

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

func (r *UserResolver) UsageStatistics(ctx context.Context) (*userUsageStatisticsResolver, error) {
	if envvar.SourcegraphDotComMode() {
		if err := backend.CheckSiteAdminOrSameUser(ctx, r.db, r.user.ID); err != nil {
			return nil, err
		}
	}

	stats, err := usagestats.GetByUserID(ctx, r.db, r.user.ID)
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

// No longer used, only here for backwards compatibility with IDE and browser extensions.
// Functionality removed in https://github.com/sourcegraph/sourcegraph/pull/38826.
func (*schemaResolver) LogUserEvent(ctx context.Context, args *struct {
	Event        string
	UserCookieID string
}) (*EmptyResponse, error) {
	return nil, nil
}

type Event struct {
	Event          string
	UserCookieID   string
	FirstSourceURL *string
	LastSourceURL  *string
	URL            string
	Source         string
	Argument       *string
	CohortID       *string
	Referrer       *string
	PublicArgument *string
	UserProperties *string
	DeviceID       *string
	InsertID       *string
	EventID        *int32
}

type EventBatch struct {
	Events *[]Event
}

func (r *schemaResolver) LogEvent(ctx context.Context, args *Event) (*EmptyResponse, error) {
	if args == nil {
		return nil, nil
	}

	return r.LogEvents(ctx, &EventBatch{Events: &[]Event{*args}})
}

func (r *schemaResolver) LogEvents(ctx context.Context, args *EventBatch) (*EmptyResponse, error) {
	if !conf.EventLoggingEnabled() || args.Events == nil {
		return nil, nil
	}

	decode := func(v *string) (payload json.RawMessage, _ error) {
		if v != nil {
			if err := json.Unmarshal([]byte(*v), &payload); err != nil {
				return nil, err
			}
		}

		return payload, nil
	}

	events := make([]usagestats.Event, 0, len(*args.Events))
	for _, args := range *args.Events {
		if strings.HasPrefix(args.Event, "search.latencies.frontend.") {
			argumentPayload, err := decode(args.Argument)
			if err != nil {
				return nil, err
			}

			if err := exportPrometheusSearchLatencies(args.Event, argumentPayload); err != nil {
				log15.Error("export prometheus search latencies", "error", err)
			}

			// Future(slimsag): implement actual event logging for these events
			continue
		}

		if strings.HasPrefix(args.Event, "search.ranking.") {
			argumentPayload, err := decode(args.Argument)
			if err != nil {
				return nil, err
			}
			if err := exportPrometheusSearchRanking(argumentPayload); err != nil {
				log15.Error("exportPrometheusSearchRanking", "error", err)
			}
			continue
		}

		argumentPayload, err := decode(args.Argument)
		if err != nil {
			return nil, err
		}

		publicArgumentPayload, err := decode(args.PublicArgument)
		if err != nil {
			return nil, err
		}

		userPropertiesPayload, err := decode(args.UserProperties)
		if err != nil {
			return nil, err
		}

		events = append(events, usagestats.Event{
			EventName:        args.Event,
			URL:              args.URL,
			UserID:           actor.FromContext(ctx).UID,
			UserCookieID:     args.UserCookieID,
			FirstSourceURL:   args.FirstSourceURL,
			LastSourceURL:    args.LastSourceURL,
			Source:           args.Source,
			Argument:         argumentPayload,
			EvaluatedFlagSet: featureflag.GetEvaluatedFlagSet(ctx),
			CohortID:         args.CohortID,
			Referrer:         args.Referrer,
			PublicArgument:   publicArgumentPayload,
			UserProperties:   userPropertiesPayload,
			DeviceID:         args.DeviceID,
			EventID:          args.EventID,
			InsertID:         args.InsertID,
		})
	}

	if err := usagestats.LogEvents(ctx, r.db, events); err != nil {
		return nil, err
	}

	return nil, nil
}

var (
	searchLatenciesFrontendCodeLoad = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_search_latency_frontend_code_load_seconds",
		Help:    "Milliseconds the webapp frontend spends waiting for search result code snippets to load.",
		Buckets: trace.UserLatencyBuckets,
	}, nil)
	searchLatenciesFrontendFirstResult = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_search_latency_frontend_first_result_seconds",
		Help:    "Milliseconds the webapp frontend spends waiting for the first search result to load.",
		Buckets: trace.UserLatencyBuckets,
	}, []string{"type"})
)

// exportPrometheusSearchLatencies exports Prometheus search latency metrics given a GraphQL
// LogEvent payload.
func exportPrometheusSearchLatencies(event string, payload json.RawMessage) error {
	var v struct {
		DurationMS float64 `json:"durationMs"`
	}
	if err := json.Unmarshal([]byte(payload), &v); err != nil {
		return err
	}
	if event == "search.latencies.frontend.code-load" {
		searchLatenciesFrontendCodeLoad.WithLabelValues().Observe(v.DurationMS / 1000.0)
	}
	if strings.HasPrefix(event, "search.latencies.frontend.") && strings.HasSuffix(event, ".first-result") {
		searchType := strings.TrimSuffix(strings.TrimPrefix(event, "search.latencies.frontend."), ".first-result")
		searchLatenciesFrontendFirstResult.WithLabelValues(searchType).Observe(v.DurationMS / 1000.0)
	}
	return nil
}

var searchRankingResultClicked = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "src_search_ranking_result_clicked",
	Help:    "the index of the search result which was clicked on by the user",
	Buckets: prometheus.LinearBuckets(1, 1, 10),
}, []string{"type"})

func exportPrometheusSearchRanking(payload json.RawMessage) error {
	var v struct {
		Index float64 `json:"index"`
		Type  string  `json:"type"`
	}
	if err := json.Unmarshal([]byte(payload), &v); err != nil {
		return err
	}
	searchRankingResultClicked.WithLabelValues(v.Type).Observe(v.Index)
	return nil
}
