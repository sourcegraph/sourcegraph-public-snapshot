package shared

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	env.BaseConfig

	InsecureDev bool

	Address string

	DiagnosticsSecret string

	Dotcom struct {
		URL          string
		AccessToken  string
		InternalMode bool
	}

	Anthropic struct {
		AllowedModels []string
		AccessToken   string
	}

	OpenAI struct {
		AllowedModels []string
		AccessToken   string
		OrgID         string
	}

	AllowedEmbeddingsModels []string

	AllowAnonymous bool

	SourcesSyncInterval time.Duration

	BigQuery struct {
		ProjectID string
		Dataset   string
		Table     string

		EventBufferSize int
	}

	Trace TraceConfig

	ActorConcurrencyLimit codygateway.ActorConcurrencyLimitConfig
	ActorRateLimitAlert   codygateway.ActorRateLimitAlertConfig
}

type TraceConfig struct {
	Policy       policy.TracePolicy
	GCPProjectID string
}

func (c *Config) Load() {
	c.InsecureDev = env.InsecureDev
	c.Address = c.Get("CODY_GATEWAY_ADDR", ":9992", "Address to serve Cody Gateway on.")
	c.DiagnosticsSecret = c.GetOptional("CODY_GATEWAY_DIAGNOSTICS_SECRET", "Secret for accessing diagnostics - "+
		"should be used as 'Authorization: Bearer $secret' header when accessing diagnostics endpoints.")

	c.Dotcom.AccessToken = c.Get("CODY_GATEWAY_DOTCOM_ACCESS_TOKEN", "", "The Sourcegraph.com access token to be used.")
	c.Dotcom.URL = c.Get("CODY_GATEWAY_DOTCOM_API_URL", "https://sourcegraph.com/.api/graphql", "Custom override for the dotcom API endpoint")
	c.Dotcom.InternalMode = c.GetBool("CODY_GATEWAY_DOTCOM_INTERNAL_MODE", "false", "Only allow tokens associated with active internal and dev licenses to be used.") ||
		c.GetBool("CODY_GATEWAY_DOTCOM_DEV_LICENSES_ONLY", "false", "DEPRECATED, use CODY_GATEWAY_DOTCOM_INTERNAL_MODE")

	c.Anthropic.AccessToken = c.Get("CODY_GATEWAY_ANTHROPIC_ACCESS_TOKEN", "", "The Anthropic access token to be used.")
	c.Anthropic.AllowedModels = splitMaybe(c.Get("CODY_GATEWAY_ANTHROPIC_ALLOWED_MODELS",
		strings.Join([]string{
			"claude-v1",
			"claude-v1.0",
			"claude-v1.2",
			"claude-v1.3",
			"claude-instant-v1",
			"claude-instant-v1.0",
		}, ","),
		"Anthropic models that can be used."))

	c.OpenAI.AccessToken = c.GetOptional("CODY_GATEWAY_OPENAI_ACCESS_TOKEN", "The OpenAI access token to be used.")
	c.OpenAI.OrgID = c.GetOptional("CODY_GATEWAY_OPENAI_ORG_ID", "The OpenAI organization to count billing towards. Setting this ensures we always use the correct negotiated terms.")
	c.OpenAI.AllowedModels = splitMaybe(c.Get("CODY_GATEWAY_OPENAI_ALLOWED_MODELS",
		strings.Join([]string{"gpt-4", "gpt-3.5-turbo"}, ","),
		"OpenAI models that can to be used."),
	)

	c.AllowedEmbeddingsModels = splitMaybe(c.Get("CODY_GATEWAY_ALLOWED_EMBEDDINGS_MODELS", strings.Join([]string{"openai/text-embedding-ada-002"}, ","), "The models allowed for embeddings generation."))

	c.AllowAnonymous = c.GetBool("CODY_GATEWAY_ALLOW_ANONYMOUS", "false", "Allow anonymous access to Cody Gateway.")
	c.SourcesSyncInterval = c.GetInterval("CODY_GATEWAY_SOURCES_SYNC_INTERVAL", "2m", "The interval at which to sync actor sources.")

	c.BigQuery.ProjectID = c.Get("CODY_GATEWAY_BIGQUERY_PROJECT_ID", os.Getenv("GOOGLE_CLOUD_PROJECT"), "The project ID for the BigQuery events.")
	c.BigQuery.Dataset = c.Get("CODY_GATEWAY_BIGQUERY_DATASET", "cody_gateway", "The dataset for the BigQuery events.")
	c.BigQuery.Table = c.Get("CODY_GATEWAY_BIGQUERY_TABLE", "events", "The table for the BigQuery events.")
	c.BigQuery.EventBufferSize = c.GetInt("CODY_GATEWAY_BIGQUERY_EVENT_BUFFER_SIZE", "100", "The number of events allowed to buffer when submitting BigQuery events - set to 0 to disable.")

	c.Trace.Policy = policy.TracePolicy(c.Get("CODY_GATEWAY_TRACE_POLICY", "all", "Trace policy, one of 'all', 'selective', 'none'."))
	c.Trace.GCPProjectID = c.Get("CODY_GATEWAY_TRACE_GCP_PROJECT_ID", os.Getenv("GOOGLE_CLOUD_PROJECT"), "Google Cloud Traces project ID.")

	c.ActorConcurrencyLimit.Percentage = float32(c.GetPercent("CODY_GATEWAY_ACTOR_CONCURRENCY_LIMIT_PERCENTAGE", "50", "The percentage of daily rate limit to be allowed as concurrent requests limit from an actor.")) / 100
	c.ActorConcurrencyLimit.Interval = c.GetInterval("CODY_GATEWAY_ACTOR_CONCURRENCY_LIMIT_INTERVAL", "10s", "The interval at which to check the concurrent requests limit from an actor.")

	c.ActorRateLimitAlert.Threshold = float32(c.GetPercent("CODY_GATEWAY_ACTOR_RATE_LIMIT_ALERT_THRESHOLD", "90", "The the percentage of the rate limit usage to trigger an alert.")) / 100
	c.ActorRateLimitAlert.Interval = c.GetInterval("CODY_GATEWAY_ACTOR_RATE_LIMIT_ALERT_INTERVAL", "12h", "The minimum time interval between each alert for the same actor.")
	c.ActorRateLimitAlert.SlackWebhookURL = c.Get("CODY_GATEWAY_ACTOR_RATE_LIMIT_ALERT_SLACK_WEBHOOK_URL", "", "The Slack webhook URL to send alerts to.")
}

func (c *Config) Validate() error {
	_, err := url.Parse(c.Dotcom.URL)
	if err != nil {
		return err
	}

	if c.Anthropic.AccessToken != "" && len(c.Anthropic.AllowedModels) == 0 {
		c.AddError(errors.New("must provide allowed models for Anthropic"))
	}

	if c.OpenAI.AccessToken != "" && len(c.OpenAI.AllowedModels) == 0 {
		c.AddError(errors.New("must provide allowed models for OpenAI"))
	}

	if len(c.AllowedEmbeddingsModels) == 0 {
		c.AddError(errors.New("must provide allowed models for embeddings generation"))
	}

	return nil
}

// splitMaybe splits on commas, but only returns at least one element if the input
// is non-empty.
func splitMaybe(input string) []string {
	if input == "" {
		return nil
	}
	return strings.Split(input, ",")
}
