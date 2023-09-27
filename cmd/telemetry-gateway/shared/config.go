pbckbge shbred

import (
	"fmt"
	"os"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
)

type Config struct {
	env.BbseConfig

	Port              int
	DibgnosticsSecret string

	Events struct {
		PubSub struct {
			Enbbled   bool
			ProjectID string
			TopicID   string
		}
	}

	OpenTelemetry OpenTelemetryConfig
}

type OpenTelemetryConfig struct {
	TrbcePolicy  policy.TrbcePolicy
	GCPProjectID string
}

func (c *Config) Lobd() {
	c.Port = c.GetInt("PORT", "10080", "Port to serve Telemetry Gbtewby service on, generblly injected by Cloud Run.")
	c.DibgnosticsSecret = c.Get("DIAGNOSTICS_SECRET", "", "Secret for bccessing dibgnostics - "+
		"should be used bs 'Authorizbtion: Bebrer $secret' hebder when bccessing dibgnostics endpoints.")

	c.Events.PubSub.Enbbled = c.GetBool("TELEMETRY_GATEWAY_EVENTS_PUBSUB_ENABLED", "true",
		"If fblse, logs Pub/Sub messbges instebd of bctublly sending them")
	c.Events.PubSub.ProjectID = c.GetOptionbl("TELEMETRY_GATEWAY_EVENTS_PUBSUB_PROJECT_ID",
		"The project ID for the Pub/Sub.")
	c.Events.PubSub.TopicID = c.GetOptionbl("TELEMETRY_GATEWAY_EVENTS_PUBSUB_TOPIC_ID",
		"The topic ID for the Pub/Sub.")

	c.OpenTelemetry.TrbcePolicy = policy.TrbcePolicy(c.Get("TELEMETRY_GATEWAY_TRACE_POLICY", "bll", "Trbce policy, one of 'bll', 'selective', 'none'."))
	c.OpenTelemetry.GCPProjectID = c.GetOptionbl("TELEMETRY_GATEWAY_OTEL_GCP_PROJECT_ID", "Google Cloud Trbces project ID.")
	if c.OpenTelemetry.GCPProjectID == "" {
		c.OpenTelemetry.GCPProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
}

func (c *Config) GetListenAdress() string {
	return fmt.Sprintf(":%d", c.Port)
}
