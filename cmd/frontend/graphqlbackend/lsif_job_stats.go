package graphqlbackend

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/lsif"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

const lsifJobStatsGQLID = "lsifJobStats"

func (r *schemaResolver) LSIFJobStats(ctx context.Context) (*lsifJobStatsResolver, error) {
	return lsifJobStatsByGQLID(ctx, marshalLSIFJobStatsGQLID(lsifJobStatsGQLID))
}

func lsifJobStatsByGQLID(ctx context.Context, id graphql.ID) (*lsifJobStatsResolver, error) {
	lsifJobStatsID, err := unmarshalLSIFJobStatsGQLID(id)
	if err != nil {
		return nil, err
	}
	if lsifJobStatsID != lsifJobStatsGQLID {
		return nil, fmt.Errorf("lsif job stats not found: %q", lsifJobStatsID)
	}

	var stats *types.LSIFJobStats
	if err := lsif.TraceRequestAndUnmarshalPayload(ctx, "/jobs/stats", nil, &stats); err != nil {
		return nil, err
	}

	return &lsifJobStatsResolver{stats: stats}, nil
}

type lsifJobStatsResolver struct {
	stats *types.LSIFJobStats
}

func (r *lsifJobStatsResolver) ID() graphql.ID {
	return marshalLSIFJobStatsGQLID(lsifJobStatsGQLID)
}

func (r *lsifJobStatsResolver) ProcessingCount() int32 { return r.stats.ProcessingCount }
func (r *lsifJobStatsResolver) ErroredCount() int32    { return r.stats.ErroredCount }
func (r *lsifJobStatsResolver) CompletedCount() int32  { return r.stats.CompletedCount }
func (r *lsifJobStatsResolver) QueuedCount() int32     { return r.stats.QueuedCount }
func (r *lsifJobStatsResolver) ScheduledCount() int32  { return r.stats.ScheduledCount }

func marshalLSIFJobStatsGQLID(lsifJobStatsID string) graphql.ID {
	return relay.MarshalID("LSIFJobStats", lsifJobStatsID)
}

func unmarshalLSIFJobStatsGQLID(id graphql.ID) (lsifJobStatsID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifJobStatsID)
	return
}
