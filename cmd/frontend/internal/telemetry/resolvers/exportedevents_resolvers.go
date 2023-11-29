package resolvers

import (
	"context"
	"encoding/json"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type ExportedEventResolver struct {
	event database.ExportedTelemetryEvent
}

var _ graphqlbackend.ExportedEventResolver = &ExportedEventResolver{}

func (r *ExportedEventResolver) ID() graphql.ID {
	return relay.MarshalID("ExportedEvent", r.event.ID)
}

func (r *ExportedEventResolver) ExportedAt() *graphql.Time {
	return &graphql.Time{Time: r.event.ExportedAt}
}

func (r *ExportedEventResolver) Payload() (json.RawMessage, error) {
	payload, err := protojson.Marshal(r.event.Payload)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal payload of event ID %q", r.event.ID)
	}
	return json.RawMessage(payload), nil
}

type ExportedEventsConnectionResolver struct {
	ctx         context.Context
	diagnostics database.TelemetryEventsExportQueueDiagnosticsStore

	exported []database.ExportedTelemetryEvent
}

var _ graphqlbackend.ExportedEventsConnectionResolver = &ExportedEventsConnectionResolver{}

func (r *ExportedEventsConnectionResolver) Nodes() []graphqlbackend.ExportedEventResolver {
	nodes := make([]graphqlbackend.ExportedEventResolver, len(r.exported))
	for i, event := range r.exported {
		nodes[i] = &ExportedEventResolver{event: event}
	}
	return nodes
}

func (r *ExportedEventsConnectionResolver) TotalCount() (int32, error) {
	count, err := r.diagnostics.CountRecentlyExported(r.ctx)
	if err != nil {
		return 0, errors.Wrap(err, "CountRecentlyExported")
	}
	return int32(count), nil
}

func (r *ExportedEventsConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	if len(r.exported) == 0 {
		return graphqlutil.HasNextPage(false)
	}
	lastEvent := r.exported[len(r.exported)-1]
	ts, err := lastEvent.Timestamp.MarshalText()
	if err != nil {
		return graphqlutil.HasNextPage(false)
	}
	return graphqlutil.EncodeCursor(pointers.Ptr(string(ts)))
}
