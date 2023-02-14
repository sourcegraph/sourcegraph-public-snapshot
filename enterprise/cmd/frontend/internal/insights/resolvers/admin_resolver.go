package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/scheduler"
	insightsstore "github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ graphqlbackend.InsightSeriesMetadataPayloadResolver = &insightSeriesMetadataPayloadResolver{}
var _ graphqlbackend.InsightSeriesMetadataResolver = &insightSeriesMetadataResolver{}
var _ graphqlbackend.InsightSeriesQueryStatusResolver = &insightSeriesQueryStatusResolver{}

func (r *Resolver) UpdateInsightSeries(ctx context.Context, args *graphqlbackend.UpdateInsightSeriesArgs) (graphqlbackend.InsightSeriesMetadataPayloadResolver, error) {
	actr := actor.FromContext(ctx)
	if err := auth.CheckUserIsSiteAdmin(ctx, r.postgresDB, actr.UID); err != nil {
		return nil, err
	}

	if args.Input.Enabled != nil {
		err := r.dataSeriesStore.SetSeriesEnabled(ctx, args.Input.SeriesId, *args.Input.Enabled)
		if err != nil {
			return nil, err
		}
	}

	series, err := r.dataSeriesStore.GetDataSeries(ctx, insightsstore.GetDataSeriesArgs{IncludeDeleted: true, SeriesID: args.Input.SeriesId})
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
	if err := auth.CheckUserIsSiteAdmin(ctx, r.postgresDB, actr.UID); err != nil {
		return nil, err
	}

	// this will get the queue information from the primary postgres database
	seriesStatus, err := queryrunner.QueryAllSeriesStatus(ctx, r.workerBaseStore)
	if err != nil {
		return nil, err
	}

	// need to do a manual join with metadata since this lives in a separate database.
	seriesMetadata, err := r.dataSeriesStore.GetDataSeries(ctx, insightsstore.GetDataSeriesArgs{IncludeDeleted: true})
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

func (r *Resolver) InsightViewDebug(ctx context.Context, args graphqlbackend.InsightViewDebugArgs) (graphqlbackend.InsightViewDebugResolver, error) {
	actr := actor.FromContext(ctx)
	if err := auth.CheckUserIsSiteAdmin(ctx, r.postgresDB, actr.UID); err != nil {
		return nil, err
	}
	var viewId string
	err := relay.UnmarshalSpec(args.Id, &viewId)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling the insight view id")
	}

	// 🚨 SECURITY: This debug resolver is restricted to admins only so looking up the series does not check for the users authorization
	viewSeries, err := r.insightStore.Get(ctx, insightsstore.InsightQueryArgs{UniqueID: viewId, WithoutAuthorization: true})
	if err != nil {
		return nil, err
	}

	resolver := &insightViewDebugResolver{
		insightViewID:   viewId,
		viewSeries:      viewSeries,
		workerBaseStore: r.workerBaseStore,
		backfillStore:   scheduler.NewBackfillStore(r.insightsDB),
	}
	return resolver, nil
}

type insightSeriesMetadataPayloadResolver struct {
	series *types.InsightSeries
}

func (i *insightSeriesMetadataPayloadResolver) Series(_ context.Context) graphqlbackend.InsightSeriesMetadataResolver {
	return &insightSeriesMetadataResolver{series: i.series}
}

type insightSeriesMetadataResolver struct {
	series *types.InsightSeries
}

func (i *insightSeriesMetadataResolver) SeriesId(_ context.Context) (string, error) {
	return i.series.SeriesID, nil
}

func (i *insightSeriesMetadataResolver) Query(_ context.Context) (string, error) {
	return i.series.Query, nil
}

func (i *insightSeriesMetadataResolver) Enabled(_ context.Context) (bool, error) {
	return i.series.Enabled, nil
}

type insightSeriesQueryStatusResolver struct {
	status types.InsightSeriesStatus
}

