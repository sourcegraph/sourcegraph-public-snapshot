package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
)

var _ graphqlbackend.InsightSeriesMetadataPayloadResolver = &insightSeriesMetadataPayloadResolver{}
var _ graphqlbackend.InsightSeriesMetadataResolver = &insightSeriesMetadataResolver{}
var _ graphqlbackend.InsightSeriesQueryStatusResolver = &insightSeriesQueryStatusResolver{}

func (r *Resolver) UpdateInsightSeries(ctx context.Context, args *graphqlbackend.UpdateInsightSeriesArgs) (graphqlbackend.InsightSeriesMetadataPayloadResolver, error) {
	actr := actor.FromContext(ctx)
	if err := backend.CheckUserIsSiteAdmin(ctx, r.postgresDB, actr.UID); err != nil {
		return nil, err
	}

	if args.Input.Enabled != nil {
		err := r.dataSeriesStore.SetSeriesEnabled(ctx, args.Input.SeriesId, *args.Input.Enabled)
		if err != nil {
			return nil, err
		}
	}

	series, err := r.dataSeriesStore.GetDataSeries(ctx, store.GetDataSeriesArgs{IncludeDeleted: true, SeriesID: args.Input.SeriesId})
	if err != nil {
		return nil, err
	}
	if len(series) == 0 {
		return nil, errors.Newf("unable to fetch series with series_id: %v", args.Input.SeriesId)
	}
	return &insightSeriesMetadataPayloadResolver{series: &series[0]}, nil
}

func (r *Resolver) InsightSeriesQueryStatus(ctx context.Context) ([]graphqlbackend.InsightSeriesQueryStatusResolver, error) {
	actr := actor.FromContext(ctx)
	if err := backend.CheckUserIsSiteAdmin(ctx, r.postgresDB, actr.UID); err != nil {
		return nil, err
	}

	// this will get the queue information from the primary postgres database
	seriesStatus, err := queryrunner.QueryAllSeriesStatus(ctx, r.workerBaseStore)
	if err != nil {
		return nil, err
	}

	// need to do a manual join with metadata since this lives in a separate database.
	seriesMetadata, err := r.dataSeriesStore.GetDataSeries(ctx, store.GetDataSeriesArgs{IncludeDeleted: true})
	if err != nil {
		return nil, err
	}
	// index the metadata by seriesId to perform lookups
	metadataBySeries := make(map[string]*types.InsightSeries)
	for i, series := range seriesMetadata {
		metadataBySeries[series.SeriesID] = &seriesMetadata[i]
	}

	var resolvers []graphqlbackend.InsightSeriesQueryStatusResolver
	// we will treat the results from the queue as the "primary" and perform a left join on query metadata. That way
	// we never have a situation where we can't inspect the records in the queue, that's the entire point of this operation.
	for _, status := range seriesStatus {
		if series, ok := metadataBySeries[status.SeriesId]; ok {
			status.Query = series.Query
			status.Enabled = series.Enabled
		}

		resolvers = append(resolvers, &insightSeriesQueryStatusResolver{status: status})
	}
	return resolvers, nil
}

type insightSeriesMetadataPayloadResolver struct {
	series *types.InsightSeries
}

func (i *insightSeriesMetadataPayloadResolver) Series(ctx context.Context) graphqlbackend.InsightSeriesMetadataResolver {
	return &insightSeriesMetadataResolver{series: i.series}
}

type insightSeriesMetadataResolver struct {
	series *types.InsightSeries
}

func (i *insightSeriesMetadataResolver) SeriesId(ctx context.Context) (string, error) {
	return i.series.SeriesID, nil
}

func (i *insightSeriesMetadataResolver) Query(ctx context.Context) (string, error) {
	return i.series.Query, nil
}

func (i *insightSeriesMetadataResolver) Enabled(ctx context.Context) (bool, error) {
	return i.series.Enabled, nil
}

type insightSeriesQueryStatusResolver struct {
	status types.InsightSeriesStatus
}

func (i *insightSeriesQueryStatusResolver) SeriesId(ctx context.Context) (string, error) {
	return i.status.SeriesId, nil
}

func (i *insightSeriesQueryStatusResolver) Query(ctx context.Context) (string, error) {
	return i.status.Query, nil
}

func (i *insightSeriesQueryStatusResolver) Enabled(ctx context.Context) (bool, error) {
	return i.status.Enabled, nil
}

func (i *insightSeriesQueryStatusResolver) Errored(ctx context.Context) (int32, error) {
	return int32(i.status.Errored), nil
}

func (i *insightSeriesQueryStatusResolver) Completed(ctx context.Context) (int32, error) {
	return int32(i.status.Completed), nil
}

func (i *insightSeriesQueryStatusResolver) Processing(ctx context.Context) (int32, error) {
	return int32(i.status.Processing), nil
}

func (i *insightSeriesQueryStatusResolver) Failed(ctx context.Context) (int32, error) {
	return int32(i.status.Failed), nil
}

func (i *insightSeriesQueryStatusResolver) Queued(ctx context.Context) (int32, error) {
	return int32(i.status.Queued), nil
}
