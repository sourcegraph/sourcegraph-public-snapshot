package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type Config struct {
	Port              int
	DiagnosticsSecret string

	Events struct {
		PubSub struct {
			Enabled   bool
			ProjectID *string
			TopicID   *string
		}
	}
}

func (c *Config) Load(env *runtime.Env) {
	c.Port = env.GetInt("PORT", "10080", "Port to serve Telemetry Gateway service on, generally injected by Cloud Run.")
	c.DiagnosticsSecret = env.Get("DIAGNOSTICS_SECRET", "", "Secret for accessing diagnostics - "+
		"should be used as 'Authorization: Bearer $secret' header when accessing diagnostics endpoints.")

	c.Events.PubSub.Enabled = env.GetBool("TELEMETRY_GATEWAY_EVENTS_PUBSUB_ENABLED", "true",
		"If false, logs Pub/Sub messages instead of actually sending them")
	c.Events.PubSub.ProjectID = env.GetOptional("TELEMETRY_GATEWAY_EVENTS_PUBSUB_PROJECT_ID",
		"The project ID for the Pub/Sub.")
	c.Events.PubSub.TopicID = env.GetOptional("TELEMETRY_GATEWAY_EVENTS_PUBSUB_TOPIC_ID",
		"The topic ID for the Pub/Sub.")
}

func (c *Config) GetListenAdress() string {
	return fmt.Sprintf(":%d", c.Port)
}