func (i *insightSeriesQueryStatusResolver) SeriesId(_ context.Context) (string, error) {
	return i.status.SeriesId, nil
}

func (i *insightSeriesQueryStatusResolver) Query(_ context.Context) (string, error) {
	return i.status.Query, nil
}

func (i *insightSeriesQueryStatusResolver) Enabled(_ context.Context) (bool, error) {
	return i.status.Enabled, nil
}

func (i *insightSeriesQueryStatusResolver) Errored(_ context.Context) (int32, error) {
	return int32(i.status.Errored), nil
}

func (i *insightSeriesQueryStatusResolver) Completed(_ context.Context) (int32, error) {
	return int32(i.status.Completed), nil
}

func (i *insightSeriesQueryStatusResolver) Processing(_ context.Context) (int32, error) {
	return int32(i.status.Processing), nil
}

func (i *insightSeriesQueryStatusResolver) Failed(_ context.Context) (int32, error) {
	return int32(i.status.Failed), nil
}

func (i *insightSeriesQueryStatusResolver) Queued(_ context.Context) (int32, error) {
	return int32(i.status.Queued), nil
}

type insightViewDebugResolver struct {
	insightViewID   string
	viewSeries      []types.InsightViewSeries
	workerBaseStore *basestore.Store
	backfillStore   *scheduler.BackfillStore
}

func (r *insightViewDebugResolver) Raw(ctx context.Context) ([]string, error) {
	type queueDebug struct {
		types.InsightSeriesStatus
		searchErrors []types.InsightSearchFailure
	}

	type insightDebugInfo struct {
		QueueStatus queueDebug
		Backfills   []scheduler.SeriesBackfillDebug
	}

	ids := make([]string, 0, len(r.viewSeries))
	for i := 0; i < len(r.viewSeries); i++ {
		ids = append(ids, r.viewSeries[i].SeriesID)
	}

	// this will get the queue information from the primary postgres database
	seriesStatus, err := queryrunner.QuerySeriesStatus(ctx, r.workerBaseStore, ids)
	if err != nil {
		return nil, err
	}

	// index the metadata by seriesId to perform lookups
	queueStatusBySeries := make(map[string]*types.InsightSeriesStatus)
	for i, status := range seriesStatus {
		queueStatusBySeries[status.SeriesId] = &seriesStatus[i]
	}

	var viewDebug []string
	// we will treat the results from the queue as the "secondary" and left join it to the series metadata.

	for _, series := range r.viewSeries {
		// Build the Queue Info
		status := types.InsightSeriesStatus{
			SeriesId: series.SeriesID,
			Query:    series.Query,
			Enabled:  true,
		}
		if tmpStatus, ok := queueStatusBySeries[series.SeriesID]; ok {
			status.Completed = tmpStatus.Completed
			status.Enabled = tmpStatus.Enabled
			status.Errored = tmpStatus.Errored
			status.Failed = tmpStatus.Failed
			status.Queued = tmpStatus.Queued
			status.Processing = tmpStatus.Processing
		}
		seriesErrors, err := queryrunner.QuerySeriesSearchFailures(ctx, r.workerBaseStore, series.SeriesID, 100)
		if err != nil {
			return nil, err
		}

		// Build the Backfill Info
		backfillDebugInfo, err := r.backfillStore.LoadSeriesBackfillsDebugInfo(ctx, series.InsightSeriesID)
		if err != nil {
			return nil, err
		}

		seriesDebug := insightDebugInfo{
			QueueStatus: queueDebug{
				searchErrors:        seriesErrors,
				InsightSeriesStatus: status,
			},
			Backfills: backfillDebugInfo,
		}
		debugJson, err := json.Marshal(seriesDebug)
		if err != nil {
			return nil, err
		}
		viewDebug = append(viewDebug, string(debugJson))

	}
	return viewDebug, nil
}

