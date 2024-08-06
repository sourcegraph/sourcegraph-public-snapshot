package analytics

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"
)

type BigQueryClient struct {
	*bigquery.Client
	ProjectID string
	Dataset   *bigquery.Dataset
	Table     *bigquery.Table
}

type event struct {
	UUID         string          `json:"uuid"`
	UserID       string          `json:"user_id"`
	RecordedAt   time.Time       `json:"recorded_at"`
	Command      string          `json:"command"`
	Version      string          `json:"version"`
	FlagsAndArgs json.RawMessage `json:"flags_and_args,omitempty"`
	Duration     time.Duration   `json:"duration,omitempty"`
	Error        string          `json:"error,omitempty"`
	Data         json.RawMessage `json:"data,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

// Save implements the bigquery.ValueSaver interface which allows it
// to be used on Table.Inserter()
func (e event) Save() (map[string]bigquery.Value, string, error) {
	durationInterval := &bigquery.IntervalValue{
		Seconds: int32(e.Duration.Seconds()),
	}
	m := map[string]bigquery.Value{
		"uuid":           e.UUID,
		"user_id":        e.UserID,
		"recorded_at":    e.RecordedAt,
		"command":        e.Command,
		"version":        e.Version,
		"duration":       durationInterval.String(),
		"error":          e.Error,
		"flags_and_args": string(e.FlagsAndArgs),
		"metadata":       string(e.Metadata),
	}

	insertID := e.UUID
	return m, insertID, nil
}

func NewEvent(i invocation) *event {
	var e event
	e.UUID = i.uuid.String()
	e.UserID = i.GetUserID()
	if t := i.GetEndTime(); t != nil {
		e.RecordedAt = *t
	}
	e.Command = i.GetCommand()
	e.Version = i.GetVersion()
	e.Duration = i.GetDuration()
	e.Error = i.GetError()

	flagsAndArgs := struct {
		Flags map[string]any `json:"flags"`
		Args  []any          `json:"args"`
	}{
		Flags: i.GetFlags(),
		Args:  i.GetArgs(),
	}

	e.FlagsAndArgs, _ = json.Marshal(flagsAndArgs)

	metadata := map[string]any{
		"success":   i.IsSuccess(),
		"failed":    i.IsFailed(),
		"cancelled": i.IsCancelled(),
		"panicked":  i.IsPanicked(),
		"os":        i.GetOS(),
	}

	e.Metadata, _ = json.Marshal(metadata)

	return &e
}

const (
	SGLocalDev           = "sourcegraph-local-dev"
	AnalyticsDatasetName = "sg_analytics"
	EventsTableName      = "events"
)

func NewBigQueryClient(ctx context.Context, project, datasetName, tableName string) (*BigQueryClient, error) {
	client, err := bigquery.NewClient(ctx, project)
	if err != nil {
		return nil, err
	}

	dataset := client.Dataset(datasetName)
	return &BigQueryClient{
		Client:    client,
		ProjectID: project,
		Dataset:   dataset,
		Table:     dataset.Table(tableName),
	}, nil
}

func (bq *BigQueryClient) InsertEvent(ctx context.Context, ev event) error {
	ins := bq.Table.Inserter()
	return ins.Put(ctx, ev)
}
