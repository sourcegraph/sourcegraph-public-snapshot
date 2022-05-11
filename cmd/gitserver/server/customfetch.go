package server

import (
	"context"
	"os/exec"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

var customGitFetch = conf.Cached(func() any {
	exp := conf.ExperimentalFeatures()
	return buildCustomFetchMappings(exp.CustomGitFetch)
})

func buildCustomFetchMappings(c []*schema.CustomGitFetchMapping) map[string][]string {
	if c == nil {
		return map[string][]string{}
	}

	cgm := map[string][]string{}
	for _, mapping := range c {
		cgm[mapping.DomainPath] = strings.Fields(mapping.Fetch)
	}

	return cgm
}

func customFetchCmd(ctx context.Context, remoteURL *vcs.URL) *exec.Cmd {
	cgm := customGitFetch().(map[string][]string)
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
