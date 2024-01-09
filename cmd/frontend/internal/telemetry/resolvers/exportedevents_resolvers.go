package resolvers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func decodeExportedEventsCursor(cursor string) (*time.Time, error) {
	cursor, err := graphqlutil.DecodeCursor(&cursor)
	if err != nil {
		return nil, errors.Wrap(err, "invalid cursor")
	}
	t, err := time.Parse(time.RFC3339, cursor)
	if err != nil {
		return nil, errors.Wrap(err, "invalid cursor data")
	}
	return &t, nil
}

func encodeExportedEventsCursor(t time.Time) *graphqlutil.PageInfo {
	ts, err := t.MarshalText()
	if err != nil {
		return graphqlutil.HasNextPage(false)
	}
	return graphqlutil.EncodeCursor(pointers.Ptr(string(ts)))
}

type ExportedEventResolver struct {
	event database.ExportedTelemetryEvent
}

var _ graphqlbackend.ExportedEventResolver = &ExportedEventResolver{}

func (r *ExportedEventResolver) ID() graphql.ID {
	return relay.MarshalID("ExportedEvent", r.event.ID)
}

func (r *ExportedEventResolver) ExportedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.event.ExportedAt}
}

func (r *ExportedEventResolver) Payload() (graphqlbackend.JSONValue, error) {
	payload, err := protojson.Marshal(r.event.Payload)
	if err != nil {
		return graphqlbackend.JSONValue{Value: struct{}{}},
			errors.Wrapf(err, "failed to marshal payload of event ID %q", r.event.ID)
	}
	return graphqlbackend.JSONValue{Value: json.RawMessage(payload)}, nil
}

type ExportedEventsConnectionResolver struct {
	ctx         context.Context
	diagnostics database.TelemetryEventsExportQueueDiagnosticsStore

	limit    int
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
	if len(r.exported) == 0 || len(r.exported) < r.limit {
		return graphqlutil.HasNextPage(false)
	}
	lastEvent := r.exported[len(r.exported)-1]
	return encodeExportedEventsCursor(lastEvent.Timestamp)
}
