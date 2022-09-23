package apiclient

import (
	"context"
	"os/exec"
	"runtime"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/version"
)

type TelemetryOptions struct {
	OS              string
	Architecture    string
	DockerVersion   string
	ExecutorVersion string
	GitVersion      string
	IgniteVersion   string
	SrcCliVersion   string
}

func NewTelemetryOptions(ctx context.Context, useFirecracker bool) TelemetryOptions {
	t := TelemetryOptions{
		OS:              runtime.GOOS,
		Architecture:    runtime.GOARCH,
		ExecutorVersion: version.Version(),
	}

	var err error

	t.GitVersion, err = getGitVersion(ctx)
	if err != nil {
		log15.Error("Failed to get git version", "err", err)
	}

	t.SrcCliVersion, err = getSrcVersion(ctx)
	if err != nil {
		log15.Error("Failed to get src-cli version", "err", err)
	}

	t.DockerVersion, err = getDockerVersion(ctx)
	if err != nil {
		log15.Error("Failed to get docker version", "err", err)
	}

	if useFirecracker {
		t.IgniteVersion, err = getIgniteVersion(ctx)
		if err != nil {
			log15.Error("Failed to get ignite version", "err", err)
		}
	}

	return t
}

func getGitVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "version")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "git version "), nil
}

func getSrcVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "src", "version", "-client-only")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "Current version: "), nil
}

func getDockerVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "version", "-f", "{{.Server.Version}}")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getIgniteVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "ignite", "version", "-o", "short")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
