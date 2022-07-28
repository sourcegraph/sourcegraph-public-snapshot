package server

import (
	"context"
	"os/exec"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

var insecureEnableGitFetch = env.Get("INSECURE_ENABLE_CUSTOM_GIT_FETCH", "false", "enable custom git fetch")
var customGitFetch = conf.Cached[map[string][]string](func() map[string][]string {
	exp := conf.ExperimentalFeatures()
	return buildCustomFetchMappings(exp.CustomGitFetch)
})

func buildCustomFetchMappings(c []*schema.CustomGitFetchMapping) map[string][]string {
	if c == nil || insecureEnableGitFetch == "false" {
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
