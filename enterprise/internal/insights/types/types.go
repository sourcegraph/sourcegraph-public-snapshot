package types

import (
	"time"
)

// InsightViewSeries is an abstraction of a complete Code Insight. This type materializes a view with any associated series.
type InsightViewSeries struct {
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
}

type Insight struct {
	UniqueID    string
	Title       string
	Description string
	Series      []InsightViewSeries
	Filters     InsightViewFilters
}

type InsightViewFilters struct {
	IncludeRepoRegex *string
	ExcludeRepoRegex *string
}

// InsightViewSeriesMetadata contains metadata about a viewable insight series such as render properties.
type InsightViewSeriesMetadata struct {
	Label  string
	Stroke string
}

// InsightView is a single insight view that may or may not have any associated series.
type InsightView struct {
	ID          int
	Title       string
	Description string
	UniqueID    string
}

// InsightSeries is a single data series for a Code Insight. This contains some metadata about the data series, as well
// as its unique series ID.
type InsightSeries struct {
	ID                 int
	SeriesID           string
	Query              string
	CreatedAt          time.Time
	OldestHistoricalAt time.Time
	LastRecordedAt     time.Time
	NextRecordingAfter time.Time
	LastSnapshotAt     time.Time
	NextSnapshotAfter  time.Time
	BackfillQueuedAt   time.Time
	Enabled            bool
}

type DirtyQuery struct {
	ID      int
	Query   string
	ForTime time.Time
	DirtyAt time.Time
	Reason  string
}

type DirtyQueryAggregate struct {
	Count   int
	ForTime time.Time
	Reason  string
}

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
