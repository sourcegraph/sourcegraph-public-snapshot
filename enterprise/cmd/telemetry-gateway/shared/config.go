package shared

import (
	"os"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

type Config struct {
	env.BaseConfig

	InsecureDev bool

	Address string

	BigQuery struct {
		ProjectID string
		Dataset   string
		Table     string

		MaxEventPayloadSize int
		EventBufferSize     int
	}

	Trace TraceConfig

	ActorConcurrencyLimit codygateway.ActorConcurrencyLimitConfig
}

type TraceConfig struct {
	Policy       policy.TracePolicy
	GCPProjectID string
}

func (c *Config) Load() {
	c.InsecureDev = env.InsecureDev
	c.Address = c.Get("TELEMETRY_GATEWAY_ADDR", ":9992", "Address to serve Cody Gateway on.")

	c.BigQuery.ProjectID = c.Get("TELEMETRY_GATEWAY_BIGQUERY_PROJECT_ID", os.Getenv("GOOGLE_CLOUD_PROJECT"), "The project ID for the BigQuery events.")
	c.BigQuery.Dataset = c.Get("TELEMETRY_GATEWAY_BIGQUERY_DATASET", "cody_gateway", "The dataset for the BigQuery events.")
	c.BigQuery.Table = c.Get("TELEMETRY_GATEWAY_BIGQUERY_TABLE", "events", "The table for the BigQuery events.")
	c.BigQuery.MaxEventPayloadSize = c.GetInt("TELEMETRY_GATEWAY_BIGQUERY_MAX_EVENT_PAYLOAD_SIZE", "1000", "The number of events allowed in a single payload when submitting BigQuery events.")

	c.Trace.Policy = policy.TracePolicy(c.Get("TELEMETRY_GATEWAY_TRACE_POLICY", "all", "Trace policy, one of 'all', 'selective', 'none'."))
	c.Trace.GCPProjectID = c.Get("TELEMETRY_GATEWAY_TRACE_GCP_PROJECT_ID", os.Getenv("GOOGLE_CLOUD_PROJECT"), "Google Cloud Traces project ID.")
}

func (c *Config) Validate() error {
	return nil
}
