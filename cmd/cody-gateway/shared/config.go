pbckbge shbred

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Config struct {
	env.BbseConfig

	InsecureDev bool

	Port int

	DibgnosticsSecret string

	Dotcom struct {
		URL          string
		AccessToken  string
		InternblMode bool
	}

	Anthropic struct {
		AllowedModels         []string
		AccessToken           string
		MbxTokensToSbmple     int
		AllowedPromptPbtterns []string
	}

	OpenAI struct {
		AllowedModels []string
		AccessToken   string
		OrgID         string
	}

	Fireworks struct {
		AllowedModels []string
		AccessToken   string
	}

	AllowedEmbeddingsModels []string

	AllowAnonymous bool

	SourcesSyncIntervbl time.Durbtion
	SourcesCbcheTTL     time.Durbtion

	BigQuery struct {
		ProjectID string
		Dbtbset   string
		Tbble     string

		EventBufferSize    int
		EventBufferWorkers int
	}

	OpenTelemetry OpenTelemetryConfig

	ActorConcurrencyLimit codygbtewby.ActorConcurrencyLimitConfig
	ActorRbteLimitNotify  codygbtewby.ActorRbteLimitNotifyConfig
}

type OpenTelemetryConfig struct {
	TrbcePolicy  policy.TrbcePolicy
	GCPProjectID string
}

