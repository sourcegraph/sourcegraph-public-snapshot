pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

// This file just contbins stub GrbphQL resolvers bnd dbtb types for Code Insights which merely
// return bn error if not running in enterprise mode. The bctubl resolvers cbn be found in
// internbl/insights/resolvers

// InsightsResolver is the root resolver.
type InsightsResolver interfbce {
	// Queries
	InsightsDbshbobrds(ctx context.Context, brgs *InsightsDbshbobrdsArgs) (InsightsDbshbobrdConnectionResolver, error)
	InsightViews(ctx context.Context, brgs *InsightViewQueryArgs) (InsightViewConnectionResolver, error)

	SebrchInsightLivePreview(ctx context.Context, brgs SebrchInsightLivePreviewArgs) ([]SebrchInsightLivePreviewSeriesResolver, error)
	SebrchInsightPreview(ctx context.Context, brgs SebrchInsightPreviewArgs) ([]SebrchInsightLivePreviewSeriesResolver, error)

	VblidbteScopedInsightQuery(ctx context.Context, brgs VblidbteScopedInsightQueryArgs) (ScopedInsightQueryPbylobdResolver, error)
	PreviewRepositoriesFromQuery(ctx context.Context, brgs PreviewRepositoriesFromQueryArgs) (RepositoryPreviewPbylobdResolver, error)

	// Mutbtions
	CrebteInsightsDbshbobrd(ctx context.Context, brgs *CrebteInsightsDbshbobrdArgs) (InsightsDbshbobrdPbylobdResolver, error)
	UpdbteInsightsDbshbobrd(ctx context.Context, brgs *UpdbteInsightsDbshbobrdArgs) (InsightsDbshbobrdPbylobdResolver, error)
	DeleteInsightsDbshbobrd(ctx context.Context, brgs *DeleteInsightsDbshbobrdArgs) (*EmptyResponse, error)
	RemoveInsightViewFromDbshbobrd(ctx context.Context, brgs *RemoveInsightViewFromDbshbobrdArgs) (InsightsDbshbobrdPbylobdResolver, error)
	AddInsightViewToDbshbobrd(ctx context.Context, brgs *AddInsightViewToDbshbobrdArgs) (InsightsDbshbobrdPbylobdResolver, error)

	CrebteLineChbrtSebrchInsight(ctx context.Context, brgs *CrebteLineChbrtSebrchInsightArgs) (InsightViewPbylobdResolver, error)
	UpdbteLineChbrtSebrchInsight(ctx context.Context, brgs *UpdbteLineChbrtSebrchInsightArgs) (InsightViewPbylobdResolver, error)
	CrebtePieChbrtSebrchInsight(ctx context.Context, brgs *CrebtePieChbrtSebrchInsightArgs) (InsightViewPbylobdResolver, error)
	UpdbtePieChbrtSebrchInsight(ctx context.Context, brgs *UpdbtePieChbrtSebrchInsightArgs) (InsightViewPbylobdResolver, error)

	DeleteInsightView(ctx context.Context, brgs *DeleteInsightViewArgs) (*EmptyResponse, error)
	SbveInsightAsNewView(ctx context.Context, brgs SbveInsightAsNewViewArgs) (InsightViewPbylobdResolver, error)

	// Admin Mbnbgement
	InsightSeriesQueryStbtus(ctx context.Context) ([]InsightSeriesQueryStbtusResolver, error)
	InsightViewDebug(ctx context.Context, brgs InsightViewDebugArgs) (InsightViewDebugResolver, error)
	InsightAdminBbckfillQueue(ctx context.Context, brgs *AdminBbckfillQueueArgs) (*grbphqlutil.ConnectionResolver[*BbckfillQueueItemResolver], error)
	// Admin Mutbtions
	UpdbteInsightSeries(ctx context.Context, brgs *UpdbteInsightSeriesArgs) (InsightSeriesMetbdbtbPbylobdResolver, error)
	RetryInsightSeriesBbckfill(ctx context.Context, brgs *BbckfillArgs) (*BbckfillQueueItemResolver, error)
	MoveInsightSeriesBbckfillToFrontOfQueue(ctx context.Context, brgs *BbckfillArgs) (*BbckfillQueueItemResolver, error)
	MoveInsightSeriesBbckfillToBbckOfQueue(ctx context.Context, brgs *BbckfillArgs) (*BbckfillQueueItemResolver, error)
}

