package resolvers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/teestore"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Resolver is the GraphQL resolver of all things related to telemetry V2.
type Resolver struct {
	logger   log.Logger
	db       database.DB
	teestore *teestore.Store
}

var _ graphqlbackend.TelemetryResolver = &Resolver{}

// New returns a new Resolver whose store uses the given database
func New(logger log.Logger, db database.DB) graphqlbackend.TelemetryResolver {
	return &Resolver{
		logger:   logger,
		db:       db,
		teestore: teestore.NewStore(db.TelemetryEventsExportQueue(), db.EventLogs()),
	}
}

func (r *Resolver) ExportedEvents(ctx context.Context, args *graphqlbackend.ExportedEventsArgs) (graphqlbackend.ExportedEventsConnectionResolver, error) {
	// 🚨 SECURITY: Caller must be a site admin.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	first := int(args.First)
	if first == 0 {
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

func (r *Resolver) RecordEvents(ctx context.Context, args *graphqlbackend.RecordEventsArgs) (*graphqlbackend.EmptyResponse, error) {
	if args == nil || len(args.Events) == 0 {
		return nil, errors.New("no events provided")
	}
	gatewayEvents, err := newTelemetryGatewayEvents(ctx, time.Now(), telemetrygatewayv1.DefaultEventIDFunc, args.Events)
	if err != nil {
		// This is an important failure, make sure we surface it, as it could be
		// an implementation error.
		data, _ := json.Marshal(args.Events)
		r.logger.Error("failed to convert telemetry events to internal format",
			log.Error(err),
			log.String("eventData", string(data)))
		return nil, errors.Wrap(err, "invalid events provided")
	}
	if err := r.teestore.StoreEvents(ctx, gatewayEvents); err != nil {
		return nil, errors.Wrap(err, "error storing events")
	}
	return &graphqlbackend.EmptyResponse{}, nil
}
