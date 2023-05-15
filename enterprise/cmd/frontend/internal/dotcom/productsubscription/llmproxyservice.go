package productsubscription

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

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
	Count int
}

type llmProxyService struct {
	opts LLMProxyServiceOptions
}

func (s *llmProxyService) UsageForSubscription(ctx context.Context, uuid string) ([]SubscriptionUsage, error) {
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
SELECT
	DATE(created_at) as date,
	COUNT(*) as count
FROM
	%s.%s
WHERE
	source = @source
	AND identifier = @identifier
	AND name = @eventName
	AND DATE(created_at) >= DATE(TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY))
GROUP BY
	date
ORDER BY
	date DESC`,
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
	}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "executing query")
	}

	results := make([]SubscriptionUsage, 0)
	for {
		var row struct {
			Date  bigquery.NullDate
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
