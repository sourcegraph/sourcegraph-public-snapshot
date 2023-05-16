package shared

import (
	"net/url"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

type Config struct {
	env.BaseConfig

	InsecureDev bool

	Address string

	Dotcom struct {
		URL             string
		AccessToken     string
		DevLicensesOnly bool
	}

	Anthropic struct {
		AccessToken string
	}

	OpenAI struct {
		AccessToken string
		OrgID       string
	}

	AllowAnonymous bool

	SourcesSyncInterval time.Duration

	BigQuery struct {
		ProjectID string
		Dataset   string
		Table     string
	}

	Trace TraceConfig
}

type TraceConfig struct {
	Policy       policy.TracePolicy
	GCPProjectID string
}

func (c *Config) Load() {
	c.InsecureDev = env.InsecureDev
	c.Address = c.Get("LLM_PROXY_ADDR", ":9992", "Address to serve LLM proxy on.")
	c.Dotcom.AccessToken = c.Get("LLM_PROXY_DOTCOM_ACCESS_TOKEN", "", "The Sourcegraph.com access token to be used.")
	c.Dotcom.URL = c.Get("LLM_PROXY_DOTCOM_API_URL", "https://sourcegraph.com/.api/graphql", "Custom override for the dotcom API endpoint")
	c.Dotcom.DevLicensesOnly = c.GetBool("LLM_PROXY_DOTCOM_DEV_LICENSES_ONLY", "false", "Only allow tokens associated with active dev licenses to be used.")

	c.Anthropic.AccessToken = c.Get("LLM_PROXY_ANTHROPIC_ACCESS_TOKEN", "", "The Anthropic access token to be used.")
	c.OpenAI.AccessToken = c.GetOptional("LLM_PROXY_OPENAI_ACCESS_TOKEN", "The OpenAI access token to be used.")
	c.OpenAI.OrgID = c.GetOptional("LLM_PROXY_OPENAI_ORG_ID", "The OpenAI organization to count billing towards. Setting this ensures we always use the correct negotiated terms.")
	c.AllowAnonymous = c.GetBool("LLM_PROXY_ALLOW_ANONYMOUS", "false", "Allow anonymous access to LLM proxy.")
	c.SourcesSyncInterval = c.GetInterval("LLM_PROXY_SOURCES_SYNC_INTERVAL", "2m", "The interval at which to sync actor sources.")

	c.BigQuery.ProjectID = c.Get("LLM_PROXY_BIGQUERY_PROJECT_ID", os.Getenv("GOOGLE_CLOUD_PROJECT"), "The project ID for the BigQuery events.")
	c.BigQuery.Dataset = c.Get("LLM_PROXY_BIGQUERY_DATASET", "llm_proxy", "The dataset for the BigQuery events.")
	c.BigQuery.Table = c.Get("LLM_PROXY_BIGQUERY_TABLE", "events", "The table for the BigQuery events.")

	c.Trace.Policy = policy.TracePolicy(c.Get("LLM_PROXY_TRACE_POLICY", "all", "Trace policy, one of 'all', 'selective', 'none'."))
	c.Trace.GCPProjectID = c.Get("LLM_PROXY_TRACE_GCP_PROJECT_ID", os.Getenv("GOOGLE_CLOUD_PROJECT"), "Google Cloud Traces project ID.")
}

func (c *Config) Validate() error {
	_, err := url.Parse(c.Dotcom.URL)
	if err != nil {
		return err
	}

	return nil
}
