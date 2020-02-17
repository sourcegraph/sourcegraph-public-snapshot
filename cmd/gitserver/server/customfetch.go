package server

import (
	"context"
	"net/url"
	"os/exec"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"gopkg.in/inconshreveable/log15.v2"
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
		cgm[mapping.DomainPath] = strings.Fields(mapping.Fetch)
	}

	return cgm
}

func extractDomainPath(cloneURL string) (string, error) {
	gitURL, err := url.Parse(cloneURL)
	if err != nil {
		return "", err
	}

	return path.Join(gitURL.Host, gitURL.Path), nil
}

func customFetchCmd(ctx context.Context, urlVal string) *exec.Cmd {
	cgm := customGitFetch().(map[string][]string)
	if len(cgm) == 0 {
		return nil
	}

	dp, err := extractDomainPath(urlVal)
	if err != nil {
		log15.Error("failed to extract domain and path", "url", urlVal, "err", err)
		return nil
	}
	cmdParts := cgm[dp]
	if len(cmdParts) == 0 {
		return nil
	}
	return exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
}
