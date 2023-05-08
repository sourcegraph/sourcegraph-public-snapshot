package shared

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Address string

	Dotcom struct {
		AccessToken string
	}

	Anthropic struct {
		AccessToken string
	}

	AllowAnonymous bool

	SourcesSyncInterval time.Duration
}

func (c *Config) Load() {
	c.Address = c.Get("LLM_PROXY_ADDR", ":9992", "Address to serve LLM proxy on.")
	c.Dotcom.AccessToken = c.Get("LLM_PROXY_DOTCOM_ACCESS_TOKEN", "", "The Sourcegraph.com access token to be used.")
	c.Anthropic.AccessToken = c.Get("LLM_PROXY_ANTHROPIC_ACCESS_TOKEN", "", "The Anthropic access token to be used.")
	c.AllowAnonymous = c.GetBool("LLM_PROXY_ALLOW_ANONYMOUS", "false", "Allow anonymous access to LLM proxy.")
	c.SourcesSyncInterval = c.GetInterval("LLM_PROXY_SOURCES_SYNC_INTERVAL", "2m", "The interval at which to sync actor sources.")
}