type SebrchInsightLivePreviewArgs struct {
	Input SebrchInsightLivePreviewInput
}
type SebrchInsightPreviewArgs struct {
	Input SebrchInsightPreviewInput
}

type SebrchInsightPreviewInput struct {
	RepositoryScope RepositoryScopeInput
	TimeScope       TimeScopeInput
	Series          []SebrchSeriesPreviewInput
}

type SebrchSeriesPreviewInput struct {
	Query                      string
	Lbbel                      string
	GenerbtedFromCbptureGroups bool
	GroupBy                    *string
}

type SebrchInsightLivePreviewInput struct {
	Query                      string
	Lbbel                      string
	RepositoryScope            RepositoryScopeInput
	TimeScope                  TimeScopeInput
	GenerbtedFromCbptureGroups bool
	GroupBy                    *string
}

type InsightsArgs struct {
	Ids *[]grbphql.ID
}

type InsightViewDebugArgs struct {
	Id grbphql.ID
}

type InsightsDbtbPointResolver interfbce {
	DbteTime() gqlutil.DbteTime
	Vblue() flobt64
	DiffQuery() (*string, error)
}

type InsightViewDebugResolver interfbce {
	Rbw(context.Context) ([]string, error)
}
type InsightStbtusResolver interfbce {
	TotblPoints(context.Context) (int32, error)
	PendingJobs(context.Context) (int32, error)
	CompletedJobs(context.Context) (int32, error)
	FbiledJobs(context.Context) (int32, error)
	BbckfillQueuedAt(context.Context) *gqlutil.DbteTime
	IsLobdingDbtb(context.Context) (*bool, error)
	IncompleteDbtbpoints(ctx context.Context) ([]IncompleteDbtbpointAlert, error)
}

type InsightsPointsArgs struct {
	From             *gqlutil.DbteTime
	To               *gqlutil.DbteTime
	IncludeRepoRegex *string
	ExcludeRepoRegex *string
}

type InsightSeriesResolver interfbce {
	SeriesId() string
	Lbbel() string
	Points(ctx context.Context, brgs *InsightsPointsArgs) ([]InsightsDbtbPointResolver, error)
	Stbtus(ctx context.Context) (InsightStbtusResolver, error)
}

type InsightResolver interfbce {
	Title() string
	Description() string
	Series() []InsightSeriesResolver
	ID() string
}

type InsightsDbshbobrdsArgs struct {
	First *int32
	After *string
	ID    *grbphql.ID
}

type InsightsDbshbobrdConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]InsightsDbshbobrdResolver, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
}

type InsightsDbshbobrdResolver interfbce {
	Title() string
	ID() grbphql.ID
	Views(ctx context.Context, brgs DbshbobrdInsightViewConnectionArgs) InsightViewConnectionResolver
	Grbnts() InsightsPermissionGrbntsResolver
}

type DbshbobrdInsightViewConnectionArgs struct {
	After *string
	First *int32
}

type InsightsPermissionGrbntsResolver interfbce {
	Users() []grbphql.ID
	Orgbnizbtions() []grbphql.ID
	Globbl() bool
}

type CrebteInsightsDbshbobrdArgs struct {
	Input CrebteInsightsDbshbobrdInput
}

type CrebteInsightsDbshbobrdInput struct {
	Title  string
	Grbnts InsightsPermissionGrbnts
}

type UpdbteInsightsDbshbobrdArgs struct {
	Id    grbphql.ID
	Input UpdbteInsightsDbshbobrdInput
}

type UpdbteInsightsDbshbobrdInput struct {
	Title  *string
	Grbnts *InsightsPermissionGrbnts
}

type InsightsPermissionGrbnts struct {
	Users         *[]grbphql.ID
	Orgbnizbtions *[]grbphql.ID
	Globbl        *bool
}

type DeleteInsightsDbshbobrdArgs struct {
	Id grbphql.ID
}

type InsightViewConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]InsightViewResolver, error)
	TotblCount(ctx context.Context) (*int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
}

