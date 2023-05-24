package types

import (
	"time"
)

// InsightViewSeries is an abstraction of a complete Code Insight. This type materializes a view with any associated series.
type InsightViewSeries struct {
	ViewID                        int
	DashboardViewID               int
	InsightSeriesID               int
	UniqueID                      string
	SeriesID                      string
	Title                         string
	Description                   string
	Query                         string
	CreatedAt                     time.Time
	OldestHistoricalAt            time.Time
	LastRecordedAt                time.Time
	NextRecordingAfter            time.Time
	LastSnapshotAt                time.Time
	NextSnapshotAfter             time.Time
	BackfillQueuedAt              *time.Time
	Label                         string
	LineColor                     string
	Repositories                  []string
	SampleIntervalUnit            string
	SampleIntervalValue           int
	DefaultFilterIncludeRepoRegex *string
	DefaultFilterExcludeRepoRegex *string
	DefaultFilterSearchContexts   []string
	OtherThreshold                *float64
	PresentationType              PresentationType
	GeneratedFromCaptureGroups    bool
	JustInTime                    bool
	GenerationMethod              GenerationMethod
	IsFrozen                      bool
	SeriesSortMode                *SeriesSortMode
	SeriesSortDirection           *SeriesSortDirection
	SeriesLimit                   *int32
	GroupBy                       *string
	BackfillAttempts              int32
	SupportsAugmentation          bool
	RepositoryCriteria            *string
	SeriesNumSamples              *int32
}

type Insight struct {
	ViewID           int
	DashboardViewId  int
	UniqueID         string
	Title            string
	Description      string
	Series           []InsightViewSeries
	Filters          InsightViewFilters
	OtherThreshold   *float64
	PresentationType PresentationType
	IsFrozen         bool
	SeriesOptions    SeriesDisplayOptions
}

type InsightViewFilters struct {
	IncludeRepoRegex *string
	ExcludeRepoRegex *string
	SearchContexts   []string
}

// InsightViewSeriesMetadata contains metadata about a viewable insight series such as render properties.
type InsightViewSeriesMetadata struct {
	Label  string
	Stroke string
}

// InsightView is a single insight view that may or may not have any associated series.
type InsightView struct {
	ID                  int
	Title               string
	Description         string
	UniqueID            string
	Filters             InsightViewFilters
	OtherThreshold      *float64
	PresentationType    PresentationType
	IsFrozen            bool
	SeriesSortMode      *SeriesSortMode
	SeriesSortDirection *SeriesSortDirection
	SeriesLimit         *int32
	SeriesNumSamples    *int32
}

// InsightSeries is a single data series for a Code Insight. This contains some metadata about the data series, as well
// as its unique series ID.
type InsightSeries struct {
	ID                         int
	SeriesID                   string
	Query                      string
	CreatedAt                  time.Time
	OldestHistoricalAt         time.Time
	LastRecordedAt             time.Time
	NextRecordingAfter         time.Time
	LastSnapshotAt             time.Time
	NextSnapshotAfter          time.Time
	BackfillQueuedAt           time.Time
	Enabled                    bool
	Repositories               []string
	SampleIntervalUnit         string
	SampleIntervalValue        int
	GeneratedFromCaptureGroups bool
	JustInTime                 bool
	GenerationMethod           GenerationMethod
	GroupBy                    *string
	BackfillAttempts           int32
	SupportsAugmentation       bool
	RepositoryCriteria         *string
}

type IntervalUnit string

const (
	Month IntervalUnit = "MONTH"
	Day   IntervalUnit = "DAY"
	Week  IntervalUnit = "WEEK"
	Year  IntervalUnit = "YEAR"
	Hour  IntervalUnit = "HOUR"
)

// GenerationMethod represents the method of execution for which to populate time series data for an insight series. This is effectively an enum of values.
type GenerationMethod string

const (
	Search         GenerationMethod = "search"
	SearchCompute  GenerationMethod = "search-compute"
	LanguageStats  GenerationMethod = "language-stats"
	MappingCompute GenerationMethod = "mapping-compute"
)

type Dashboard struct {
	ID           int
	Title        string
	InsightIDs   []string // shallow references
	UserIdGrants []int64
	OrgIdGrants  []int64
	GlobalGrant  bool
	Save         bool // temporarily save dashboards from being cleared during setting migration
}

type InsightSeriesStatus struct {
	SeriesId   string
	Query      string
	Enabled    bool
	Errored    int
	Processing int
	Queued     int
	Failed     int
	Completed  int
}

type InsightSearchFailure struct {
	Query          string
	QueuedAt       time.Time
	State          string
	FailureMessage string
	RecordTime     *time.Time
	PersistMode    string
}

type PresentationType string

const (
	Line PresentationType = "LINE"
	Pie  PresentationType = "PIE"
)

type SeriesSortMode string

const (
	ResultCount     SeriesSortMode = "RESULT_COUNT"    // Sorts by the number of results for the most recent datapoint of a series.
	DateAdded       SeriesSortMode = "DATE_ADDED"      // Sorts by the date of the earliest datapoint in the series.
	Lexicographical SeriesSortMode = "LEXICOGRAPHICAL" // Sorts by label: first by semantic version and then alphabetically.
)

type SeriesSortDirection string

const (
	Asc  SeriesSortDirection = "ASC"
	Desc SeriesSortDirection = "DESC"
)

type SeriesDisplayOptions struct {
	SortOptions *SeriesSortOptions
	Limit       *int32
	NumSamples  *int32
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
	Timestamp time.Time
	Snapshot  bool
}

type SearchAggregationMode string

const (
	REPO_AGGREGATION_MODE          SearchAggregationMode = "REPO"
	PATH_AGGREGATION_MODE          SearchAggregationMode = "PATH"
	AUTHOR_AGGREGATION_MODE        SearchAggregationMode = "AUTHOR"
	CAPTURE_GROUP_AGGREGATION_MODE SearchAggregationMode = "CAPTURE_GROUP"
	REPO_METADATA_AGGREGATION_MODE SearchAggregationMode = "REPO_METADATA"
)

var SearchAggregationModes = []SearchAggregationMode{REPO_AGGREGATION_MODE, PATH_AGGREGATION_MODE, AUTHOR_AGGREGATION_MODE, CAPTURE_GROUP_AGGREGATION_MODE, REPO_METADATA_AGGREGATION_MODE}

type AggregationNotAvailableReasonType string

const (
	INVALID_QUERY                      AggregationNotAvailableReasonType = "INVALID_QUERY"
	INVALID_AGGREGATION_MODE_FOR_QUERY AggregationNotAvailableReasonType = "INVALID_AGGREGATION_MODE_FOR_QUERY"
	TIMEOUT_EXTENSION_AVAILABLE        AggregationNotAvailableReasonType = "TIMEOUT_EXTENSION_AVAILABLE"
	TIMEOUT_NO_EXTENSION_AVAILABLE     AggregationNotAvailableReasonType = "TIMEOUT_NO_EXTENSION_AVAILABLE"
	ERROR_OCCURRED                     AggregationNotAvailableReasonType = "ERROR_OCCURRED"
)

const (
	NO_REPO_METADATA_TEXT = "No metadata"
)
