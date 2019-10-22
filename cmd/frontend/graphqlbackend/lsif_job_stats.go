package graphqlbackend

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

const singletonLsifJobStatsGQLID = "lsifJobStats"

func (r *schemaResolver) LsifJobStats(ctx context.Context) (*lsifJobStatsResolver, error) {
	return lsifJobStatsByGQLID(ctx, marshalLsifJobStatsGQLID(singletonLsifJobStatsGQLID))
}

func lsifJobStatsByGQLID(ctx context.Context, id graphql.ID) (*lsifJobStatsResolver, error) {
	lsifJobStatsGQLID, err := unmarshalLsifJobStatsGQLID(id)
	if err != nil {
		return nil, err
	}
	if lsifJobStatsGQLID != singletonLsifJobStatsGQLID {
		return nil, fmt.Errorf("lsif job stats not found: %q", lsifJobStatsGQLID)
	}

	var stats *types.LsifJobStats
	if err := lsifRequest(ctx, "jobs/stats", nil, &stats); err != nil {
		return nil, err
	}

	return &lsifJobStatsResolver{stats: stats}, nil
}

type lsifJobStatsResolver struct {
	stats *types.LsifJobStats
}

func (r *lsifJobStatsResolver) ID() graphql.ID {

	return marshalLsifJobStatsGQLID(singletonLsifJobStatsGQLID)
}

func (r *lsifJobStatsResolver) Active() int32    { return r.stats.Active }
func (r *lsifJobStatsResolver) Queued() int32    { return r.stats.Queued }
func (r *lsifJobStatsResolver) Scheduled() int32 { return r.stats.Scheduled }
func (r *lsifJobStatsResolver) Completed() int32 { return r.stats.Completed }
func (r *lsifJobStatsResolver) Failed() int32    { return r.stats.Failed }

func marshalLsifJobStatsGQLID(lsifJobStatsID string) graphql.ID {
	return relay.MarshalID("LsifJobStats", lsifJobStatsID)
}

func unmarshalLsifJobStatsGQLID(id graphql.ID) (lsifJobStatsID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifJobStatsID)
	return
}
