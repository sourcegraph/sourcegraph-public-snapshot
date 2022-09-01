package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

var customGitFetchCmdConf = env.Get("CUSTOM_GIT_FETCH_CONF", "", "custom git fetch command configuration")

func customGitFetch() map[string][]string {
	if customGitFetchCmdConf == "" {
		return map[string][]string{}
	}

	r, err := ioutil.ReadFile(customGitFetchCmdConf)
	if err != nil {
		return map[string][]string{}
	}

	var cc []*schema.CustomGitFetchMapping
	err = json.Unmarshal(r, &cc)
	if err != nil {
		return map[string][]string{}
	}

	cgm := map[string][]string{}
	for _, mapping := range cc {
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
