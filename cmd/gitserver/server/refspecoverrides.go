package server

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

// HACK(keegancsmith) workaround to experiment with cloning less in a large
// monorepo. https://github.com/sourcegraph/customer/issues/19
var refspecOverrides = strings.Fields(env.Get("SRC_GITSERVER_REFSPECS", "", "EXPERIMENTAL: override refspec we fetch. Space separated."))

// HACK(keegancsmith) workaround to experiment with cloning less in a large
// monorepo. https://github.com/sourcegraph/customer/issues/19
func useRefspecOverrides() bool {
	return len(refspecOverrides) > 0
}

// HACK(keegancsmith) workaround to experiment with cloning less in a large
// monorepo. https://github.com/sourcegraph/customer/issues/19
func refspecOverridesFetchCmd(ctx context.Context, remoteURL *vcs.URL) *exec.Cmd {
	// Perform automatic repository maintenance at the end of fetch
	gc := "--auto-gc"
	if e := os.Getenv("SRC_ENABLE_GC_AUTO"); e == "false" {
		gc = "--no-auto-gc"
	}
	return exec.CommandContext(ctx, "git", append([]string{"fetch", gc, "--progress", "--prune", remoteURL.String()}, refspecOverrides...)...)
}
