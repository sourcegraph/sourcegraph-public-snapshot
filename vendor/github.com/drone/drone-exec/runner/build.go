package runner

import (
	"errors"
	"fmt"

	"context"

	// log "github.com/Sirupsen/logrus"
	"github.com/drone/drone-exec/docker"
	"github.com/drone/drone-exec/parser"
	"github.com/drone/drone-exec/runner/script"
	"github.com/samalba/dockerclient"
)

var ErrNoImage = errors.New("Yaml must specify an image for every step")

// Default clone plugin.
var DefaultCloner = "plugins/drone-git"

// Default cache plugin.
var DefaultCacher = "plugins/drone-cache"

type Build struct {
	tree *parser.Tree
	typ  parser.NodeType
}

func (b *Build) Run(ctx context.Context, state *State) error {
	return b.RunNode(ctx, state, "")
}

func (b *Build) RunNode(ctx context.Context, state *State, typ parser.NodeType) error {
	b.typ = typ
	return b.walk(ctx, b.tree.Root, "", state)
}

func (b *Build) walk(ctx context.Context, node parser.Node, key string, state *State) (err error) {
	switch node := node.(type) {
	case *parser.ListNode:
		for i, n := range node.Nodes {
			if node.Nodes != nil {
				key = node.Keys[i]
			}

			err = b.walk(ctx, n, key, state)
			if err != nil {
				break
			}
		}

	case *parser.FilterNode:
		if isMatch(node, state) {
			if err := b.walk(ctx, node.Node, key, state); err != nil {
				return err
			}
		}

	case *parser.DockerNode:
		if b.typ != node.NodeType {
			break
		}
		if len(node.Image) == 0 {
			break
		}
		// auth for accessing private docker registries
		var auth *dockerclient.AuthConfig
		// auth to nil if password or token not set
		if len(node.AuthConfig.Password) != 0 || len(node.AuthConfig.RegistryToken) != 0 {
			auth = &dockerclient.AuthConfig{
				Username:      node.AuthConfig.Username,
				Password:      node.AuthConfig.Password,
				Email:         node.AuthConfig.Email,
				RegistryToken: node.AuthConfig.RegistryToken,
			}
		}

		// Set up monitor.
		mon := state.Monitor(string(node.Type()), key, node)
		outw, errw := mon.Logger()

		// Record the final state.
		var allowableFailure bool
		defer func() {
			failed := state.Failed()
			if allowableFailure {
				// Not recorded in state because state can't
				// transition from failure back to non-failure.
				failed = true
			}
			mon.End(!failed, node.AllowFailure)

			// Prevent future steps from being run if this was
			// canceled.
			if err == context.Canceled || err == context.DeadlineExceeded {
				state.Exit(1)
			}
		}()

		switch node.Type() {

		case parser.NodeBuild:
			// TODO(bradrydzewski) this should be handled by the when block
			// by defaulting the build steps to run when not failure. This is
			// required now that we support multi-build steps.
			if state.Failed() {
				mon.Skip()
				return
			}

			mon.Start()

			conf := toContainerConfig(node)
			conf.Env = append(conf.Env, toEnv(state)...)
			conf.WorkingDir = state.Workspace.Path
			if state.Repo.IsPrivate {
				script.Encode(state.Workspace, conf, node)
			} else {
				script.Encode(nil, conf, node)
			}

			recordExitCode := func(code int) {
				if code != 0 && node.AllowFailure {
					allowableFailure = true
				} else {
					state.Exit(code)
				}
			}

			info, err := docker.Run(ctx, state.Client, conf, auth, node.Pull, outw, errw, errw)
			if err != nil {
				recordExitCode(255)
				fmt.Fprintln(errw, err)
			} else if info.State.ExitCode != 0 {
				recordExitCode(info.State.ExitCode)
			}

		case parser.NodeCompose:
			mon.Start()

			conf := toContainerConfig(node)
			_, err := docker.Start(state.Client, conf, auth, node.Pull, errw)
			if err != nil {
				fmt.Fprintln(errw, err)
				state.Exit(255)
			}

		default:
			mon.Start()

			conf := toContainerConfig(node)
			conf.Cmd = toCommand(state, node)
			info, err := docker.Run(ctx, state.Client, conf, auth, node.Pull, outw, errw, errw)
			if err != nil {
				state.Exit(255)
				fmt.Fprintln(errw, err)
			} else if info.State.ExitCode != 0 {
				state.Exit(info.State.ExitCode)
			}
		}
	}

	return nil
}

func expectMatch() {

}

func maybeResolveImage() {}

func maybeEscalate(conf dockerclient.ContainerConfig, node *parser.DockerNode) {
	if node.Image == "plugins/drone-docker" || node.Image == "plugins/drone-gcr" {
		return
	}
	conf.Volumes = nil
	conf.HostConfig.NetworkMode = ""
	conf.HostConfig.Privileged = true
	conf.Entrypoint = []string{}
	conf.Cmd = []string{}
}

// shouldEscalate is a helper function that returns true
// if the plugin should be escalated to start the container
// in privileged mode.
func shouldEscalate(node *parser.DockerNode) bool {
	return node.Image == "plugins/drone-docker" ||
		node.Image == "plugins/drone-gcr"
}
