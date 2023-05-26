package productsubscription

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/completions/types"
	llmproxy "github.com/sourcegraph/sourcegraph/enterprise/internal/llm-proxy"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var llmProxySACredentialFilePath = env.Get("LLM_PROXY_BIGQUERY_ACCESS_CREDENTIALS_FILE", "", "")

type LLMProxyService interface {
	UsageForSubscription(ctx context.Context, uuid string) ([]SubscriptionUsage, error)
}

type LLMProxyServiceOptions struct {
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

func NewLLMProxyService() *llmProxyService {
	opts := LLMProxyServiceOptions{}

	d := conf.Get().Dotcom
	if d != nil && d.LlmProxy != nil {
		opts.BigQuery.CredentialFilePath = llmProxySACredentialFilePath
		opts.BigQuery.ProjectID = d.LlmProxy.BigQueryGoogleProjectID
		opts.BigQuery.Dataset = d.LlmProxy.BigQueryDataset
		opts.BigQuery.EventsTable = d.LlmProxy.BigQueryTable
	}

	return NewLLMProxyServiceWithOptions(opts)
}

func NewLLMProxyServiceWithOptions(opts LLMProxyServiceOptions) *llmProxyService {
	return &llmProxyService{
		opts: opts,
	}
}

type SubscriptionUsage struct {
	Date  time.Time
	Model string
	Count int
}

type llmProxyService struct {
	opts LLMProxyServiceOptions
}

func (s *llmProxyService) UsageForSubscription(ctx context.Context, feature types.CompletionsFeature, uuid string) ([]SubscriptionUsage, error) {
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
			Value: llmproxy.ProductSubscriptionActorSourceName,
		},
		{
			Name:  "identifier",
			Value: uuid,
		},
		{
			Name:  "eventName",
			Value: llmproxy.EventNameCompletionsStarted,
		},
		{
			Name:  llmproxy.CompletionsEventFeatureMetadataField,
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
			Count int
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
