package service

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type Config struct {
	Events struct {
		PubSub struct {
			Enabled   bool
			ProjectID *string
			TopicID   *string
		}

		StreamPublishConcurrency int
	}
}

func (c *Config) Load(env *runtime.Env) {
	c.Events.PubSub.Enabled = env.GetBool("TELEMETRY_GATEWAY_EVENTS_PUBSUB_ENABLED", "true",
		"If false, logs Pub/Sub messages instead of actually sending them")
	c.Events.PubSub.ProjectID = env.GetOptional("TELEMETRY_GATEWAY_EVENTS_PUBSUB_PROJECT_ID",
		"The project ID for the Pub/Sub.")
	c.Events.PubSub.TopicID = env.GetOptional("TELEMETRY_GATEWAY_EVENTS_PUBSUB_TOPIC_ID",
		"The topic ID for the Pub/Sub.")
	c.Events.StreamPublishConcurrency = env.GetInt("TELEMETRY_GATEWAY_EVENTS_STREAM_PUBLISH_CONCURRENCY", "250",
		"Per-stream concurrent publishing limit.")
}