func (r *Resolver) InsightAdminBackfillQueue(ctx context.Context, args *graphqlbackend.AdminBackfillQueueArgs) (*graphqlutil.ConnectionResolver[*graphqlbackend.BackfillQueueItemResolver], error) {
	// 🚨 SECURITY
	// only admin users can access this resolver
	actr := actor.FromContext(ctx)
	if err := auth.CheckUserIsSiteAdmin(ctx, r.postgresDB, actr.UID); err != nil {
		return nil, err
	}
	store := &adminBackfillQueueConnectionStore{
		args:          args,
		backfillStore: scheduler.NewBackfillStore(r.insightsDB),
		logger:        r.logger.Scoped("backfillqueue", "insights admin backfill queue resolver"),
	}

	// `STATE` is the default enum value in the graphql schema.
	orderBy := "STATE"
	if args.OrderBy != "" {
		orderBy = args.OrderBy
	}

	resolver, err := graphqlutil.NewConnectionResolver[*graphqlbackend.BackfillQueueItemResolver](
		store,
		&args.ConnectionResolverArgs,
		&graphqlutil.ConnectionResolverOptions{
			OrderBy: database.OrderBy{
				{Field: string(orderByToDBBackfillColumn(orderBy))},
				{Field: string(scheduler.BackfillID)},
			},
			Ascending: !args.Descending})
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

type adminBackfillQueueConnectionStore struct {
	backfillStore *scheduler.BackfillStore
	logger        log.Logger
	args          *graphqlbackend.AdminBackfillQueueArgs
}

// ComputeTotal returns the total count of all the items in the connection, independent of pagination arguments.
func (a *adminBackfillQueueConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	filterArgs := scheduler.BackfillQueueArgs{}
	if a.args != nil {
		filterArgs.States = a.args.States
		filterArgs.TextSearch = a.args.TextSearch
	}

	count, err := a.backfillStore.GetBackfillQueueTotalCount(ctx, filterArgs)
	if err != nil {
		return nil, err
	}
	return i32Ptr(&count), nil
}

func (a *adminBackfillQueueConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*graphqlbackend.BackfillQueueItemResolver, error) {
	filterArgs := scheduler.BackfillQueueArgs{PaginationArgs: args}
	if a.args != nil {
		filterArgs.States = a.args.States
		filterArgs.TextSearch = a.args.TextSearch
	}
	backfillItems, err := a.backfillStore.GetBackfillQueueInfo(ctx, filterArgs)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*graphqlbackend.BackfillQueueItemResolver, 0, len(backfillItems))
	for _, item := range backfillItems {
		resolvers = append(resolvers, &graphqlbackend.BackfillQueueItemResolver{
			BackfillID:   item.ID,
			InsightTitle: item.InsightTitle,
			CreatorID:    nil,
			Label:        item.SeriesLabel,
			Query:        item.SeriesSearchQuery,
			BackfillStatus: &backfillStatusResolver{
				queueItem: item,
			},
		})
	}

	return resolvers, nil
}

// MarshalCursor returns cursor for a node and is called for generating start and end cursors.
func (a *adminBackfillQueueConnectionStore) MarshalCursor(node *graphqlbackend.BackfillQueueItemResolver, orderBy database.OrderBy) (*string, error) {
	// This is the enum the client requested ordering by
	column := orderBy[0].Field
	var value string

	switch scheduler.BackfillQueueColumn(column) {
	case scheduler.State:
		value = strings.ToLower(node.BackfillStatus.State())
	case scheduler.QueuePosition:
		pos := node.BackfillStatus.QueuePosition()
		if pos != nil {
			value = fmt.Sprintf("%d", pos)
		} else {
			value = "NULL"
		}
	default:
		return nil, errors.New(fmt.Sprintf("invalid OrderBy.Field. Expected: one of (STATE, QUEUE_POSITION). Actual: %s", column))
	}

	// format of the "Value" is the value for the sorted by column ie `QUEUED` followed by the ID of the current node
	cursor := marshalBackfillItemCursor(
		&itypes.Cursor{
			Column: string(dbToOrderBy(scheduler.BackfillQueueColumn(column))),
			Value:  fmt.Sprintf("%s@%d", value, node.IDInt32()),
		},
	)

	return &cursor, nil

}

