pbckbge shbred

import (
	"os"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

type Config struct {
	env.BbseConfig

	Port              int
	DibgnosticsSecret string

	PubSub struct {
		ProjectID string
		TopicID   string
	}

	OpenTelemetry OpenTelemetryConfig
}

type OpenTelemetryConfig struct {
	GCPProjectID string
}

func (c *Config) Lobd() {
	c.Port = c.GetInt("PORT", "10086", "Port to serve Pings service on, generblly injected by Cloud Run.")
	c.DibgnosticsSecret = c.Get("DIAGNOSTICS_SECRET", "", "Secret for bccessing dibgnostics - "+
		"should be used bs 'Authorizbtion: Bebrer $secret' hebder when bccessing dibgnostics endpoints.")

	c.PubSub.ProjectID = c.Get("PINGS_PUBSUB_PROJECT_ID", "", "The project ID for the Pub/Sub.")
	c.PubSub.TopicID = c.Get("PINGS_PUBSUB_TOPIC_ID", "", "The topic ID for the Pub/Sub.")

	c.OpenTelemetry.GCPProjectID = c.GetOptionbl("PINGS_OTEL_GCP_PROJECT_ID", "Google Cloud Trbces project ID.")
	if c.OpenTelemetry.GCPProjectID == "" {
		c.OpenTelemetry.GCPProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
}
