package util

import (
	"context"
	"strings"
)

// GetGitVersion returns the version of git installed on the host.
func GetGitVersion(ctx context.Context, runner CmdRunner) (string, error) {
	out, err := execOutput(ctx, runner, "git", "version")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(out, "git version "), nil
}

// GetSrcVersion returns the version of src installed on the host.
func GetSrcVersion(ctx context.Context, runner CmdRunner) (string, error) {
	out, err := execOutput(ctx, runner, "src", "version", "-client-only")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(out, "Current version: "), nil
}

// GetDockerVersion returns the version of docker installed on the host.
func GetDockerVersion(ctx context.Context, runner CmdRunner) (string, error) {
	return execOutput(ctx, runner, "docker", "version", "-f", "{{.Server.Version}}")
}

// GetIgniteVersion returns the version of ignite installed on the host.
func GetIgniteVersion(ctx context.Context, runner CmdRunner) (string, error) {
	return execOutput(ctx, runner, "ignite", "version", "-o", "short")
}
