package resolvers

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Resolver is the GraphQL resolver of all things related to telemetry V2.
type Resolver struct {
	logger log.Logger
	db     database.DB
}

// New returns a new Resolver whose store uses the given database
func New(logger log.Logger, db database.DB) graphqlbackend.TelemetryResolver {
	return &Resolver{logger: logger, db: db}
}

var _ graphqlbackend.TelemetryResolver = &Resolver{}

func (r *Resolver) RecordEvents(ctx context.Context, args *graphqlbackend.RecordEventsArgs) (*graphqlbackend.EmptyResponse, error) {
	return nil, errors.New("not implemented")
}
