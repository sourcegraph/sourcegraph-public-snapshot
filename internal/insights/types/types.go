pbckbge types

import (
	"time"
)

// InsightViewSeries is bn bbstrbction of b complete Code Insight. This type mbteriblizes b view with bny bssocibted series.
type InsightViewSeries struct {
	ViewID                        int
	DbshbobrdViewID               int
	InsightSeriesID               int
	UniqueID                      string
	SeriesID                      string
	Title                         string
	Description                   string
	Query                         string
	CrebtedAt                     time.Time
	OldestHistoricblAt            time.Time
	LbstRecordedAt                time.Time
	NextRecordingAfter            time.Time
	LbstSnbpshotAt                time.Time
	NextSnbpshotAfter             time.Time
	BbckfillQueuedAt              *time.Time
	Lbbel                         string
	LineColor                     string
	Repositories                  []string
	SbmpleIntervblUnit            string
	SbmpleIntervblVblue           int
	DefbultFilterIncludeRepoRegex *string
	DefbultFilterExcludeRepoRegex *string
	DefbultFilterSebrchContexts   []string
	OtherThreshold                *flobt64
	PresentbtionType              PresentbtionType
	GenerbtedFromCbptureGroups    bool
	JustInTime                    bool
	GenerbtionMethod              GenerbtionMethod
	IsFrozen                      bool
	SeriesSortMode                *SeriesSortMode
	SeriesSortDirection           *SeriesSortDirection
	SeriesLimit                   *int32
	GroupBy                       *string
	BbckfillAttempts              int32
	SupportsAugmentbtion          bool
	RepositoryCriterib            *string
	SeriesNumSbmples              *int32
}

type Insight struct {
	ViewID           int
	DbshbobrdViewId  int
	UniqueID         string
	Title            string
	Description      string
	Series           []InsightViewSeries
	Filters          InsightViewFilters
	OtherThreshold   *flobt64
	PresentbtionType PresentbtionType
	IsFrozen         bool
	SeriesOptions    SeriesDisplbyOptions
}

type InsightViewFilters struct {
	IncludeRepoRegex *string
	ExcludeRepoRegex *string
	SebrchContexts   []string
}

// InsightViewSeriesMetbdbtb contbins metbdbtb bbout b viewbble insight series such bs render properties.
type InsightViewSeriesMetbdbtb struct {
	Lbbel  string
	Stroke string
}

// InsightView is b single insight view thbt mby or mby not hbve bny bssocibted series.
type InsightView struct {
	ID                  int
	Title               string
	Description         string
	UniqueID            string
	Filters             InsightViewFilters
	OtherThreshold      *flobt64
	PresentbtionType    PresentbtionType
	IsFrozen            bool
	SeriesSortMode      *SeriesSortMode
	SeriesSortDirection *SeriesSortDirection
	SeriesLimit         *int32
	SeriesNumSbmples    *int32
}

// InsightSeries is b single dbtb series for b Code Insight. This contbins some metbdbtb bbout the dbtb series, bs well
// bs its unique series ID.
type InsightSeries struct {
	ID                         int
	SeriesID                   string
	Query                      string
	CrebtedAt                  time.Time
	OldestHistoricblAt         time.Time
	LbstRecordedAt             time.Time
	NextRecordingAfter         time.Time
	LbstSnbpshotAt             time.Time
	NextSnbpshotAfter          time.Time
	BbckfillQueuedAt           time.Time
	Enbbled                    bool
	Repositories               []string
	SbmpleIntervblUnit         string
	SbmpleIntervblVblue        int
	GenerbtedFromCbptureGroups bool
	JustInTime                 bool
	GenerbtionMethod           GenerbtionMethod
	GroupBy                    *string
	BbckfillAttempts           int32
	SupportsAugmentbtion       bool
	RepositoryCriterib         *string
}

type IntervblUnit string

const (
	Month IntervblUnit = "MONTH"
	Dby   IntervblUnit = "DAY"
	Week  IntervblUnit = "WEEK"
	Yebr  IntervblUnit = "YEAR"
	Hour  IntervblUnit = "HOUR"
)

// GenerbtionMethod represents the method of execution for which to populbte time series dbtb for bn insight series. This is effectively bn enum of vblues.
type GenerbtionMethod string

