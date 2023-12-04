package shared

import (
	"os"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Port              int
	DiagnosticsSecret string

	PubSub struct {
		ProjectID string
		TopicID   string
	}

	OpenTelemetry OpenTelemetryConfig
}

type OpenTelemetryConfig struct {
	GCPProjectID string
}

func (c *Config) Load() {
	c.Port = c.GetInt("PORT", "10086", "Port to serve Pings service on, generally injected by Cloud Run.")
	c.DiagnosticsSecret = c.Get("DIAGNOSTICS_SECRET", "", "Secret for accessing diagnostics - "+
		"should be used as 'Authorization: Bearer $secret' header when accessing diagnostics endpoints.")

	c.PubSub.ProjectID = c.Get("PINGS_PUBSUB_PROJECT_ID", "", "The project ID for the Pub/Sub.")
	c.PubSub.TopicID = c.Get("PINGS_PUBSUB_TOPIC_ID", "", "The topic ID for the Pub/Sub.")

	c.OpenTelemetry.GCPProjectID = c.GetOptional("PINGS_OTEL_GCP_PROJECT_ID", "Google Cloud Traces project ID.")
	if c.OpenTelemetry.GCPProjectID == "" {
		c.OpenTelemetry.GCPProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
}
