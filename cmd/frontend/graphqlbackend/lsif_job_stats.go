package graphqlbackend

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/lsif"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

const singletonLSIFJobStatsGQLID = "lsifJobStats"

func (r *schemaResolver) LSIFJobStats(ctx context.Context) (*lsifJobStatsResolver, error) {
	return lsifJobStatsByGQLID(ctx, marshalLSIFJobStatsGQLID(singletonLSIFJobStatsGQLID))
}

func lsifJobStatsByGQLID(ctx context.Context, id graphql.ID) (*lsifJobStatsResolver, error) {
	lsifJobStatsID, err := unmarshalLSIFJobStatsGQLID(id)
	if err != nil {
		return nil, err
	}
	if lsifJobStatsID != singletonLSIFJobStatsGQLID {
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
	return marshalLSIFJobStatsGQLID(singletonLSIFJobStatsGQLID)
}

func (r *lsifJobStatsResolver) Active() int32    { return r.stats.Active }
func (r *lsifJobStatsResolver) Queued() int32    { return r.stats.Queued }
func (r *lsifJobStatsResolver) Scheduled() int32 { return r.stats.Scheduled }
func (r *lsifJobStatsResolver) Completed() int32 { return r.stats.Completed }
func (r *lsifJobStatsResolver) Failed() int32    { return r.stats.Failed }

func marshalLSIFJobStatsGQLID(lsifJobStatsID string) graphql.ID {
	return relay.MarshalID("LSIFJobStats", lsifJobStatsID)
}

func unmarshalLSIFJobStatsGQLID(id graphql.ID) (lsifJobStatsID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifJobStatsID)
	return
}
