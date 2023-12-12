package shared

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	env.BaseConfig

	InsecureDev bool

	Port int

	DiagnosticsSecret string

	Dotcom struct {
		URL          string
		AccessToken  string
		InternalMode bool
	}

	Anthropic struct {
		AllowedModels          []string
		AccessToken            string
		MaxTokensToSample      int
		AllowedPromptPatterns  []string
		RequestBlockingEnabled bool
	}

	OpenAI struct {
		AllowedModels []string
		AccessToken   string
		OrgID         string
	}

	Fireworks struct {
		AllowedModels                      []string
		AccessToken                        string
		LogSelfServeCodeCompletionRequests bool
		DisableSingleTenant                bool
	}

	AllowedEmbeddingsModels []string

	AllowAnonymous bool

	SourcesSyncInterval time.Duration
	SourcesCacheTTL     time.Duration

	BigQuery struct {
		ProjectID string
		Dataset   string
		Table     string

		EventBufferSize    int
		EventBufferWorkers int
	}

	OpenTelemetry OpenTelemetryConfig

	ActorConcurrencyLimit       codygateway.ActorConcurrencyLimitConfig
	ActorRateLimitNotify        codygateway.ActorRateLimitNotifyConfig
	AutoFlushStreamingResponses bool
}

type OpenTelemetryConfig struct {
	TracePolicy  policy.TracePolicy
	GCPProjectID string
}

