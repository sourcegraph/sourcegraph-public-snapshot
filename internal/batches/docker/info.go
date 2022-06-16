package docker

import (
	"bytes"
	"context"
	"strconv"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/exec"
)

// NCPU returns the number of CPU cores available to Docker.
func NCPU(ctx context.Context) (int, error) {
	dctx, cancel, err := withFastCommandContext(ctx)
	if err != nil {
		return 0, err
	}
	defer cancel()

	args := []string{"info", "--format", "{{ .NCPU }}"}
	out, err := exec.CommandContext(dctx, "docker", args...).CombinedOutput()
	if errors.IsDeadlineExceeded(err) || errors.IsDeadlineExceeded(dctx.Err()) {
		return 0, newFastCommandTimeoutError(dctx, args...)
	} else if err != nil {
		return 0, err
	}

	dcpu, err := strconv.Atoi(string(bytes.TrimSpace(out)))
	if err != nil {
		return 0, errors.Wrap(err, "parsing docker cpu count")
	}

	return dcpu, nil
}
