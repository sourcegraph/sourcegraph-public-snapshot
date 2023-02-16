package util

import (
	"context"
	"strings"
)

// GetGitVersion returns the version of git installed on the host.
func GetGitVersion(runner CmdRunner, ctx context.Context) (string, error) {
	out, err := execOutput(runner, ctx, "git", "version")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(out, "git version "), nil
}

// GetSrcVersion returns the version of src installed on the host.
func GetSrcVersion(runner CmdRunner, ctx context.Context) (string, error) {
	out, err := execOutput(runner, ctx, "src", "version", "-client-only")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(out, "Current version: "), nil
}

// GetDockerVersion returns the version of docker installed on the host.
func GetDockerVersion(runner CmdRunner, ctx context.Context) (string, error) {
	return execOutput(runner, ctx, "docker", "version", "-f", "{{.Server.Version}}")
}

// GetIgniteVersion returns the version of ignite installed on the host.
func GetIgniteVersion(runner CmdRunner, ctx context.Context) (string, error) {
	return execOutput(runner, ctx, "ignite", "version", "-o", "short")
}