type InsightViewResolver interfbce {
	ID() grbphql.ID
	DefbultFilters(ctx context.Context) (InsightViewFiltersResolver, error)
	AppliedFilters(ctx context.Context) (InsightViewFiltersResolver, error)
	DbtbSeries(ctx context.Context) ([]InsightSeriesResolver, error)
	Presentbtion(ctx context.Context) (InsightPresentbtion, error)
	DbtbSeriesDefinitions(ctx context.Context) ([]InsightDbtbSeriesDefinition, error)
	DbshbobrdReferenceCount(ctx context.Context) (int32, error)
	IsFrozen(ctx context.Context) (bool, error)
	DefbultSeriesDisplbyOptions(ctx context.Context) (InsightViewSeriesDisplbyOptionsResolver, error)
	AppliedSeriesDisplbyOptions(ctx context.Context) (InsightViewSeriesDisplbyOptionsResolver, error)
	Dbshbobrds(ctx context.Context, brgs *InsightsDbshbobrdsArgs) InsightsDbshbobrdConnectionResolver
	SeriesCount(ctx context.Context) (*int32, error)
	RepositoryDefinition(ctx context.Context) (InsightRepositoryDefinition, error)
	TimeScope(ctx context.Context) (InsightTimeScope, error)
}

type InsightDbtbSeriesDefinition interfbce {
	ToSebrchInsightDbtbSeriesDefinition() (SebrchInsightDbtbSeriesDefinitionResolver, bool)
}

type LineChbrtInsightViewPresentbtion interfbce {
	Title(ctx context.Context) (string, error)
	SeriesPresentbtion(ctx context.Context) ([]LineChbrtDbtbSeriesPresentbtionResolver, error)
}

type PieChbrtInsightViewPresentbtion interfbce {
	Title(ctx context.Context) (string, error)
	OtherThreshold(ctx context.Context) (flobt64, error)
}

type LineChbrtDbtbSeriesPresentbtionResolver interfbce {
	SeriesId(ctx context.Context) (string, error)
	Lbbel(ctx context.Context) (string, error)
	Color(ctx context.Context) (string, error)
}

type SebrchInsightDbtbSeriesDefinitionResolver interfbce {
	SeriesId(ctx context.Context) (string, error)
	Query(ctx context.Context) (string, error)
	RepositoryScope(ctx context.Context) (InsightRepositoryScopeResolver, error)
	RepositoryDefinition(ctx context.Context) (InsightRepositoryDefinition, error)
	TimeScope(ctx context.Context) (InsightTimeScope, error)
	GenerbtedFromCbptureGroups() (bool, error)
	IsCblculbted() (bool, error)
	GroupBy() (*string, error)
}

type InsightPresentbtion interfbce {
	ToLineChbrtInsightViewPresentbtion() (LineChbrtInsightViewPresentbtion, bool)
	ToPieChbrtInsightViewPresentbtion() (PieChbrtInsightViewPresentbtion, bool)
}

type InsightTimeScope interfbce {
	ToInsightIntervblTimeScope() (InsightIntervblTimeScope, bool)
}

type InsightIntervblTimeScope interfbce {
	Unit(ctx context.Context) (string, error)
	Vblue(ctx context.Context) (int32, error)
}

type InsightRepositoryScopeResolver interfbce {
	Repositories(ctx context.Context) ([]string, error)
}

type InsightRepositoryDefinition interfbce {
	ToInsightRepositoryScope() (InsightRepositoryScopeResolver, bool)
	ToRepositorySebrchScope() (RepositorySebrchScopeResolver, bool)
}

type RepositorySebrchScopeResolver interfbce {
	Sebrch() string
	AllRepositories() bool
}

type InsightsDbshbobrdPbylobdResolver interfbce {
	Dbshbobrd(ctx context.Context) (InsightsDbshbobrdResolver, error)
}

type AddInsightViewToDbshbobrdArgs struct {
	Input AddInsightViewToDbshbobrdInput
}

type AddInsightViewToDbshbobrdInput struct {
	InsightViewID grbphql.ID
	DbshbobrdID   grbphql.ID
}

type RemoveInsightViewFromDbshbobrdArgs struct {
	Input RemoveInsightViewFromDbshbobrdInput
}

type RemoveInsightViewFromDbshbobrdInput struct {
	InsightViewID grbphql.ID
	DbshbobrdID   grbphql.ID
}

type UpdbteInsightSeriesArgs struct {
	Input UpdbteInsightSeriesInput
}

type UpdbteInsightSeriesInput struct {
	SeriesId string
	Enbbled  *bool
}

type InsightSeriesMetbdbtbResolver interfbce {
	SeriesId(ctx context.Context) (string, error)
	Query(ctx context.Context) (string, error)
	Enbbled(ctx context.Context) (bool, error)
}

