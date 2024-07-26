package resolvers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetrystore"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

// Resolver is the GraphQL resolver of all things related to telemetry V2.
type Resolver struct {
	logger         log.Logger
	db             database.DB
	telemetryStore telemetry.EventsStore
}

var _ graphqlbackend.TelemetryResolver = &Resolver{}

// New returns a new Resolver whose store uses the given database
func New(logger log.Logger, db database.DB) graphqlbackend.TelemetryResolver {
	return &Resolver{
		logger:         logger,
		db:             db,
		telemetryStore: telemetrystore.New(db.TelemetryEventsExportQueue(), db.EventLogs()),
	}
}

func (r *Resolver) ExportedEvents(ctx context.Context, args *graphqlbackend.ExportedEventsArgs) (graphqlbackend.ExportedEventsConnectionResolver, error) {
	// 🚨 SECURITY: Caller must be a site admin.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	first := int(args.First)
	if first <= 0 {
		first = 50
	}
	var before *time.Time
	if args.After != nil {
		var err error
		before, err = decodeExportedEventsCursor(*args.After)
		if err != nil {
			return nil, errors.Wrap(err, "invalid cursor data")
		}
	}

	exported, err := r.db.TelemetryEventsExportQueue().ListRecentlyExported(ctx, first, before)
	if err != nil {
		return nil, errors.Wrap(err, "ListRecentlyExported")
	}

	return &ExportedEventsConnectionResolver{
		ctx:         ctx,
		diagnostics: r.db.TelemetryEventsExportQueue(),
		limit:       first,
		exported:    exported,
	}, nil
}

// knownBadEvents collects 'feature':'action' combinations that produce invalid
// events we already know about and/or have already shipped fixes for, and silences
// error logs for them.
var knownBadEvents = map[string]string{
	// Noisy one fixed in https://github.com/sourcegraph/cody/pull/4077
	// VSCode 1.18+
	"cody.completion": "persistence:present",
}

func (r *Resolver) RecordEvents(ctx context.Context, args *graphqlbackend.RecordEventsArgs) (*graphqlbackend.EmptyResponse, error) {
	if args == nil || len(args.Events) == 0 {
		return nil, errors.New("no events provided")
	}

	gatewayEvents := make([]*telemetrygatewayv1.Event, 0, len(args.Events))
	for _, ev := range args.Events {
		gatewayEvent, err := convertToTelemetryGatewayEvent(ctx, time.Now(), telemetrygatewayv1.DefaultEventIDFunc, ev)
		if err != nil {
			if knownAction, ok := knownBadEvents[ev.Feature]; ok && knownAction == ev.Action {
				// We already know this event is a problem - just error out
				// without logging
				return nil, errors.Wrap(err, "known invalid event provided")
			}

			// This is an important failure, make sure we surface it, as it could be
			// an implementation error.
			data, _ := json.Marshal(args.Events)
			trace.Logger(ctx, r.logger).Error("failed to convert telemetry event to internal format",
				log.Error(err),
				log.String("eventData", string(data)))
			return nil, errors.Wrap(err, "invalid event provided")
		}
		gatewayEvents = append(gatewayEvents, gatewayEvent)
	}

	if err := r.telemetryStore.StoreEvents(ctx, gatewayEvents); err != nil {
		// This is an important failure, make sure we surface it, as it could be
		// an implementation error.
		data, _ := json.Marshal(args.Events)
		trace.Logger(ctx, r.logger).Error("error storing events",
			log.Error(err),
			log.String("eventData", string(data)))
		return nil, errors.Wrap(err, "error storing events")
	}
	return &graphqlbackend.EmptyResponse{}, nil
}