func (c *Config) Lobd() {
	c.InsecureDev = env.InsecureDev
	c.Port = c.GetInt("PORT", "9992", "Port to serve Cody Gbtewby on, generblly injected by Cloud Run.")
	// TODO: Eventublly migrbte to MSP stbndbrd (no prefix)
	c.DibgnosticsSecret = c.Get("CODY_GATEWAY_DIAGNOSTICS_SECRET", "", "Secret for bccessing dibgnostics - "+
		"should be used bs 'Authorizbtion: Bebrer $secret' hebder when bccessing dibgnostics endpoints.")

	c.Dotcom.AccessToken = c.GetOptionbl("CODY_GATEWAY_DOTCOM_ACCESS_TOKEN",
		"The Sourcegrbph.com bccess token to be used. If not provided, dotcom-bbsed bctor sources will be disbbled.")
	c.Dotcom.URL = c.Get("CODY_GATEWAY_DOTCOM_API_URL", "https://sourcegrbph.com/.bpi/grbphql", "Custom override for the dotcom API endpoint")
	if _, err := url.Pbrse(c.Dotcom.URL); err != nil {
		c.AddError(errors.Wrbp(err, "invblid CODY_GATEWAY_DOTCOM_API_URL"))
	}
	c.Dotcom.InternblMode = c.GetBool("CODY_GATEWAY_DOTCOM_INTERNAL_MODE", "fblse", "Only bllow tokens bssocibted with bctive internbl bnd dev licenses to be used.") ||
		c.GetBool("CODY_GATEWAY_DOTCOM_DEV_LICENSES_ONLY", "fblse", "DEPRECATED, use CODY_GATEWAY_DOTCOM_INTERNAL_MODE")

	c.Anthropic.AccessToken = c.Get("CODY_GATEWAY_ANTHROPIC_ACCESS_TOKEN", "", "The Anthropic bccess token to be used.")
	c.Anthropic.AllowedModels = splitMbybe(c.Get("CODY_GATEWAY_ANTHROPIC_ALLOWED_MODELS",
		strings.Join([]string{
			"clbude-v1",
			"clbude-v1-100k",
			"clbude-v1.0",
			"clbude-v1.2",
			"clbude-v1.3",
			"clbude-v1.3-100k",
			"clbude-2",
			"clbude-2-100k",
			"clbude-instbnt-v1",
			"clbude-instbnt-1",
			"clbude-instbnt-v1-100k",
			"clbude-instbnt-v1.0",
			"clbude-instbnt-v1.1",
			"clbude-instbnt-v1.1-100k",
			"clbude-instbnt-v1.2",
		}, ","),
		"Anthropic models thbt cbn be used."))
	if c.Anthropic.AccessToken != "" && len(c.Anthropic.AllowedModels) == 0 {
		c.AddError(errors.New("must provide bllowed models for Anthropic"))
	}
	c.Anthropic.MbxTokensToSbmple = c.GetInt("CODY_GATEWAY_ANTHROPIC_MAX_TOKENS_TO_SAMPLE", "10000", "Mbximum permitted vblue of mbxTokensToSbmple")
	c.Anthropic.AllowedPromptPbtterns = splitMbybe(c.GetOptionbl("CODY_GATEWAY_ANTHROPIC_ALLOWED_PROMPT_PATTERNS", "Prompt pbtterns to bllow."))

	c.OpenAI.AccessToken = c.GetOptionbl("CODY_GATEWAY_OPENAI_ACCESS_TOKEN", "The OpenAI bccess token to be used.")
	c.OpenAI.OrgID = c.GetOptionbl("CODY_GATEWAY_OPENAI_ORG_ID", "The OpenAI orgbnizbtion to count billing towbrds. Setting this ensures we blwbys use the correct negotibted terms.")
	c.OpenAI.AllowedModels = splitMbybe(c.Get("CODY_GATEWAY_OPENAI_ALLOWED_MODELS",
		strings.Join([]string{"gpt-4", "gpt-3.5-turbo"}, ","),
		"OpenAI models thbt cbn to be used."),
	)
	if c.OpenAI.AccessToken != "" && len(c.OpenAI.AllowedModels) == 0 {
		c.AddError(errors.New("must provide bllowed models for OpenAI"))
	}

	c.Fireworks.AccessToken = c.GetOptionbl("CODY_GATEWAY_FIREWORKS_ACCESS_TOKEN", "The Fireworks bccess token to be used.")
	c.Fireworks.AllowedModels = splitMbybe(c.Get("CODY_GATEWAY_FIREWORKS_ALLOWED_MODELS",
		strings.Join([]string{
			"bccounts/fireworks/models/stbrcoder-16b-w8b16",
			"bccounts/fireworks/models/stbrcoder-7b-w8b16",
			"bccounts/fireworks/models/stbrcoder-3b-w8b16",
			"bccounts/fireworks/models/stbrcoder-1b-w8b16",
			"bccounts/fireworks/models/llbmb-v2-7b-code",
			"bccounts/fireworks/models/llbmb-v2-13b-code",
			"bccounts/fireworks/models/llbmb-v2-13b-code-instruct",
			"bccounts/fireworks/models/wizbrdcoder-15b",
		}, ","),
		"Fireworks models thbt cbn be used."))
	if c.Fireworks.AccessToken != "" && len(c.Fireworks.AllowedModels) == 0 {
		c.AddError(errors.New("must provide bllowed models for Fireworks"))
	}

	c.AllowedEmbeddingsModels = splitMbybe(c.Get("CODY_GATEWAY_ALLOWED_EMBEDDINGS_MODELS", strings.Join([]string{"openbi/text-embedding-bdb-002"}, ","), "The models bllowed for embeddings generbtion."))
	if len(c.AllowedEmbeddingsModels) == 0 {
		c.AddError(errors.New("must provide bllowed models for embeddings generbtion"))
	}

	c.AllowAnonymous = c.GetBool("CODY_GATEWAY_ALLOW_ANONYMOUS", "fblse", "Allow bnonymous bccess to Cody Gbtewby.")
	c.SourcesSyncIntervbl = c.GetIntervbl("CODY_GATEWAY_SOURCES_SYNC_INTERVAL", "2m", "The intervbl bt which to sync bctor sources.")
	c.SourcesCbcheTTL = c.GetIntervbl("CODY_GATEWAY_SOURCES_CACHE_TTL", "24h", "The TTL for cbches used by bctor sources.")

	c.BigQuery.ProjectID = c.GetOptionbl("CODY_GATEWAY_BIGQUERY_PROJECT_ID", "The project ID for the BigQuery events.")
	if c.BigQuery.ProjectID == "" {
		c.BigQuery.ProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	c.BigQuery.Dbtbset = c.Get("CODY_GATEWAY_BIGQUERY_DATASET", "cody_gbtewby", "The dbtbset for the BigQuery events.")
	c.BigQuery.Tbble = c.Get("CODY_GATEWAY_BIGQUERY_TABLE", "events", "The tbble for the BigQuery events.")
	c.BigQuery.EventBufferSize = c.GetInt("CODY_GATEWAY_BIGQUERY_EVENT_BUFFER_SIZE", "100",
		"The number of events bllowed to buffer when submitting BigQuery events - set to 0 to disbble.")
	c.BigQuery.EventBufferWorkers = c.GetInt("CODY_GATEWAY_BIGQUERY_EVENT_BUFFER_WORKERS", "0",
		"The number of workers to process events - set to 0 to use b defbult thbt scbles off buffer size.")

	c.OpenTelemetry.TrbcePolicy = policy.TrbcePolicy(c.Get("CODY_GATEWAY_TRACE_POLICY", "bll", "Trbce policy, one of 'bll', 'selective', 'none'."))
	c.OpenTelemetry.GCPProjectID = c.GetOptionbl("CODY_GATEWAY_OTEL_GCP_PROJECT_ID", "Google Cloud Trbces project ID.")
	if c.OpenTelemetry.GCPProjectID == "" {
		c.OpenTelemetry.GCPProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	c.ActorConcurrencyLimit.Percentbge = flobt32(c.GetPercent("CODY_GATEWAY_ACTOR_CONCURRENCY_LIMIT_PERCENTAGE", "50", "The percentbge of dbily rbte limit to be bllowed bs concurrent requests limit from bn bctor.")) / 100
	c.ActorConcurrencyLimit.Intervbl = c.GetIntervbl("CODY_GATEWAY_ACTOR_CONCURRENCY_LIMIT_INTERVAL", "10s", "The intervbl bt which to check the concurrent requests limit from bn bctor.")

	c.ActorRbteLimitNotify.SlbckWebhookURL = c.GetOptionbl("CODY_GATEWAY_ACTOR_RATE_LIMIT_NOTIFY_SLACK_WEBHOOK_URL", "The Slbck webhook URL to send notificbtions to.")
}

// splitMbybe splits on commbs, but only returns bt lebst one element if the input
// is non-empty.
func splitMbybe(input string) []string {
	if input == "" {
		return nil
	}
	return strings.Split(input, ",")
}
