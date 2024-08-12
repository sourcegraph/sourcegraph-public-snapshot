package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// This file just contains stub GraphQL resolvers and data types for Code Insights which merely
// return an error if not running in enterprise mode. The actual resolvers can be found in
// internal/insights/resolvers

// InsightsResolver is the root resolver.
type InsightsResolver interface {
	// Queries
	InsightsDashboards(ctx context.Context, args *InsightsDashboardsArgs) (InsightsDashboardConnectionResolver, error)
	InsightViews(ctx context.Context, args *InsightViewQueryArgs) (InsightViewConnectionResolver, error)

	SearchInsightLivePreview(ctx context.Context, args SearchInsightLivePreviewArgs) ([]SearchInsightLivePreviewSeriesResolver, error)
	SearchInsightPreview(ctx context.Context, args SearchInsightPreviewArgs) ([]SearchInsightLivePreviewSeriesResolver, error)

	ValidateScopedInsightQuery(ctx context.Context, args ValidateScopedInsightQueryArgs) (ScopedInsightQueryPayloadResolver, error)
	PreviewRepositoriesFromQuery(ctx context.Context, args PreviewRepositoriesFromQueryArgs) (RepositoryPreviewPayloadResolver, error)

	// Mutations
	CreateInsightsDashboard(ctx context.Context, args *CreateInsightsDashboardArgs) (InsightsDashboardPayloadResolver, error)
	UpdateInsightsDashboard(ctx context.Context, args *UpdateInsightsDashboardArgs) (InsightsDashboardPayloadResolver, error)
	DeleteInsightsDashboard(ctx context.Context, args *DeleteInsightsDashboardArgs) (*EmptyResponse, error)
	RemoveInsightViewFromDashboard(ctx context.Context, args *RemoveInsightViewFromDashboardArgs) (InsightsDashboardPayloadResolver, error)
	AddInsightViewToDashboard(ctx context.Context, args *AddInsightViewToDashboardArgs) (InsightsDashboardPayloadResolver, error)

	CreateLineChartSearchInsight(ctx context.Context, args *CreateLineChartSearchInsightArgs) (InsightViewPayloadResolver, error)
	UpdateLineChartSearchInsight(ctx context.Context, args *UpdateLineChartSearchInsightArgs) (InsightViewPayloadResolver, error)
	CreatePieChartSearchInsight(ctx context.Context, args *CreatePieChartSearchInsightArgs) (InsightViewPayloadResolver, error)
	UpdatePieChartSearchInsight(ctx context.Context, args *UpdatePieChartSearchInsightArgs) (InsightViewPayloadResolver, error)

	DeleteInsightView(ctx context.Context, args *DeleteInsightViewArgs) (*EmptyResponse, error)
	SaveInsightAsNewView(ctx context.Context, args SaveInsightAsNewViewArgs) (InsightViewPayloadResolver, error)

	// Admin Management
	InsightSeriesQueryStatus(ctx context.Context) ([]InsightSeriesQueryStatusResolver, error)
	InsightViewDebug(ctx context.Context, args InsightViewDebugArgs) (InsightViewDebugResolver, error)
	InsightAdminBackfillQueue(ctx context.Context, args *AdminBackfillQueueArgs) (*gqlutil.ConnectionResolver[*BackfillQueueItemResolver], error)
	// Admin Mutations
	UpdateInsightSeries(ctx context.Context, args *UpdateInsightSeriesArgs) (InsightSeriesMetadataPayloadResolver, error)
	RetryInsightSeriesBackfill(ctx context.Context, args *BackfillArgs) (*BackfillQueueItemResolver, error)
	MoveInsightSeriesBackfillToFrontOfQueue(ctx context.Context, args *BackfillArgs) (*BackfillQueueItemResolver, error)
	MoveInsightSeriesBackfillToBackOfQueue(ctx context.Context, args *BackfillArgs) (*BackfillQueueItemResolver, error)
}

type SearchInsightLivePreviewArgs struct {
	Input SearchInsightLivePreviewInput
}
type SearchInsightPreviewArgs struct {
	Input SearchInsightPreviewInput
}

