package productsubscription

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var codyGatewaySACredentialFilePath = func() string {
	if v := env.Get("CODY_GATEWAY_BIGQUERY_ACCESS_CREDENTIALS_FILE", "", "BigQuery credentials for the Cody Gateway service"); v != "" {
		return v
	}
	return env.Get("LLM_PROXY_BIGQUERY_ACCESS_CREDENTIALS_FILE", "", "DEPRECATED: Use CODY_GATEWAY_BIGQUERY_ACCESS_CREDENTIALS_FILE instead")
}()

type CodyGatewayService interface {
	UsageForSubscription(ctx context.Context, uuid string) ([]SubscriptionUsage, error)
}

type CodyGatewayServiceOptions struct {
	BigQuery ServiceBigQueryOptions
}

type ServiceBigQueryOptions struct {
	CredentialFilePath string
	ProjectID          string
	Dataset            string
	EventsTable        string
}

func (o ServiceBigQueryOptions) IsConfigured() bool {
	return o.ProjectID != "" && o.Dataset != "" && o.EventsTable != ""
}

func NewCodyGatewayService() *codyGatewayService {
	opts := CodyGatewayServiceOptions{}

	d := conf.Get().Dotcom
	if d != nil && d.CodyGateway != nil {
		opts.BigQuery.CredentialFilePath = codyGatewaySACredentialFilePath
		opts.BigQuery.ProjectID = d.CodyGateway.BigQueryGoogleProjectID
		opts.BigQuery.Dataset = d.CodyGateway.BigQueryDataset
		opts.BigQuery.EventsTable = d.CodyGateway.BigQueryTable
	}

	return NewCodyGatewayServiceWithOptions(opts)
}

func NewCodyGatewayServiceWithOptions(opts CodyGatewayServiceOptions) *codyGatewayService {
	return &codyGatewayService{
		opts: opts,
	}
}

type SubscriptionUsage struct {
	Date  time.Time
	Model string
	Count int64
}

type codyGatewayService struct {
	opts CodyGatewayServiceOptions
}

func (s *codyGatewayService) CompletionsUsageForActor(ctx context.Context, feature types.CompletionsFeature, actorSource codygateway.ActorSource, actorID string) ([]SubscriptionUsage, error) {
	if !s.opts.BigQuery.IsConfigured() {
		// Not configured, nothing we can do.
		return nil, nil
	}

	client, err := bigquery.NewClient(ctx, s.opts.BigQuery.ProjectID, gcpClientOptions(s.opts.BigQuery.CredentialFilePath)...)
	if err != nil {
		return nil, errors.Wrap(err, "creating BigQuery client")
	}
	defer client.Close()

	tbl := client.Dataset(s.opts.BigQuery.Dataset).Table(s.opts.BigQuery.EventsTable)

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
			Value: codygateway.EventNameCompletionsFinished,
		},
		{
			Name:  codygateway.CompletionsEventFeatureMetadataField,
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

func (s *codyGatewayService) EmbeddingsUsageForActor(ctx context.Context, actorSource codygateway.ActorSource, actorID string) ([]SubscriptionUsage, error) {
	if !s.opts.BigQuery.IsConfigured() {
		// Not configured, nothing we can do.
		return nil, nil
	}

	client, err := bigquery.NewClient(ctx, s.opts.BigQuery.ProjectID, gcpClientOptions(s.opts.BigQuery.CredentialFilePath)...)
	if err != nil {
		return nil, errors.Wrap(err, "creating BigQuery client")
	}
	defer client.Close()

	tbl := client.Dataset(s.opts.BigQuery.Dataset).Table(s.opts.BigQuery.EventsTable)

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
			Value: codygateway.EventNameEmbeddingsFinished,
		},
		{
			Name:  codygateway.CompletionsEventFeatureMetadataField,
			Value: codygateway.CompletionsEventFeatureEmbeddings,
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

func gcpClientOptions(credentialFilePath string) []option.ClientOption {
	if credentialFilePath != "" {
		return []option.ClientOption{option.WithCredentialsFile(credentialFilePath)}
	}

	return nil
}
