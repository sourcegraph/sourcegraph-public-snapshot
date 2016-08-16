package worker

import (
	"crypto/tls"
	"os"

	"context"

	"github.com/samalba/dockerclient"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/sysreq"
)

// AddSysReqCheck populates sysreq with the checks needed for the worker to
// function
func AddSysReqCheck() {
	sysreq.AddCheck("Docker", func(ctx context.Context) (problem, fix string, err error) {
		// TODO(sqs!native-ci): copied temporarily from
		// https://github.com/drone/drone-exec/pull/13, godep
		// update when that is merged into drone-exec
		daemonURL := os.Getenv("DOCKER_HOST")
		if daemonURL == "" {
			daemonURL = "unix:///var/run/docker.sock"
		}
		var tlsConfig *tls.Config
		if path := os.Getenv("DOCKER_CERT_PATH"); os.Getenv("DOCKER_TLS_VERIFY") != "" && path != "" {
			var err error
			tlsConfig, err = dockerclient.TLSConfigFromCertPath(path)
			if err != nil {
				return "", "", err
			}
		}

		problem = "Could not contact Docker host. Docker is required for Sourcegraph to build and analyze code."
		fix = "Install Docker if you haven't already (see https://docs.docker.com/engine/installation/). Then check the DOCKER_HOST (and possibly DOCKER_CERT_PATH and DOCKER_TLS_VERIFY) environment variables. Set them so they point to a Docker host. If you're running on OS X, pass Sourcegraph the environment vars in `docker-machine env $(docker-machine ls -q)`. See https://docs.docker.com/machine/reference/env/ for more information."

		client, err := dockerclient.NewDockerClient(daemonURL, tlsConfig)
		if err != nil {
			return problem, fix, err
		}

		if _, err := client.Version(); err != nil {
			return problem, fix, err
		}

		return "", "", nil
	})
}