type SearchInsightPreviewInput struct {
	RepositoryScope RepositoryScopeInput
	TimeScope       TimeScopeInput
	Series          []SearchSeriesPreviewInput
}

type SearchSeriesPreviewInput struct {
	Query                      string
	Label                      string
	GeneratedFromCaptureGroups bool
	GroupBy                    *string
}

type SearchInsightLivePreviewInput struct {
	Query                      string
	Label                      string
	RepositoryScope            RepositoryScopeInput
	TimeScope                  TimeScopeInput
	GeneratedFromCaptureGroups bool
	GroupBy                    *string
}

type InsightsArgs struct {
	Ids *[]graphql.ID
}

type InsightViewDebugArgs struct {
	Id graphql.ID
}

type InsightsDataPointResolver interface {
	DateTime() gqlutil.DateTime
	Value() float64
	DiffQuery() (*string, error)
	PointInTimeQuery() (*string, error)
}

type InsightViewDebugResolver interface {
	Raw(context.Context) ([]string, error)
}

type InsightStatusResolver interface {
	TotalPoints(context.Context) (int32, error)
	PendingJobs(context.Context) (int32, error)
	CompletedJobs(context.Context) (int32, error)
	FailedJobs(context.Context) (int32, error)
	BackfillQueuedAt(context.Context) *gqlutil.DateTime
	IsLoadingData(context.Context) (*bool, error)
	IncompleteDatapoints(ctx context.Context) ([]IncompleteDatapointAlert, error)
}

type InsightsPointsArgs struct {
	From             *gqlutil.DateTime
	To               *gqlutil.DateTime
	IncludeRepoRegex *string
	ExcludeRepoRegex *string
}

type InsightSeriesResolver interface {
	SeriesId() string
	Label() string
	Points(ctx context.Context, args *InsightsPointsArgs) ([]InsightsDataPointResolver, error)
	Status(ctx context.Context) (InsightStatusResolver, error)
}

type InsightResolver interface {
	Title() string
	Description() string
	Series() []InsightSeriesResolver
	ID() string
}

type InsightsDashboardsArgs struct {
	First *int32
	After *string
	ID    *graphql.ID
}

type InsightsDashboardConnectionResolver interface {
	Nodes(ctx context.Context) ([]InsightsDashboardResolver, error)
	PageInfo(ctx context.Context) (*gqlutil.PageInfo, error)
}

type InsightsDashboardResolver interface {
	Title() string
	ID() graphql.ID
	Views(ctx context.Context, args DashboardInsightViewConnectionArgs) InsightViewConnectionResolver
	Grants() InsightsPermissionGrantsResolver
}

type DashboardInsightViewConnectionArgs struct {
	After *string
	First *int32
}

type InsightsPermissionGrantsResolver interface {
	Users() []graphql.ID
	Organizations() []graphql.ID
	Global() bool
}

type CreateInsightsDashboardArgs struct {
	Input CreateInsightsDashboardInput
}

type CreateInsightsDashboardInput struct {
	Title  string
	Grants InsightsPermissionGrants
}

type UpdateInsightsDashboardArgs struct {
	Id    graphql.ID
	Input UpdateInsightsDashboardInput
}

type UpdateInsightsDashboardInput struct {
	Title  *string
	Grants *InsightsPermissionGrants
}

type InsightsPermissionGrants struct {
	Users         *[]graphql.ID
	Organizations *[]graphql.ID
	Global        *bool
}

type DeleteInsightsDashboardArgs struct {
	Id graphql.ID
}

type InsightViewConnectionResolver interface {
	Nodes(ctx context.Context) ([]InsightViewResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*gqlutil.PageInfo, error)
}