// UnmarshalCursor returns node id from after/before cursor string.
func (a *adminBackfillQueueConnectionStore) UnmarshalCursor(cursor string, orderBy database.OrderBy) (*string, error) {
	backfillCursor, err := unmarshalBackfillItemCursor(&cursor)
	if err != nil {
		return nil, err
	}

	orderByColumn := scheduler.BackfillQueueColumn(orderBy[0].Field)
	cursorColumn := orderByToDBBackfillColumn(backfillCursor.Column)
	if cursorColumn != orderByColumn {
		return nil, errors.New("Invalid cursor. Expected one of (STATE, QUEUE_POSITION)")
	}

	csv := ""
	values := strings.Split(backfillCursor.Value, "@")
	if len(values) != 2 {
		return nil, errors.New("Invalid cursor. Expected Value: <orderbyvalue>@<id>")
	}
	switch orderByColumn {
	case scheduler.State:
		csv = fmt.Sprintf("'%v', %v", values[0], values[1])
	case scheduler.BackfillID, scheduler.QueuePosition:
		csv = fmt.Sprintf("%v, %v", values[0], values[1])
	default:
		return nil, errors.New("Invalid OrderBy Field.")
	}

	return &csv, err
}

const backfillCursorKind = "InsightsAdminBackfillItem"

func marshalBackfillItemCursor(cursor *itypes.Cursor) string {
	return string(relay.MarshalID(backfillCursorKind, cursor))
}

func unmarshalBackfillItemCursor(cursor *string) (*itypes.Cursor, error) {
	if cursor == nil {
		return nil, nil
	}
	if kind := relay.UnmarshalKind(graphql.ID(*cursor)); kind != backfillCursorKind {
		return nil, errors.Errorf("cannot unmarshal repository cursor type: %q", kind)
	}
	var spec *itypes.Cursor
	if err := relay.UnmarshalSpec(graphql.ID(*cursor), &spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func i32Ptr(n *int) *int32 {
	if n != nil {
		tmp := int32(*n)
		return &tmp
	}
	return nil
}

type backfillStatusResolver struct {
	queueItem scheduler.BackfillQueueItem
}

func (r *backfillStatusResolver) State() string {
	return strings.ToUpper(r.queueItem.BackfillState)
}

func (r *backfillStatusResolver) QueuePosition() *int32 {
	return i32Ptr(r.queueItem.QueuePosition)
}

func (r *backfillStatusResolver) Cost() *int32 {
	return i32Ptr(r.queueItem.BackfillCost)
}

func (r *backfillStatusResolver) PercentComplete() *int32 {
	return i32Ptr(r.queueItem.PercentComplete)
}

func (r *backfillStatusResolver) CreatedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.queueItem.BackfillCreatedAt)
}

func (r *backfillStatusResolver) StartedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.queueItem.BackfillStartedAt)
}

func (r *backfillStatusResolver) CompletedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.queueItem.BackfillCompletedAt)
}
func (r *backfillStatusResolver) Errors() *[]string {
	return r.queueItem.Errors
}

func (r *backfillStatusResolver) Runtime() *string {
	if r.queueItem.RuntimeDuration != nil {
		tmp := r.queueItem.RuntimeDuration.String()
		return &tmp
	}
	return nil
}

func orderByToDBBackfillColumn(ob string) scheduler.BackfillQueueColumn {
	switch ob {
	case "STATE":
		return scheduler.State
	case "QUEUE_POSITION":
		return scheduler.QueuePosition
	default:
		return ""
	}
}

func dbToOrderBy(dbField scheduler.BackfillQueueColumn) scheduler.BackfillQueueColumn {
	switch dbField {
	case scheduler.State:
		return "STATE"
	case scheduler.QueuePosition:
		return "QUEUE_POSITION"
	default:
		return "STATE" //default
	}
}
