package service

import (
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type Config struct {
	PubSub struct {
		ProjectID string
		TopicID   string
	}
}

func (c *Config) Load(env *runtime.Env) {
	c.PubSub.ProjectID = env.Get("PINGS_PUBSUB_PROJECT_ID", "", "The project ID for the Pub/Sub.")
	c.PubSub.TopicID = env.Get("PINGS_PUBSUB_TOPIC_ID", "", "The topic ID for the Pub/Sub.")
}
