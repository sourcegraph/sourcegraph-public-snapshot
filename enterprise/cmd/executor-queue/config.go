package main

import (
	apiserver "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor-queue/internal/server"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Port int
}

func (c *Config) Load() {
	c.Port = c.GetInt("EXECUTOR_QUEUE_API_PORT", "3191", "The port to listen on.")
}

func (c *Config) ServerOptions() apiserver.ServerOptions {
	return apiserver.ServerOptions{
		Port: c.Port,
	}
}