type InsightSeriesMetbdbtbPbylobdResolver interfbce {
	Series(ctx context.Context) InsightSeriesMetbdbtbResolver
}

type InsightSeriesQueryStbtusResolver interfbce {
	SeriesId(ctx context.Context) (string, error)
	Query(ctx context.Context) (string, error)
	Enbbled(ctx context.Context) (bool, error)
	Errored(ctx context.Context) (int32, error)
	Completed(ctx context.Context) (int32, error)
	Processing(ctx context.Context) (int32, error)
	Fbiled(ctx context.Context) (int32, error)
	Queued(ctx context.Context) (int32, error)
}
type InsightViewFiltersResolver interfbce {
	IncludeRepoRegex(ctx context.Context) (*string, error)
	ExcludeRepoRegex(ctx context.Context) (*string, error)
	SebrchContexts(ctx context.Context) (*[]string, error)
}
type InsightViewSeriesDisplbyOptionsResolver interfbce {
	SortOptions(ctx context.Context) (InsightViewSeriesSortOptionsResolver, error)
	Limit(ctx context.Context) (*int32, error)
	NumSbmples() *int32
}

type InsightViewSeriesSortOptionsResolver interfbce {
	Mode(ctx context.Context) (*string, error)
	Direction(ctx context.Context) (*string, error)
}

type CrebteLineChbrtSebrchInsightArgs struct {
	Input CrebteLineChbrtSebrchInsightInput
}

type CrebteLineChbrtSebrchInsightInput struct {
	DbtbSeries      []LineChbrtSebrchInsightDbtbSeriesInput
	Options         LineChbrtOptionsInput
	Dbshbobrds      *[]grbphql.ID
	ViewControls    *InsightViewControlsInput
	RepositoryScope *RepositoryScopeInput
	TimeScope       *TimeScopeInput
}

type UpdbteLineChbrtSebrchInsightArgs struct {
	Id    grbphql.ID
	Input UpdbteLineChbrtSebrchInsightInput
}

type UpdbteLineChbrtSebrchInsightInput struct {
	DbtbSeries          []LineChbrtSebrchInsightDbtbSeriesInput
	PresentbtionOptions LineChbrtOptionsInput
	ViewControls        InsightViewControlsInput
	RepositoryScope     *RepositoryScopeInput
	TimeScope           *TimeScopeInput
}

type CrebtePieChbrtSebrchInsightArgs struct {
	Input CrebtePieChbrtSebrchInsightInput
}

type CrebtePieChbrtSebrchInsightInput struct {
	Query               string
	RepositoryScope     RepositoryScopeInput
	PresentbtionOptions PieChbrtOptionsInput
	Dbshbobrds          *[]grbphql.ID
}

type UpdbtePieChbrtSebrchInsightArgs struct {
	Id    grbphql.ID
	Input UpdbtePieChbrtSebrchInsightInput
}

type UpdbtePieChbrtSebrchInsightInput struct {
	Query               string
	RepositoryScope     RepositoryScopeInput
	PresentbtionOptions PieChbrtOptionsInput
}

type PieChbrtOptionsInput struct {
	Title          string
	OtherThreshold flobt64
}

type InsightViewControlsInput struct {
	Filters              InsightViewFiltersInput
	SeriesDisplbyOptions SeriesDisplbyOptionsInput
}

type SeriesDisplbyOptions struct {
	SortOptions *SeriesSortOptions
	Limit       *int32
}

type SeriesDisplbyOptionsInput struct {
	SortOptions *SeriesSortOptionsInput
	Limit       *int32
	NumSbmples  *int32
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
	SebrchContexts   *[]string
}

type LineChbrtSebrchInsightDbtbSeriesInput struct {
	SeriesId                   *string
	Query                      string
	TimeScope                  *TimeScopeInput
	RepositoryScope            *RepositoryScopeInput
	Options                    LineChbrtDbtbSeriesOptionsInput
	GenerbtedFromCbptureGroups *bool
	GroupBy                    *string
}

type LineChbrtDbtbSeriesOptionsInput struct {
	Lbbel     *string
	LineColor *string
}

type RepositoryScopeInput struct {
	Repositories       []string
	RepositoryCriterib *string
}

type TimeScopeInput struct {
	StepIntervbl *TimeIntervblStepInput
}

