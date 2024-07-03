package codygatewayevents

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayactor"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ServiceOptions struct {
	BigQuery ServiceBigQueryOptions
}

type ServiceBigQueryOptions struct {
	ClientOptions []option.ClientOption
	ProjectID     string
	Dataset       string
	EventsTable   string
}

func (o ServiceBigQueryOptions) IsConfigured() bool {
	return o.ProjectID != "" && o.Dataset != "" && o.EventsTable != ""
}

// NewService returns a service for interacting with Cody Gateway usage events.
func NewService(opts ServiceOptions) *Service {
	return &Service{
		opts: opts,
	}
}

type SubscriptionUsage struct {
	Date  time.Time
	Model string
	Count int64
}

type Service struct {
	opts ServiceOptions
}

func (s *Service) CompletionsUsageForActor(ctx context.Context, feature types.CompletionsFeature, actorSource codygatewayactor.ActorSource, actorID string) (_ []SubscriptionUsage, err error) {
	if !s.opts.BigQuery.IsConfigured() {
		// Not configured, nothing we can do.
		return nil, nil
	}

	var tr trace.Trace
	tr, ctx = trace.New(ctx, "CompletionsUsageForActor",
		attribute.String("feature", string(feature)),
		attribute.String("actorSource", string(actorSource)),
		attribute.String("actorID", actorID))
	defer tr.EndWithErrIfNotContext(&err)

	client, err := bigquery.NewClient(ctx, s.opts.BigQuery.ProjectID, s.opts.BigQuery.ClientOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "creating BigQuery client")
	}
	defer client.Close()

	tbl := client.Dataset(s.opts.BigQuery.Dataset).Table(s.opts.BigQuery.EventsTable)
	tr.AddEvent("bigquery.NewClient",
		attribute.String("projectID", s.opts.BigQuery.ProjectID),
		attribute.String("dataset", s.opts.BigQuery.Dataset),
		attribute.String("table", s.opts.BigQuery.EventsTable))

	// Count events with the name for made requests for each day in the last 7 days.
	query := fmt.Sprintf(`
WITH date_range AS (
	SELECT DATE(date) AS date
	FROM UNNEST(
		GENERATE_TIMESTAMP_ARRAY(
			TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY),
			CURRENT_TIMESTAMP(),
			INTERVAL 1 DAY
		)
	) AS date
),
models AS (
	SELECT
		DISTINCT(STRING(JSON_QUERY(events.metadata, '$.model'))) AS model
	FROM
		%s.%s
	WHERE
		source = @source
		AND identifier = @identifier
		AND name = @eventName
		AND STRING(JSON_QUERY(events.metadata, '$.feature')) = @feature
		AND STRING(JSON_QUERY(events.metadata, '$.model')) IS NOT NULL
),
date_range_with_models AS (
	SELECT date_range.date, models.model
	FROM date_range
	CROSS JOIN models
)
SELECT
	date_range_with_models.date AS date,
	date_range_with_models.model AS model,
	IFNULL(COUNT(events.date), 0) AS count
FROM
	date_range_with_models
LEFT JOIN (
	SELECT
		DATE(created_at) AS date,
		metadata
	FROM
		%s.%s
	WHERE
		source = @source
		AND identifier = @identifier
		AND name = @eventName
		AND STRING(JSON_QUERY(events.metadata, '$.feature')) = @feature
	) events
ON
	date_range_with_models.date = events.date
	AND STRING(JSON_QUERY(events.metadata, '$.model')) = date_range_with_models.model
GROUP BY
	date_range_with_models.date, date_range_with_models.model
ORDER BY
	date_range_with_models.date DESC, date_range_with_models.model ASC`,
		tbl.DatasetID,
		tbl.TableID,
		tbl.DatasetID,
		tbl.TableID,
	)

	q := client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "source",
			Value: actorSource,
		},
		{
			Name:  "identifier",
			Value: actorID,
		},
		{
			Name:  "eventName",
			Value: EventNameCompletionsFinished,
		},
		{
			Name:  CompletionsEventFeatureMetadataField,
			Value: feature,
		},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "executing query")
	}

	results := make([]SubscriptionUsage, 0)
	for {
		var row struct {
			Date  bigquery.NullDate
			Model string
			Count int64
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, errors.Wrap(err, "reading query result")
		}
		results = append(results, SubscriptionUsage{
			Date:  row.Date.Date.In(time.UTC),
			Model: row.Model,
			Count: row.Count,
		})
	}

	return results, nil
}

