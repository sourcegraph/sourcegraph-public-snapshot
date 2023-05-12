package shared

import (
	"net/url"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
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

	AllowAnonymous bool

	SourcesSyncInterval time.Duration

	BigQuery struct {
		ProjectID string
		Dataset   string
		Table     string
	}
}

func (c *Config) Load() {
	c.InsecureDev = env.InsecureDev
	c.Address = c.Get("LLM_PROXY_ADDR", ":9992", "Address to serve LLM proxy on.")
	c.Dotcom.AccessToken = c.Get("LLM_PROXY_DOTCOM_ACCESS_TOKEN", "", "The Sourcegraph.com access token to be used.")
	c.Dotcom.URL = c.Get("LLM_PROXY_DOTCOM_API_URL", "https://sourcegraph.com/.api/graphql", "Custom override for the dotcom API endpoint")
	c.Dotcom.DevLicensesOnly = c.GetBool("LLM_PROXY_DOTCOM_DEV_LICENSES_ONLY", "false", "Only allow tokens associated with active dev licenses to be used.")

	c.Anthropic.AccessToken = c.Get("LLM_PROXY_ANTHROPIC_ACCESS_TOKEN", "", "The Anthropic access token to be used.")
	c.AllowAnonymous = c.GetBool("LLM_PROXY_ALLOW_ANONYMOUS", "false", "Allow anonymous access to LLM proxy.")
	c.SourcesSyncInterval = c.GetInterval("LLM_PROXY_SOURCES_SYNC_INTERVAL", "2m", "The interval at which to sync actor sources.")

	c.BigQuery.ProjectID = c.Get("LLM_PROXY_BIGQUERY_PROJECT_ID", os.Getenv("GOOGLE_CLOUD_PROJECT"), "The project ID for the BigQuery events.")
	c.BigQuery.Dataset = c.Get("LLM_PROXY_BIGQUERY_DATASET", "llm_proxy", "The dataset for the BigQuery events.")
	c.BigQuery.Table = c.Get("LLM_PROXY_BIGQUERY_TABLE", "events", "The table for the BigQuery events.")
}

func (c *Config) Validate() error {
	_, err := url.Parse(c.Dotcom.URL)
	if err != nil {
		return err
	}

	return nil
}
