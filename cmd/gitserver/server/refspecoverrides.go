package server

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/env"
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
//
// To not clone everything we instead init a bare repo and only add the
// refspecs we care about. Then we finally do a fetch.
func refspecOverridesCloneCmd(ctx context.Context, remoteURL *url.URL, tmpPath string) (*exec.Cmd, error) {
	if err := os.MkdirAll(tmpPath, os.ModePerm); err != nil {
		return nil, errors.Wrapf(err, "clone failed to create tmp dir")
	}
	cmds := [][]string{
		{"init", "--bare", "."},
		{"config", "--add", "remote.origin.url", remoteURL.String()},
		{"config", "--add", "remote.origin.mirror", "true"},
	}
	for _, refspec := range refspecOverrides {
		cmds = append(cmds, []string{"config", "--add", "remote.origin.fetch", refspec})
	}
	for _, args := range cmds {
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = tmpPath
		if err := cmd.Run(); err != nil {
			return nil, errors.Wrapf(err, "clone setup failed")
		}
	}
	cmd := exec.CommandContext(ctx, "git", "fetch", "--progress")
	cmd.Dir = tmpPath
	return cmd, nil
}

// HACK(keegancsmith) workaround to experiment with cloning less in a large
// monorepo. https://github.com/sourcegraph/customer/issues/19
func refspecOverridesFetchCmd(ctx context.Context, remoteURL *url.URL) *exec.Cmd {
	return exec.CommandContext(ctx, "git", append([]string{"fetch", "--prune", remoteURL.String()}, refspecOverrides...)...)
}
