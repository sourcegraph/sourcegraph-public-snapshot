// Package sysreq implements checking for Sourcegraph system requirements.
package sysreq

import (
	"crypto/tls"
	"os"
	"os/exec"
	"strings"
	"sync"

	droneexec "github.com/drone/drone-exec/exec"
	"github.com/samalba/dockerclient"
	"golang.org/x/net/context"
)

// Status describes the status of a system requirement.
type Status struct {
	Name    string // the required component
	Problem string // if non-empty, a description of the problem
	Fix     string // if non-empty, how to fix the problem
	Err     error  // if non-nil, the error encountered
	Skipped bool   // if true, indicates this check was skipped
}

// OK is whether the component is present, has no errors, and was not
// skipped.
func (s *Status) OK() bool {
	return s.Problem == "" && s.Fix == "" && s.Err == nil && !s.Skipped
}

func (s *Status) Failed() bool { return s.Problem != "" || s.Err != nil }

// Check checks for the presence of system requirements, such as
// Docker and Git. The skip list contains case-insensitive names of
// requirement checks (such as "Docker" and "Git") that should be
// skipped.
func Check(ctx context.Context, skip []string) []Status {
	shouldSkip := func(name string) bool {
		for _, v := range skip {
			if strings.EqualFold(name, v) {
				return true
			}
		}
		return false
	}

	statuses := make([]Status, len(checks))
	var wg sync.WaitGroup
	for i, c := range checks {
		statuses[i].Name = c.name

		if shouldSkip(c.name) {
			statuses[i].Skipped = true
			continue
		}

		wg.Add(1)
		go func(i int, c check) {
			defer wg.Done()

			finished := make(chan struct{})

			go func() {
				st, err := c.check(ctx)
				if err != nil {
					statuses[i].Err = err
				}
				if st != nil {
					statuses[i].Problem = st.problem
					statuses[i].Fix = st.fix
				}
				finished <- struct{}{}
			}()

			select {
			case <-finished:
			case <-ctx.Done():
				statuses[i].Err = context.DeadlineExceeded
			}
		}(i, c)
	}
	wg.Wait()

	return statuses
}

type status struct{ problem, fix string }

type check struct {
	name  string
	check func(context.Context) (*status, error)
}

var checks = []check{
	{
		name: "Docker",
		check: func(ctx context.Context) (*status, error) {
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
				tlsConfig, err = droneexec.TLSConfigFromCertPath(path)
				if err != nil {
					return nil, err
				}
			}

			st := &status{
				problem: "Could not contact Docker host. Docker is required for Sourcegraph to build and analyze code.",
				fix:     "Install Docker if you haven't already (see https://docs.docker.com/engine/installation/). Then check the DOCKER_HOST (and possibly DOCKER_CERT_PATH and DOCKER_TLS_VERIFY) environment variables. Set them so they point to a Docker host. If you're running on OS X, pass Sourcegraph the environment vars in `docker-machine env default`. See https://docs.docker.com/machine/reference/env/ for more information.",
			}

			client, err := dockerclient.NewDockerClient(daemonURL, tlsConfig)
			if err != nil {
				return st, err
			}

			if _, err := client.Version(); err != nil {
				return st, err
			}

			return nil, nil
		},
	},
	{
		name: "Git",
		check: func(ctx context.Context) (*status, error) {
			if _, err := exec.LookPath("git"); err != nil {
				return &status{
					problem: "Git is not installed",
					fix:     "Install Git on your system and make sure it is in your $PATH.",
				}, err
			}
			return nil, nil
		},
	},
}
