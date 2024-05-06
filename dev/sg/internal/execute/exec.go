package execute

import (
	"bytes"
	"context"
	"os"
	"os/exec"
)

func Git(ctx context.Context, args ...string) ([]byte, error) {
	cmd := GitCmd(ctx, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}

func GitCmd(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "git", args...)
}

func GHCmd(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "gh", args...)
}

func GH(ctx context.Context, args ...string) ([]byte, error) {
	cmd := GHCmd(ctx, args...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}
