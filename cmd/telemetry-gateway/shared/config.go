package shared

import (
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

type Config struct {
	env.BaseConfig

	Port              int
	DiagnosticsSecret string

	Events struct {
		PubSub struct {
			Enabled   bool
			ProjectID string
			TopicID   string
		}
	}

	OpenTelemetry OpenTelemetryConfig
}

type OpenTelemetryConfig struct {
	TracePolicy  policy.TracePolicy
	GCPProjectID string
}

func (c *Config) Load() {
	c.Port = c.GetInt("PORT", "10080", "Port to serve Telemetry Gateway service on, generally injected by Cloud Run.")
	c.DiagnosticsSecret = c.Get("DIAGNOSTICS_SECRET", "", "Secret for accessing diagnostics - "+
		"should be used as 'Authorization: Bearer $secret' header when accessing diagnostics endpoints.")

	c.Events.PubSub.Enabled = c.GetBool("TELEMETRY_GATEWAY_EVENTS_PUBSUB_ENABLED", "true",
		"If false, logs Pub/Sub messages instead of actually sending them")
	c.Events.PubSub.ProjectID = c.GetOptional("TELEMETRY_GATEWAY_EVENTS_PUBSUB_PROJECT_ID",
		"The project ID for the Pub/Sub.")
	c.Events.PubSub.TopicID = c.GetOptional("TELEMETRY_GATEWAY_EVENTS_PUBSUB_TOPIC_ID",
		"The topic ID for the Pub/Sub.")

	c.OpenTelemetry.TracePolicy = policy.TracePolicy(c.Get("TELEMETRY_GATEWAY_TRACE_POLICY", "all", "Trace policy, one of 'all', 'selective', 'none'."))
	c.OpenTelemetry.GCPProjectID = c.GetOptional("TELEMETRY_GATEWAY_OTEL_GCP_PROJECT_ID", "Google Cloud Traces project ID.")
	if c.OpenTelemetry.GCPProjectID == "" {
		c.OpenTelemetry.GCPProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
}

func (c *Config) GetListenAdress() string {
	return fmt.Sprintf(":%d", c.Port)
}
