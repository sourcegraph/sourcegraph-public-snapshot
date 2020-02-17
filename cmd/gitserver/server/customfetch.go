package server

import (
	"context"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var customFetch = strings.Fields(env.Get("SRC_GITSERVER_CUSTOM_FETCH", "",
	"EXPERIMENTAL: custom fetch command for unorthodox repo setups."))

func useCustomFetch() bool {
	return len(customFetch) > 0
}

func customFetchCmd(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, customFetch[0], customFetch[1:]...)
}