type InsightViewResolver interface {
	ID() graphql.ID
	DefaultFilters(ctx context.Context) (InsightViewFiltersResolver, error)
	AppliedFilters(ctx context.Context) (InsightViewFiltersResolver, error)
	DataSeries(ctx context.Context) ([]InsightSeriesResolver, error)
	Presentation(ctx context.Context) (InsightPresentation, error)
	DataSeriesDefinitions(ctx context.Context) ([]InsightDataSeriesDefinition, error)
	DashboardReferenceCount(ctx context.Context) (int32, error)
	IsFrozen(ctx context.Context) (bool, error)
	DefaultSeriesDisplayOptions(ctx context.Context) (InsightViewSeriesDisplayOptionsResolver, error)
	AppliedSeriesDisplayOptions(ctx context.Context) (InsightViewSeriesDisplayOptionsResolver, error)
	Dashboards(ctx context.Context, args *InsightsDashboardsArgs) InsightsDashboardConnectionResolver
	SeriesCount(ctx context.Context) (*int32, error)
	RepositoryDefinition(ctx context.Context) (InsightRepositoryDefinition, error)
	TimeScope(ctx context.Context) (InsightTimeScope, error)
}

type InsightDataSeriesDefinition interface {
	ToSearchInsightDataSeriesDefinition() (SearchInsightDataSeriesDefinitionResolver, bool)
}

type LineChartInsightViewPresentation interface {
	Title(ctx context.Context) (string, error)
	SeriesPresentation(ctx context.Context) ([]LineChartDataSeriesPresentationResolver, error)
}

type PieChartInsightViewPresentation interface {
	Title(ctx context.Context) (string, error)
	OtherThreshold(ctx context.Context) (float64, error)
}

type LineChartDataSeriesPresentationResolver interface {
	SeriesId(ctx context.Context) (string, error)
	Label(ctx context.Context) (string, error)
	Color(ctx context.Context) (string, error)
}

type SearchInsightDataSeriesDefinitionResolver interface {
	SeriesId(ctx context.Context) (string, error)
	Query(ctx context.Context) (string, error)
	RepositoryScope(ctx context.Context) (InsightRepositoryScopeResolver, error)
	RepositoryDefinition(ctx context.Context) (InsightRepositoryDefinition, error)
	TimeScope(ctx context.Context) (InsightTimeScope, error)
	GeneratedFromCaptureGroups() (bool, error)
	IsCalculated() (bool, error)
	GroupBy() (*string, error)
}

type InsightPresentation interface {
	ToLineChartInsightViewPresentation() (LineChartInsightViewPresentation, bool)
	ToPieChartInsightViewPresentation() (PieChartInsightViewPresentation, bool)
}

type InsightTimeScope interface {
	ToInsightIntervalTimeScope() (InsightIntervalTimeScope, bool)
}

type InsightIntervalTimeScope interface {
	Unit(ctx context.Context) (string, error)
	Value(ctx context.Context) (int32, error)
}

type InsightRepositoryScopeResolver interface {
	Repositories(ctx context.Context) ([]string, error)
}

type InsightRepositoryDefinition interface {
	ToInsightRepositoryScope() (InsightRepositoryScopeResolver, bool)
	ToRepositorySearchScope() (RepositorySearchScopeResolver, bool)
}

type RepositorySearchScopeResolver interface {
	Search() string
	AllRepositories() bool
}

type InsightsDashboardPayloadResolver interface {
	Dashboard(ctx context.Context) (InsightsDashboardResolver, error)
}

type AddInsightViewToDashboardArgs struct {
	Input AddInsightViewToDashboardInput
}

type AddInsightViewToDashboardInput struct {
	InsightViewID graphql.ID
	DashboardID   graphql.ID
}

type RemoveInsightViewFromDashboardArgs struct {
	Input RemoveInsightViewFromDashboardInput
}

type RemoveInsightViewFromDashboardInput struct {
	InsightViewID graphql.ID
	DashboardID   graphql.ID
}

type UpdateInsightSeriesArgs struct {
	Input UpdateInsightSeriesInput
}

type UpdateInsightSeriesInput struct {
	SeriesId string
	Enabled  *bool
}

type InsightSeriesMetadataResolver interface {
	SeriesId(ctx context.Context) (string, error)
	Query(ctx context.Context) (string, error)
	Enabled(ctx context.Context) (bool, error)
}

type InsightSeriesMetadataPayloadResolver interface {
	Series(ctx context.Context) InsightSeriesMetadataResolver
}

