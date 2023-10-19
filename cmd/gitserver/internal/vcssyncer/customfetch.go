package vcssyncer

import (
	"context"
	"os/exec"
	"path"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

var customGitFetch = conf.Cached(func() map[string][]string {
	exp := conf.ExperimentalFeatures()
	return buildCustomFetchMappings(exp.CustomGitFetch)
})

var enableCustomGitFetch = env.Get("ENABLE_CUSTOM_GIT_FETCH", "false", "Enable custom git fetch")

func buildCustomFetchMappings(c []*schema.CustomGitFetchMapping) map[string][]string {
	// this is an edge case where a CustomGitFetchMapping has been made but enableCustomGitFetch is false
	if c != nil && enableCustomGitFetch == "false" {
		logger := log.Scoped("customfetch")
		logger.Warn("a CustomGitFetchMapping is configured but ENABLE_CUSTOM_GIT_FETCH is not set")

		return map[string][]string{}
	}
	if c == nil || enableCustomGitFetch == "false" {
		return map[string][]string{}
	}

	cgm := map[string][]string{}
	for _, mapping := range c {
		cgm[mapping.DomainPath] = strings.Fields(mapping.Fetch)
	}

	return cgm
}

func customFetchCmd(ctx context.Context, remoteURL *vcs.URL) *exec.Cmd {
	cgm := customGitFetch()
	if len(cgm) == 0 {
		return nil
	}

	dp := path.Join(remoteURL.Host, remoteURL.Path)
	cmdParts := cgm[dp]
	if len(cmdParts) == 0 {
		return nil
	}
	return exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
}