func (s *Service) EmbeddingsUsageForActor(ctx context.Context, actorSource codygatewayactor.ActorSource, actorID string) (_ []SubscriptionUsage, err error) {
	if !s.opts.BigQuery.IsConfigured() {
		// Not configured, nothing we can do.
		return nil, nil
	}

	var tr trace.Trace
	tr, ctx = trace.New(ctx, "EmbeddingsUsageForActor",
		attribute.String("actorSource", string(actorSource)),
		attribute.String("actorID", actorID))
	defer tr.EndWithErrIfNotContext(&err)

	client, err := bigquery.NewClient(ctx, s.opts.BigQuery.ProjectID, s.opts.BigQuery.ClientOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "creating BigQuery client")
	}
	defer client.Close()

	tbl := client.Dataset(s.opts.BigQuery.Dataset).Table(s.opts.BigQuery.EventsTable)
	tr.AddEvent("bigquery.NewClient",
		attribute.String("projectID", s.opts.BigQuery.ProjectID),
		attribute.String("dataset", s.opts.BigQuery.Dataset),
		attribute.String("table", s.opts.BigQuery.EventsTable))

	// Count amount of tokens across all requests for made requests for each day abd model
	// in the last 7 days.
	query := fmt.Sprintf(`
WITH date_range AS (
	SELECT DATE(date) AS date
	FROM UNNEST(
		GENERATE_TIMESTAMP_ARRAY(
			TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY),
			CURRENT_TIMESTAMP(),
			INTERVAL 1 DAY
		)
	) AS date
),
models AS (
	SELECT
		DISTINCT(STRING(JSON_QUERY(events.metadata, '$.model'))) AS model
	FROM
		%s.%s
	WHERE
		source = @source
		AND identifier = @identifier
		AND name = @eventName
		AND STRING(JSON_QUERY(events.metadata, '$.feature')) = @feature
		AND STRING(JSON_QUERY(events.metadata, '$.model')) IS NOT NULL
),
date_range_with_models AS (
	SELECT date_range.date, models.model
	FROM date_range
	CROSS JOIN models
)
SELECT
	date_range_with_models.date AS date,
	date_range_with_models.model AS model,
	IFNULL(SUM(INT64(JSON_QUERY(events.metadata, '$.tokens_used'))), 0) AS count
FROM
	date_range_with_models
LEFT JOIN (
	SELECT
		DATE(created_at) AS date,
		metadata
	FROM
		%s.%s
	WHERE
		source = @source
		AND identifier = @identifier
		AND name = @eventName
		AND STRING(JSON_QUERY(events.metadata, '$.feature')) = @feature
	) events
ON
	date_range_with_models.date = events.date
	AND STRING(JSON_QUERY(events.metadata, '$.model')) = date_range_with_models.model
GROUP BY
	date_range_with_models.date, date_range_with_models.model
ORDER BY
	date_range_with_models.date DESC, date_range_with_models.model ASC`,
		tbl.DatasetID,
		tbl.TableID,
		tbl.DatasetID,
		tbl.TableID,
	)

	q := client.Query(query)
	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "source",
			Value: actorSource,
		},
		{
			Name:  "identifier",
			Value: actorID,
		},
		{
			Name:  "eventName",
			Value: EventNameEmbeddingsFinished,
		},
		{
			Name:  CompletionsEventFeatureMetadataField,
			Value: CompletionsEventFeatureEmbeddings,
		},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "executing query")
	}

	results := make([]SubscriptionUsage, 0)
	for {
		var row struct {
			Date  bigquery.NullDate
			Model string
			Count int64
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, errors.Wrap(err, "reading query result")
		}
		results = append(results, SubscriptionUsage{
			Date:  row.Date.Date.In(time.UTC),
			Model: row.Model,
			Count: row.Count,
		})
	}

	return results, nil
}
