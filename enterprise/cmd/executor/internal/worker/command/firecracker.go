package command

import (
	"fmt"
	"strings"

	"github.com/kballard/go-shellquote"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/util"
)

const (
	FirecrackerContainerDir  = "/work"
	FirecrackerDockerConfDir = "/etc/docker/cli"
)

func NewFirecrackerCommand(logger Logger, cmdRunner util.CmdRunner, vmName string, options DockerOptions) Command {
	dockerCommand := NewDockerCommand(logger, cmdRunner, FirecrackerContainerDir, options)
	innerCommand := shellquote.Join(dockerCommand.Command...)

	if options.Image == "" && len(dockerCommand.Env) > 0 {
		innerCommand = fmt.Sprintf("%s %s", strings.Join(quoteEnv(dockerCommand.Env), " "), innerCommand)
	}
	if dockerCommand.Dir != "" {
		innerCommand = fmt.Sprintf("cd %s && %s", shellquote.Join(dockerCommand.Dir), innerCommand)
	}
	dockerCommand.Command = []string{"ignite", "exec", vmName, "--", "sh", "-c", innerCommand}
	return dockerCommand
}
