package shared

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Address   string
	GRPCWebUI bool
}

func (c *Config) Load() {
	c.Address = c.Get("LLM_PROXY_ADDR", ":9992", "Address to serve LLM proxy on.")
	c.GRPCWebUI = c.GetBool("LLM_PROXY_GRPC_WEB_UI", "false", "Serve gRPC Web UI for the LLM proxy service.")
}
