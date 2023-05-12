package productsubscription

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	llmproxy "github.com/sourcegraph/sourcegraph/enterprise/internal/llm-proxy"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type llmProxyAccessResolver struct{ sub *productSubscription }

func (r llmProxyAccessResolver) Enabled() bool { return r.sub.v.LLMProxyAccess.Enabled }

func (r llmProxyAccessResolver) RateLimit(ctx context.Context) (graphqlbackend.LLMProxyRateLimit, error) {
	if !r.sub.v.LLMProxyAccess.Enabled {
		return nil, nil
	}

	var rateLimit licensing.LLMProxyRateLimit

	// Get default access from active license. Call hydrate and access field directly to
	// avoid parsing license key which is done in (*productLicense).Info(), instead just
	// relying on what we know in DB.
	activeLicense, err := r.sub.computeActiveLicense(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get active license")
	}
	var source graphqlbackend.LLMProxyRateLimitSource
	if activeLicense != nil {
		source = graphqlbackend.LLMProxyRateLimitSourcePlan
		rateLimit = licensing.NewLLMProxyRateLimit(licensing.PlanFromTags(activeLicense.LicenseTags))
	}

	// Apply overrides
	rateLimitOverrides := r.sub.v.LLMProxyAccess
	if rateLimitOverrides.RateLimit != nil {
		source = graphqlbackend.LLMProxyRateLimitSourceOverride
		rateLimit.Limit = *rateLimitOverrides.RateLimit
	}
	if rateLimitOverrides.RateIntervalSeconds != nil {
		source = graphqlbackend.LLMProxyRateLimitSourceOverride
		rateLimit.IntervalSeconds = *rateLimitOverrides.RateIntervalSeconds
	}

	return &llmProxyRateLimitResolver{v: rateLimit, source: source}, nil
}

func (r llmProxyAccessResolver) Usage(ctx context.Context) ([]graphqlbackend.LLMProxyUsageDatapoint, error) {
	d := conf.Get().Dotcom
	if d == nil || d.LlmProxy == nil || d.LlmProxy.BigQueryGoogleProjectID == "" || d.LlmProxy.BigQueryDataset == "" || d.LlmProxy.BigQueryTable == "" {
		// Not configured, nothing we can do.
		return nil, nil
	}

	client, err := bigquery.NewClient(ctx, d.LlmProxy.BigQueryGoogleProjectID)
	if err != nil {
		return nil, errors.Wrap(err, "creating BigQuery client")
	}
	defer client.Close()

	tbl := client.Dataset(d.LlmProxy.BigQueryDataset).Table(d.LlmProxy.BigQueryTable)

	// Count events with a specific name for each day in the last 7 days.
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
			Value: r.sub.UUID(),
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

	resolvers := make([]graphqlbackend.LLMProxyUsageDatapoint, 0)
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
		resolvers = append(resolvers, &llmProxyUsageDatapoint{
			date:  row.Date.Date.In(time.UTC),
			count: row.Count,
		})
	}

	return resolvers, nil
}

type llmProxyRateLimitResolver struct {
	source graphqlbackend.LLMProxyRateLimitSource
	v      licensing.LLMProxyRateLimit
}

func (r *llmProxyRateLimitResolver) Source() graphqlbackend.LLMProxyRateLimitSource { return r.source }
func (r *llmProxyRateLimitResolver) Limit() int32                                   { return r.v.Limit }
func (r *llmProxyRateLimitResolver) IntervalSeconds() int32                         { return r.v.IntervalSeconds }

type llmProxyUsageDatapoint struct {
	date  time.Time
	count int
}

func (r *llmProxyUsageDatapoint) Date() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.date}
}

func (r *llmProxyUsageDatapoint) Count() int32 {
	return int32(r.count)
}