func (c *Config) Load() {
	c.InsecureDev = env.InsecureDev
	c.Port = c.GetInt("PORT", "9992", "Port to serve Cody Gateway on, generally injected by Cloud Run.")
	// TODO: Eventually migrate to MSP standard (no prefix)
	c.DiagnosticsSecret = c.Get("CODY_GATEWAY_DIAGNOSTICS_SECRET", "", "Secret for accessing diagnostics - "+
		"should be used as 'Authorization: Bearer $secret' header when accessing diagnostics endpoints.")

	c.Dotcom.AccessToken = c.GetOptional("CODY_GATEWAY_DOTCOM_ACCESS_TOKEN",
		"The Sourcegraph.com access token to be used. If not provided, dotcom-based actor sources will be disabled.")
	c.Dotcom.URL = c.Get("CODY_GATEWAY_DOTCOM_API_URL", "https://sourcegraph.com/.api/graphql", "Custom override for the dotcom API endpoint")
	if _, err := url.Parse(c.Dotcom.URL); err != nil {
		c.AddError(errors.Wrap(err, "invalid CODY_GATEWAY_DOTCOM_API_URL"))
	}
	c.Dotcom.InternalMode = c.GetBool("CODY_GATEWAY_DOTCOM_INTERNAL_MODE", "false", "Only allow tokens associated with active internal and dev licenses to be used.") ||
		c.GetBool("CODY_GATEWAY_DOTCOM_DEV_LICENSES_ONLY", "false", "DEPRECATED, use CODY_GATEWAY_DOTCOM_INTERNAL_MODE")

	c.Anthropic.AccessToken = c.Get("CODY_GATEWAY_ANTHROPIC_ACCESS_TOKEN", "", "The Anthropic access token to be used.")
	c.Anthropic.AllowedModels = splitMaybe(c.Get("CODY_GATEWAY_ANTHROPIC_ALLOWED_MODELS",
		strings.Join([]string{
			"claude-v1",
			"claude-v1-100k",
			"claude-v1.0",
			"claude-v1.2",
			"claude-v1.3",
			"claude-v1.3-100k",
			"claude-2",
			"claude-2.0",
			"claude-2.1",
			"claude-2-100k",
			"claude-instant-v1",
			"claude-instant-1",
			"claude-instant-v1-100k",
			"claude-instant-v1.0",
			"claude-instant-v1.1",
			"claude-instant-v1.1-100k",
			"claude-instant-v1.2",
			"claude-instant-1.2-cyan",
		}, ","),
		"Anthropic models that can be used."))
	if c.Anthropic.AccessToken != "" && len(c.Anthropic.AllowedModels) == 0 {
		c.AddError(errors.New("must provide allowed models for Anthropic"))
	}
	c.Anthropic.MaxTokensToSample = c.GetInt("CODY_GATEWAY_ANTHROPIC_MAX_TOKENS_TO_SAMPLE", "10000", "Maximum permitted value of maxTokensToSample")
	c.Anthropic.AllowedPromptPatterns = splitMaybe(c.GetOptional("CODY_GATEWAY_ANTHROPIC_ALLOWED_PROMPT_PATTERNS", "Prompt patterns to allow."))
	c.Anthropic.RequestBlockingEnabled = c.GetBool("CODY_GATEWAY_ANTHROPIC_REQUEST_BLOCKING_ENABLED", "false", "Whether we should block requests that match our blocking criteria.")

	c.OpenAI.AccessToken = c.GetOptional("CODY_GATEWAY_OPENAI_ACCESS_TOKEN", "The OpenAI access token to be used.")
	c.OpenAI.OrgID = c.GetOptional("CODY_GATEWAY_OPENAI_ORG_ID", "The OpenAI organization to count billing towards. Setting this ensures we always use the correct negotiated terms.")
	c.OpenAI.AllowedModels = splitMaybe(c.Get("CODY_GATEWAY_OPENAI_ALLOWED_MODELS",
		strings.Join([]string{"gpt-4", "gpt-3.5-turbo", "gpt-4-1106-preview"}, ","),
		"OpenAI models that can to be used."),
	)
	if c.OpenAI.AccessToken != "" && len(c.OpenAI.AllowedModels) == 0 {
		c.AddError(errors.New("must provide allowed models for OpenAI"))
	}

	c.Fireworks.AccessToken = c.GetOptional("CODY_GATEWAY_FIREWORKS_ACCESS_TOKEN", "The Fireworks access token to be used.")
	c.Fireworks.AllowedModels = splitMaybe(c.Get("CODY_GATEWAY_FIREWORKS_ALLOWED_MODELS",
		strings.Join([]string{
			"accounts/fireworks/models/starcoder-16b-w8a16",
			"accounts/fireworks/models/starcoder-7b-w8a16",
			"accounts/fireworks/models/starcoder-3b-w8a16",
			"accounts/fireworks/models/starcoder-1b-w8a16",
			"accounts/sourcegraph/models/starcoder-7b",
			"accounts/sourcegraph/models/starcoder-16b",
			"accounts/fireworks/models/llama-v2-7b-code",
			"accounts/fireworks/models/llama-v2-13b-code",
			"accounts/fireworks/models/llama-v2-13b-code-instruct",
			"accounts/fireworks/models/llama-v2-34b-code-instruct",
			"accounts/fireworks/models/mistral-7b-instruct-4k",
			"accounts/fireworks/models/mixtral-8x7b-instruct",
			"accounts/fireworks/models/mixtral-8x7b-fw-chat",
		}, ","),
		"Fireworks models that can be used."))
	c.Fireworks.LogSelfServeCodeCompletionRequests = c.GetBool("CODY_GATEWAY_FIREWORKS_LOG_SELF_SERVE_COMPLETION_REQUESTS", "false", "Whether we should log self-serve code completion requests.")
	c.Fireworks.DisableSingleTenant = c.GetBool("CODY_GATEWAY_FIREWORKS_DISABLE_SINGLE_TENANT", "false", "Whether we should disable single tenant models for Fireworks.")
	if c.Fireworks.AccessToken != "" && len(c.Fireworks.AllowedModels) == 0 {
		c.AddError(errors.New("must provide allowed models for Fireworks"))
	}

	c.AllowedEmbeddingsModels = splitMaybe(c.Get("CODY_GATEWAY_ALLOWED_EMBEDDINGS_MODELS", strings.Join([]string{"openai/text-embedding-ada-002"}, ","), "The models allowed for embeddings generation."))
	if len(c.AllowedEmbeddingsModels) == 0 {
		c.AddError(errors.New("must provide allowed models for embeddings generation"))
	}

	c.AllowAnonymous = c.GetBool("CODY_GATEWAY_ALLOW_ANONYMOUS", "false", "Allow anonymous access to Cody Gateway.")
	c.SourcesSyncInterval = c.GetInterval("CODY_GATEWAY_SOURCES_SYNC_INTERVAL", "2m", "The interval at which to sync actor sources.")
	c.SourcesCacheTTL = c.GetInterval("CODY_GATEWAY_SOURCES_CACHE_TTL", "24h", "The TTL for caches used by actor sources.")

	c.BigQuery.ProjectID = c.GetOptional("CODY_GATEWAY_BIGQUERY_PROJECT_ID", "The project ID for the BigQuery events.")
	if c.BigQuery.ProjectID == "" {
		c.BigQuery.ProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	c.BigQuery.Dataset = c.Get("CODY_GATEWAY_BIGQUERY_DATASET", "cody_gateway", "The dataset for the BigQuery events.")
	c.BigQuery.Table = c.Get("CODY_GATEWAY_BIGQUERY_TABLE", "events", "The table for the BigQuery events.")
	c.BigQuery.EventBufferSize = c.GetInt("CODY_GATEWAY_BIGQUERY_EVENT_BUFFER_SIZE", "100",
		"The number of events allowed to buffer when submitting BigQuery events - set to 0 to disable.")
	c.BigQuery.EventBufferWorkers = c.GetInt("CODY_GATEWAY_BIGQUERY_EVENT_BUFFER_WORKERS", "0",
		"The number of workers to process events - set to 0 to use a default that scales off buffer size.")

	c.OpenTelemetry.TracePolicy = policy.TracePolicy(c.Get("CODY_GATEWAY_TRACE_POLICY", "all", "Trace policy, one of 'all', 'selective', 'none'."))
	c.OpenTelemetry.GCPProjectID = c.GetOptional("CODY_GATEWAY_OTEL_GCP_PROJECT_ID", "Google Cloud Traces project ID.")
	if c.OpenTelemetry.GCPProjectID == "" {
		c.OpenTelemetry.GCPProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	c.ActorConcurrencyLimit.Percentage = float32(c.GetPercent("CODY_GATEWAY_ACTOR_CONCURRENCY_LIMIT_PERCENTAGE", "50", "The percentage of daily rate limit to be allowed as concurrent requests limit from an actor.")) / 100
	c.ActorConcurrencyLimit.Interval = c.GetInterval("CODY_GATEWAY_ACTOR_CONCURRENCY_LIMIT_INTERVAL", "10s", "The interval at which to check the concurrent requests limit from an actor.")

	c.ActorRateLimitNotify.SlackWebhookURL = c.GetOptional("CODY_GATEWAY_ACTOR_RATE_LIMIT_NOTIFY_SLACK_WEBHOOK_URL", "The Slack webhook URL to send notifications to.")
	c.AutoFlushStreamingResponses = c.GetBool("CODY_GATEWAY_AUTO_FLUSH_STREAMING_RESPONSES", "false", "Whether we should flush streaming responses after every write.")
}

// splitMaybe splits on commas, but only returns at least one element if the input
// is non-empty.
func splitMaybe(input string) []string {
	if input == "" {
		return nil
	}
	return strings.Split(input, ",")
}
