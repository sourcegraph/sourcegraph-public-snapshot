package config

import "github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"

const DefaultChannel = "#william-buildchecker-webhook-test"

type Config struct {
	BuildkiteWebhookToken string
	BuildkiteToken        string
	GithubToken           string
	SlackToken            string
	SlackChannel          string
	Production            bool
	DebugPassword         string
}

func (c *Config) Load(env *runtime.Env) {
	c.BuildkiteWebhookToken = env.Get("BUILDKITE_WEBHOOK_TOKEN", "", "")
	c.BuildkiteToken = env.Get("BUILDKITE_TOKEN", "", "")
	c.SlackToken = env.Get("SLACK_TOKEN", "", "")
	c.SlackChannel = env.Get("SLACK_CHANNEL", DefaultChannel, "")
	c.Production = env.GetBool("BUILDTRACKER_PRODUCTION", "false", "")
	c.GithubToken = env.Get("GITHUB_TOKEN", "", "")

	if c.Production {
		c.DebugPassword = env.Get("BUILDTRACKER_DEBUG_PASSWORD", "", "")
	}
}