type InsightSeriesQueryStatusResolver interface {
	SeriesId(ctx context.Context) (string, error)
	Query(ctx context.Context) (string, error)
	Enabled(ctx context.Context) (bool, error)
	Errored(ctx context.Context) (int32, error)
	Completed(ctx context.Context) (int32, error)
	Processing(ctx context.Context) (int32, error)
	Failed(ctx context.Context) (int32, error)
	Queued(ctx context.Context) (int32, error)
}
type InsightViewFiltersResolver interface {
	IncludeRepoRegex(ctx context.Context) (*string, error)
	ExcludeRepoRegex(ctx context.Context) (*string, error)
	SearchContexts(ctx context.Context) (*[]string, error)
}
type InsightViewSeriesDisplayOptionsResolver interface {
	SortOptions(ctx context.Context) (InsightViewSeriesSortOptionsResolver, error)
	Limit(ctx context.Context) (*int32, error)
	NumSamples() *int32
}

type InsightViewSeriesSortOptionsResolver interface {
	Mode(ctx context.Context) (*string, error)
	Direction(ctx context.Context) (*string, error)
}

type CreateLineChartSearchInsightArgs struct {
	Input CreateLineChartSearchInsightInput
}

type CreateLineChartSearchInsightInput struct {
	DataSeries      []LineChartSearchInsightDataSeriesInput
	Options         LineChartOptionsInput
	Dashboards      *[]graphql.ID
	ViewControls    *InsightViewControlsInput
	RepositoryScope *RepositoryScopeInput
	TimeScope       *TimeScopeInput
}

type UpdateLineChartSearchInsightArgs struct {
	Id    graphql.ID
	Input UpdateLineChartSearchInsightInput
}

type UpdateLineChartSearchInsightInput struct {
	DataSeries          []LineChartSearchInsightDataSeriesInput
	PresentationOptions LineChartOptionsInput
	ViewControls        InsightViewControlsInput
	RepositoryScope     *RepositoryScopeInput
	TimeScope           *TimeScopeInput
}

type CreatePieChartSearchInsightArgs struct {
	Input CreatePieChartSearchInsightInput
}

type CreatePieChartSearchInsightInput struct {
	Query               string
	RepositoryScope     RepositoryScopeInput
	PresentationOptions PieChartOptionsInput
	Dashboards          *[]graphql.ID
}

type UpdatePieChartSearchInsightArgs struct {
	Id    graphql.ID
	Input UpdatePieChartSearchInsightInput
}

type UpdatePieChartSearchInsightInput struct {
	Query               string
	RepositoryScope     RepositoryScopeInput
	PresentationOptions PieChartOptionsInput
}

type PieChartOptionsInput struct {
	Title          string
	OtherThreshold float64
}

type InsightViewControlsInput struct {
	Filters              InsightViewFiltersInput
	SeriesDisplayOptions SeriesDisplayOptionsInput
}

type SeriesDisplayOptions struct {
	SortOptions *SeriesSortOptions
	Limit       *int32
}

type SeriesDisplayOptionsInput struct {
	SortOptions *SeriesSortOptionsInput
	Limit       *int32
	NumSamples  *int32
}

type SeriesSortOptions struct {
	Mode      *string // enum
	Direction *string // enum
}

type SeriesSortOptionsInput struct {
	Mode      string // enum
	Direction string // enum
}

type InsightViewFiltersInput struct {
	IncludeRepoRegex *string
	ExcludeRepoRegex *string
	SearchContexts   *[]string
}

type LineChartSearchInsightDataSeriesInput struct {
	SeriesId                   *string
	Query                      string
	TimeScope                  *TimeScopeInput
	RepositoryScope            *RepositoryScopeInput
	Options                    LineChartDataSeriesOptionsInput
	GeneratedFromCaptureGroups *bool
	GroupBy                    *string
}

type LineChartDataSeriesOptionsInput struct {
	Label     *string
	LineColor *string
}

type RepositoryScopeInput struct {
	Repositories       []string
	RepositoryCriteria *string
}

type TimeScopeInput struct {
	StepInterval *TimeIntervalStepInput
}

type TimeIntervalStepInput struct {
	Unit  string // this is actually an enum, not sure how that works here with graphql enums
	Value int32
}

