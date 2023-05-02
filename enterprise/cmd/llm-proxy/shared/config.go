package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Address string

	Anthropic struct {
		AccessToken string
	}
}

func (c *Config) Load() {
	c.Address = c.Get("LLM_PROXY_ADDR", ":9992", "Address to serve LLM proxy on.")
	c.Anthropic.AccessToken = c.Get("LLM_PROXY_ANTHROPIC_ACCESS_TOKEN", "", "The Anthropic access token to be used.")
}
