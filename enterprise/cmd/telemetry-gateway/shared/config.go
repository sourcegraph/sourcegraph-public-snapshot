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

	PubSub struct {
		TopicName           string
		ProjectName         string
		MaxEventPayloadSize int
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

	c.PubSub.ProjectName = c.Get("TELEMETRY_GATEWAY_PUBSUB_PROJECT", "", "The project name for the PubSub topic.")
	c.PubSub.TopicName = c.Get("TELEMETRY_GATEWAY_PUBSUB_TOPICNAME", "cody_gateway", "The PubSub topic name.")
	c.PubSub.MaxEventPayloadSize = c.GetInt("TELEMETRY_GATEWAY_PUBSUB_MAX_EVENT_PAYLOAD_SIZE", "1000", "The number of events allowed in a single payload when submitting BigQuery events.")

	c.Trace.Policy = policy.TracePolicy(c.Get("TELEMETRY_GATEWAY_TRACE_POLICY", "all", "Trace policy, one of 'all', 'selective', 'none'."))
	c.Trace.GCPProjectID = c.Get("TELEMETRY_GATEWAY_TRACE_GCP_PROJECT_ID", os.Getenv("GOOGLE_CLOUD_PROJECT"), "Google Cloud Traces project ID.")
}

func (c *Config) Validate() error {
	return nil
}
