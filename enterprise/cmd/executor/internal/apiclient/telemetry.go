package apiclient

import (
	"context"
	"os/exec"
	"runtime"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/version"
)

type telemetryOptions struct {
	os            string
	arch          string
	version       string
	srcCliVersion string
	dockerVersion string
	igniteVersion string
	gitVersion    string
}

func newTelemetryOptions(ctx context.Context) telemetryOptions {
	t := telemetryOptions{
		os:      runtime.GOOS,
		arch:    runtime.GOARCH,
		version: version.Version(),
	}

	var err error

	t.gitVersion, err = getGitVersion(ctx)
	if err != nil {
		log15.Error("Failed to get git version", "err", err)
	}

	t.srcCliVersion, err = getSrcVersion(ctx)
	if err != nil {
		log15.Error("Failed to get src-cli version", "err", err)
	}

	t.dockerVersion, err = getDockerVersion(ctx)
	if err != nil {
		log15.Error("Failed to get docker version", "err", err)
	}

	t.igniteVersion, err = getIgniteVersion(ctx)
	if err != nil {
		log15.Error("Failed to get ignite version", "err", err)
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
