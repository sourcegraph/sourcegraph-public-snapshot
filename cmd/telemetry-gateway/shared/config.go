package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

type Config struct {
	env.BaseConfig

	Port              int
	DiagnosticsSecret string

	DebugServer string

	Events struct {
		PubSub struct {
			ProjectID string
			TopicID   string
		}
	}

	AllowAnonymous bool

	OpenTelemetry OpenTelemetryConfig
}

type OpenTelemetryConfig struct {
	TracePolicy  policy.TracePolicy
	GCPProjectID string
}

func (c *Config) Load() {
	c.Port = c.GetInt("PORT", "10086", "Port to serve Telemetry Gateway service on, generally injected by Cloud Run.")
	c.DiagnosticsSecret = c.Get("DIAGNOSTICS_SECRET", "", "Secret for accessing diagnostics - "+
		"should be used as 'Authorization: Bearer $secret' header when accessing diagnostics endpoints.")

	c.Events.PubSub.ProjectID = c.Get("TELEMETRY_GATEWAY_EVENTS_PUBSUB_PROJECT_ID", "", "The project ID for the Pub/Sub.")
	c.Events.PubSub.TopicID = c.Get("TELEMETRY_GATEWAY_EVENTS_PUBSUB_TOPIC_ID", "", "The topic ID for the Pub/Sub.")
}

func (c *Config) GetListenAdress() string {
	return fmt.Sprintf(":%d", c.Port)
}