type LineChartOptionsInput struct {
	Title *string
}

type SaveInsightAsNewViewArgs struct {
	Input SaveInsightAsNewViewInput
}

type SaveInsightAsNewViewInput struct {
	InsightViewID graphql.ID
	Options       LineChartOptionsInput
	Dashboard     *graphql.ID
	ViewControls  *InsightViewControlsInput
}

type InsightViewPayloadResolver interface {
	View(ctx context.Context) (InsightViewResolver, error)
}

type InsightViewQueryArgs struct {
	First                *int32
	After                *string
	Id                   *graphql.ID
	ExcludeIds           *[]graphql.ID
	Find                 *string
	IsFrozen             *bool
	Filters              *InsightViewFiltersInput
	SeriesDisplayOptions *SeriesDisplayOptionsInput
}

type DeleteInsightViewArgs struct {
	Id graphql.ID
}

type SearchInsightLivePreviewSeriesResolver interface {
	Points(ctx context.Context) ([]InsightsDataPointResolver, error)
	Label(ctx context.Context) (string, error)
}

type IncompleteDatapointAlert interface {
	ToTimeoutDatapointAlert() (TimeoutDatapointAlert, bool)
	ToGenericIncompleteDatapointAlert() (GenericIncompleteDatapointAlert, bool)
	Time() gqlutil.DateTime
}

type TimeoutDatapointAlert interface {
	Time() gqlutil.DateTime
	Repositories(ctx context.Context) (*[]*RepositoryResolver, error)
}

type GenericIncompleteDatapointAlert interface {
	Time() gqlutil.DateTime
	Reason() string
	Repositories(ctx context.Context) (*[]*RepositoryResolver, error)
}

type ValidateScopedInsightQueryArgs struct {
	Query string
}

type ScopedInsightQueryPayloadResolver interface {
	Query(ctx context.Context) string
	IsValid(ctx context.Context) bool
	InvalidReason(ctx context.Context) *string
}

type PreviewRepositoriesFromQueryArgs struct {
	Query string
}

type RepositoryPreviewPayloadResolver interface {
	Query(ctx context.Context) string
	NumberOfRepositories(ctx context.Context) *int32
}

type BackfillQueueID struct {
	BackfillID int
	InsightID  string
}
type BackfillQueueItemResolver struct {
	BackfillID      int
	InsightUniqueID string
	InsightTitle    string
	CreatorID       *int32
	Label           string
	Query           string
	BackfillStatus  BackfillQueueStatusResolver
	GetUserResolver func(*int32) (*UserResolver, error)
}

func (r *BackfillQueueItemResolver) ID() graphql.ID {
	return relay.MarshalID("backfillQueueItem", BackfillQueueID{BackfillID: r.BackfillID, InsightID: r.InsightUniqueID})
}

func (r *BackfillQueueItemResolver) IDInt32() int32 {
	return int32(r.BackfillID)
}

func (r *BackfillQueueItemResolver) InsightViewTitle() string {
	return r.InsightTitle
}
func (r *BackfillQueueItemResolver) Creator(ctx context.Context) (*UserResolver, error) {
	return r.GetUserResolver(r.CreatorID)
}
func (r *BackfillQueueItemResolver) SeriesLabel() string {
	return r.Label
}
func (r *BackfillQueueItemResolver) SeriesSearchQuery() string {
	return r.Query
}
func (r *BackfillQueueItemResolver) BackfillQueueStatus() (BackfillQueueStatusResolver, error) {
	return r.BackfillStatus, nil
}

type BackfillQueueStatusResolver interface {
	State() string // enum
	QueuePosition() *int32
	Errors() *[]string
	Cost() *int32
	PercentComplete() *int32
	CreatedAt() *gqlutil.DateTime
	StartedAt() *gqlutil.DateTime
	CompletedAt() *gqlutil.DateTime
	Runtime() *string
}

type BackfillArgs struct {
	Id graphql.ID
}

type AdminBackfillQueueArgs struct {
	gqlutil.ConnectionResolverArgs
	OrderBy    string
	Descending bool

	//filters
	States     *[]string
	TextSearch *string
}
