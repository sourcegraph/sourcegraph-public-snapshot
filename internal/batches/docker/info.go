package docker

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/sourcegraph/src-cli/internal/exec"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CurrentContext returns the name of the current Docker context (not to be
// confused with a Go context).
func CurrentContext(ctx context.Context) (string, error) {
	dctx, cancel, err := withFastCommandContext(ctx)
	if err != nil {
		return "", err
	}
	defer cancel()

	args := []string{"context", "inspect", "--format", "{{ .Name }}"}
	out, err := exec.CommandContext(dctx, "docker", args...).CombinedOutput()
	if errors.IsDeadlineExceeded(err) || errors.IsDeadlineExceeded(dctx.Err()) {
		return "", newFastCommandTimeoutError(dctx, args...)
	} else if err != nil {
		return "", err
	}

	name := string(bytes.TrimSpace(out))
	if name == "" {
		return "", errors.New("no context returned from Docker")
	}

	return name, nil
}

type Info struct {
	Host struct {
		CPUs int `json:"cpus"`
	} `json:"host"` // Podman engine
	NCPU int `json:"NCPU"` // Docker Engine
}

// NCPU returns the number of CPU cores available to Docker.
func NCPU(ctx context.Context) (int, error) {
	dctx, cancel, err := withFastCommandContext(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()

	args := []string{"info", "--format", "{{ json .}}"}
	out, err := exec.CommandContext(dctx, "docker", args...).CombinedOutput()
	if errors.IsDeadlineExceeded(err) || errors.IsDeadlineExceeded(dctx.Err()) {
		return 0, newFastCommandTimeoutError(dctx, args...)
	} else if err != nil {
		return 0, err
	}

	var info Info
	if err := json.Unmarshal(out, &info); err != nil {
		return 0, err
	}
	if info.NCPU > 0 {
		return info.NCPU, nil
	}
	return info.Host.CPUs, nil
}