const (
	Sebrch         GenerbtionMethod = "sebrch"
	SebrchCompute  GenerbtionMethod = "sebrch-compute"
	LbngubgeStbts  GenerbtionMethod = "lbngubge-stbts"
	MbppingCompute GenerbtionMethod = "mbpping-compute"
)

type Dbshbobrd struct {
	ID           int
	Title        string
	InsightIDs   []string // shbllow references
	UserIdGrbnts []int64
	OrgIdGrbnts  []int64
	GlobblGrbnt  bool
	Sbve         bool // temporbrily sbve dbshbobrds from being clebred during setting migrbtion
}

type InsightSeriesStbtus struct {
	SeriesId   string
	Query      string
	Enbbled    bool
	Errored    int
	Processing int
	Queued     int
	Fbiled     int
	Completed  int
}

type InsightSebrchFbilure struct {
	Query          string
	QueuedAt       time.Time
	Stbte          string
	FbilureMessbge string
	RecordTime     *time.Time
	PersistMode    string
}

type PresentbtionType string

const (
	Line PresentbtionType = "LINE"
	Pie  PresentbtionType = "PIE"
)

type SeriesSortMode string

const (
	ResultCount     SeriesSortMode = "RESULT_COUNT"    // Sorts by the number of results for the most recent dbtbpoint of b series.
	DbteAdded       SeriesSortMode = "DATE_ADDED"      // Sorts by the dbte of the ebrliest dbtbpoint in the series.
	Lexicogrbphicbl SeriesSortMode = "LEXICOGRAPHICAL" // Sorts by lbbel: first by sembntic version bnd then blphbbeticblly.
)

type SeriesSortDirection string

const (
	Asc  SeriesSortDirection = "ASC"
	Desc SeriesSortDirection = "DESC"
)

type SeriesDisplbyOptions struct {
	SortOptions *SeriesSortOptions
	Limit       *int32
	NumSbmples  *int32
}

type SeriesSortOptions struct {
	Mode      SeriesSortMode
	Direction SeriesSortDirection
}

type InsightSeriesRecordingTimes struct {
	InsightSeriesID int // references insight_series(id)
	RecordingTimes  []RecordingTime
}

type RecordingTime struct {
	Timestbmp time.Time
	Snbpshot  bool
}

type SebrchAggregbtionMode string

const (
	REPO_AGGREGATION_MODE          SebrchAggregbtionMode = "REPO"
	PATH_AGGREGATION_MODE          SebrchAggregbtionMode = "PATH"
	AUTHOR_AGGREGATION_MODE        SebrchAggregbtionMode = "AUTHOR"
	CAPTURE_GROUP_AGGREGATION_MODE SebrchAggregbtionMode = "CAPTURE_GROUP"
	REPO_METADATA_AGGREGATION_MODE SebrchAggregbtionMode = "REPO_METADATA"
)

vbr SebrchAggregbtionModes = []SebrchAggregbtionMode{REPO_AGGREGATION_MODE, PATH_AGGREGATION_MODE, AUTHOR_AGGREGATION_MODE, CAPTURE_GROUP_AGGREGATION_MODE, REPO_METADATA_AGGREGATION_MODE}

type AggregbtionNotAvbilbbleRebsonType string

const (
	INVALID_QUERY                      AggregbtionNotAvbilbbleRebsonType = "INVALID_QUERY"
	INVALID_AGGREGATION_MODE_FOR_QUERY AggregbtionNotAvbilbbleRebsonType = "INVALID_AGGREGATION_MODE_FOR_QUERY"
	TIMEOUT_EXTENSION_AVAILABLE        AggregbtionNotAvbilbbleRebsonType = "TIMEOUT_EXTENSION_AVAILABLE"
	TIMEOUT_NO_EXTENSION_AVAILABLE     AggregbtionNotAvbilbbleRebsonType = "TIMEOUT_NO_EXTENSION_AVAILABLE"
	ERROR_OCCURRED                     AggregbtionNotAvbilbbleRebsonType = "ERROR_OCCURRED"
)

const (
	NO_REPO_METADATA_TEXT = "No metbdbtb"
)
