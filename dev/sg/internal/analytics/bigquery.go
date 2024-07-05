package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"

	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (e event) Save() (map[string]bigquery.Value, string, error) {
	m := map[string]bigquery.Value{
		"uuid":           e.UUID,
		"user_id":        e.UserID,
		"recorded_at":    e.RecordedAt,
		"command":        e.Command,
		"version":        e.Version,
		"flags_and_args": string(e.FlagsAndArgs),
		"duration":       e.Duration,
		"error":          e.Error,
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
	//e.FlagsAndArgs = i.GetFlagsAndArgs()
	e.Duration = i.GetDuration()
	e.Error = i.GetError()

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

func (bq *BigQueryClient) transact(ctx context.Context) (session string, commit, rollback func() error, _ error) {
	q := bq.Query("SELECT 1;")
	q.CreateSession = true
	q.Location = "US"
	job, err := q.Run(ctx)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery session: starting query: %w", err)
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery session: awaiting query: %w", err)
	}
	if err := status.Err(); err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery session: running query: %w", err)
	}

	session = job.LastStatus().Statistics.SessionInfo.SessionID

	q = bq.Query("BEGIN TRANSACTION")
	q.ConnectionProperties = append(q.ConnectionProperties, &bigquery.ConnectionProperty{
		Key:   "session_id",
		Value: session,
	})
	q.Location = "US"
	job, err = q.Run(ctx)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery transaction: starting query: %w", err)
	}
	status, err = job.Wait(ctx)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery transaction: awaiting query: %w", err)
	}
	if err := status.Err(); err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery transaction: running query: %w", err)
	}

	quitSession := func() error {
		q = bq.Query("CALL BQ.ABORT_SESSION()")
		q.ConnectionProperties = append(q.ConnectionProperties, &bigquery.ConnectionProperty{
			Key:   "session_id",
			Value: session,
		})
		q.Location = "US"
		job, err = q.Run(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to quit bigquery session: starting query")
		}
		status, err = job.Wait(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to quit bigquery session: awaiting query")
		}
		if err := status.Err(); err != nil {
			return errors.Wrap(err, "failed to quit bigquery session: running query")
		}
		return nil
	}

	return session, func() error {
			defer quitSession()

			q = bq.Query("COMMIT TRANSACTION")
			q.ConnectionProperties = append(q.ConnectionProperties, &bigquery.ConnectionProperty{
				Key:   "session_id",
				Value: session,
			})
			q.Location = "US"
			job, err = q.Run(ctx)
			if err != nil {
				return errors.Wrapf(err, "failed to commit bigquery transaction: starting query")
			}
			status, err = job.Wait(ctx)
			if err != nil {
				return errors.Wrapf(err, "failed to commit bigquery transaction: awaiting query")
			}
			if err := status.Err(); err != nil {
				return errors.Wrapf(err, "failed to commit bigquery transaction: running query")
			}
			return nil
		}, func() error {
			defer quitSession()

			q = bq.Query("ROLLBACK TRANSACTION")
			q.ConnectionProperties = append(q.ConnectionProperties, &bigquery.ConnectionProperty{
				Key:   "session_id",
				Value: session,
			})
			q.Location = "US"
			job, err = q.Run(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to rollback bigquery transaction: starting query: %w")
			}
			status, err = job.Wait(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to rollback bigquery transaction: awaiting query: %w")
			}
			if err := status.Err(); err != nil {
				return errors.Wrap(err, "failed to rollback bigquery transaction: running query: %w")
			}
			return nil
		}, nil
}
