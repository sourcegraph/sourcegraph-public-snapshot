package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseConfig(t *testing.T) {
	input := `
env:
  SRC_REPOS_DIR: $HOME/.sourcegraph/repos

commands:
  frontend:
    cmd: ulimit -n 10000 && .bin/frontend
    install: go build -o .bin/frontend github.com/sourcegraph/sourcegraph/cmd/frontend
    checkBinary: .bin/frontend
    env:
      CONFIGURATION_MODE: server
    watch:
      - lib

checks:
  docker:
    cmd: docker version
    failMessage: "Failed to run 'docker version'. Please make sure Docker is running."

commandsets:
  oss:
    - frontend
    - gitserver
  enterprise:
    checks:
      - docker
    commands:
      - frontend
      - gitserver
`

	have, err := ParseConfig([]byte(input))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	want := &Config{
		Env: map[string]string{"SRC_REPOS_DIR": "$HOME/.sourcegraph/repos"},
		Commands: map[string]Command{
			"frontend": {
				Name:        "frontend",
				Cmd:         "ulimit -n 10000 && .bin/frontend",
				Install:     "go build -o .bin/frontend github.com/sourcegraph/sourcegraph/cmd/frontend",
				CheckBinary: ".bin/frontend",
				Env:         map[string]string{"CONFIGURATION_MODE": "server"},
				Watch:       []string{"lib"},
			},
		},
		Commandsets: map[string]*Commandset{
			"oss": {
				Name:     "oss",
				Commands: []string{"frontend", "gitserver"},
			},
			"enterprise": {
				Name:     "enterprise",
				Commands: []string{"frontend", "gitserver"},
				Checks:   []string{"docker"},
			},
		},
		Checks: map[string]Check{
			"docker": {
				Name:        "docker",
				Cmd:         "docker version",
				FailMessage: "Failed to run 'docker version'. Please make sure Docker is running.",
			},
		},
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("wrong config. (-want +got):\n%s", diff)
	}
}
