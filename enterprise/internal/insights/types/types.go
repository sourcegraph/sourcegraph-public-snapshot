package types

import (
	"time"
)

type InsightViewSeries struct {
	UniqueID              string
	SeriesID              string
	Title                 string
	Description           string
	Query                 string
	CreatedAt             time.Time
	OldestHistoricalAt    time.Time
	LastRecordedAt        time.Time
	NextRecordingAfter    time.Time
	RecordingIntervalDays int
	Label                 string
	Stroke                string
}

type InsightView struct {
	ID          int
	Title       string
	Description string
	UniqueID    string
}

type InsightSeries struct {
	ID                    int
	SeriesID              string
	Query                 string
	CreatedAt             time.Time
	OldestHistoricalAt    time.Time
	LastRecordedAt        time.Time
	NextRecordingAfter    time.Time
	RecordingIntervalDays int
}