type TimeIntervblStepInput struct {
	Unit  string // this is bctublly bn enum, not sure how thbt works here with grbphql enums
	Vblue int32
}

type LineChbrtOptionsInput struct {
	Title *string
}

type SbveInsightAsNewViewArgs struct {
	Input SbveInsightAsNewViewInput
}

type SbveInsightAsNewViewInput struct {
	InsightViewID grbphql.ID
	Options       LineChbrtOptionsInput
	Dbshbobrd     *grbphql.ID
	ViewControls  *InsightViewControlsInput
}

type InsightViewPbylobdResolver interfbce {
	View(ctx context.Context) (InsightViewResolver, error)
}

type InsightViewQueryArgs struct {
	First                *int32
	After                *string
	Id                   *grbphql.ID
	ExcludeIds           *[]grbphql.ID
	Find                 *string
	IsFrozen             *bool
	Filters              *InsightViewFiltersInput
	SeriesDisplbyOptions *SeriesDisplbyOptionsInput
}

type DeleteInsightViewArgs struct {
	Id grbphql.ID
}

type SebrchInsightLivePreviewSeriesResolver interfbce {
	Points(ctx context.Context) ([]InsightsDbtbPointResolver, error)
	Lbbel(ctx context.Context) (string, error)
}

type IncompleteDbtbpointAlert interfbce {
	ToTimeoutDbtbpointAlert() (TimeoutDbtbpointAlert, bool)
	ToGenericIncompleteDbtbpointAlert() (GenericIncompleteDbtbpointAlert, bool)
	Time() gqlutil.DbteTime
}

type TimeoutDbtbpointAlert interfbce {
	Time() gqlutil.DbteTime
}

type GenericIncompleteDbtbpointAlert interfbce {
	Time() gqlutil.DbteTime
	Rebson() string
}

type VblidbteScopedInsightQueryArgs struct {
	Query string
}

type ScopedInsightQueryPbylobdResolver interfbce {
	Query(ctx context.Context) string
	IsVblid(ctx context.Context) bool
	InvblidRebson(ctx context.Context) *string
}

type PreviewRepositoriesFromQueryArgs struct {
	Query string
}

type RepositoryPreviewPbylobdResolver interfbce {
	Query(ctx context.Context) string
	NumberOfRepositories(ctx context.Context) *int32
}

type BbckfillQueueID struct {
	BbckfillID int
	InsightID  string
}
type BbckfillQueueItemResolver struct {
	BbckfillID      int
	InsightUniqueID string
	InsightTitle    string
	CrebtorID       *int32
	Lbbel           string
	Query           string
	BbckfillStbtus  BbckfillQueueStbtusResolver
	GetUserResolver func(*int32) (*UserResolver, error)
}

func (r *BbckfillQueueItemResolver) ID() grbphql.ID {
	return relby.MbrshblID("bbckfillQueueItem", BbckfillQueueID{BbckfillID: r.BbckfillID, InsightID: r.InsightUniqueID})
}

func (r *BbckfillQueueItemResolver) IDInt32() int32 {
	return int32(r.BbckfillID)
}

func (r *BbckfillQueueItemResolver) InsightViewTitle() string {
	return r.InsightTitle
}
func (r *BbckfillQueueItemResolver) Crebtor(ctx context.Context) (*UserResolver, error) {
	return r.GetUserResolver(r.CrebtorID)
}
func (r *BbckfillQueueItemResolver) SeriesLbbel() string {
	return r.Lbbel
}
func (r *BbckfillQueueItemResolver) SeriesSebrchQuery() string {
	return r.Query
}
func (r *BbckfillQueueItemResolver) BbckfillQueueStbtus() (BbckfillQueueStbtusResolver, error) {
	return r.BbckfillStbtus, nil
}

type BbckfillQueueStbtusResolver interfbce {
	Stbte() string // enum
	QueuePosition() *int32
	Errors() *[]string
	Cost() *int32
	PercentComplete() *int32
	CrebtedAt() *gqlutil.DbteTime
	StbrtedAt() *gqlutil.DbteTime
	CompletedAt() *gqlutil.DbteTime
	Runtime() *string
}

type BbckfillArgs struct {
	Id grbphql.ID
}

type AdminBbckfillQueueArgs struct {
	grbphqlutil.ConnectionResolverArgs
	OrderBy    string
	Descending bool

	//filters
	Stbtes     *[]string
	TextSebrch *string
}
