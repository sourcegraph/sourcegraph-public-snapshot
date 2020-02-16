package server

import (
	"context"
	"net/url"
	"os/exec"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

var customGitFetch = conf.Cached(func() interface{} {
	return buildCustomFetchMappings(conf.Get().ExperimentalFeatures.CustomGitFetch)
})

func buildCustomFetchMappings(c []*schema.CustomGitFetchMapping) map[string][]string {
	if c == nil {
		return map[string][]string{}
	}

	cgm := map[string][]string{}

	for _, mapping := range c {
		parts := strings.Fields(mapping.Fetch)

		// TODO(uwedeportivo): can we enforce in the schema that fetch command is not empty ? otherwise log ?
		if len(parts) > 0 {
			cgm[mapping.DomainPath] = parts
		}
	}

	return cgm
}

func domainPath(urlVal string) (string, error) {
	gitUrl, err := url.Parse(urlVal)
	if err != nil {
		return "", err
	}

	return path.Join(gitUrl.Host, gitUrl.Path), nil
}

func customFetchCmd(ctx context.Context, urlVal string) *exec.Cmd {
	cgm := customGitFetch().(map[string][]string)

	dp, err := domainPath(urlVal)
	if err != nil {
		// TODO(uwedeportivo): log here ?
		return nil
	}

	cmdParts := cgm[dp]

	if len(cmdParts) == 0 {
		return nil
	}
	return exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
}
