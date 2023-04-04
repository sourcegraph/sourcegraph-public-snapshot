package weaviate

import (
	"net"
	"os/exec"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const port = "8181"

func start(observationCtx *observation.Context, cgf *Config) error {
	logger := observationCtx.Logger

	// Initialize tracing/metrics
	observationCtx = observation.NewContext(logger, observation.Honeycomb(&honey.Dataset{
		Name:       "weaviate",
		SampleRate: 20,
	}))

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	logger.Info("listening", log.String("addr", addr))

	return exec.Command(cgf.Path).Start()
}
