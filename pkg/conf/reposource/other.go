package reposource

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type urlMismatchErr struct {
	cloneURL string
	hostURL  string
}

func (e urlMismatchErr) Error() string {
	return fmt.Sprintf("cloneURL %q did not match git host %q", e.cloneURL, e.hostURL)
}

type Other struct {
	*schema.OtherExternalServiceConnection
}

var _ RepoSource = Other{}

func (c Other) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, baseURL, match, err := parseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !match {
		return "", urlMismatchErr{cloneURL: cloneURL, hostURL: c.Url}
	}

	basePrefix := baseURL.Path
	if !strings.HasSuffix(basePrefix, "/") {
		basePrefix = basePrefix + "/"
	}
	if parsedCloneURL.Path != baseURL.Path && !strings.HasPrefix(parsedCloneURL.Path, basePrefix) {
		return "", urlMismatchErr{cloneURL: cloneURL, hostURL: c.Url}
	}
	relativeRepoPath := strings.TrimPrefix(parsedCloneURL.Path, basePrefix)

	base := url.URL{
		Host: baseURL.Host,
		Path: baseURL.Path,
	}
	return OtherRepoName(c.RepositoryPathPattern, base.String(), relativeRepoPath), nil
}

var otherRepoNameReplacer = strings.NewReplacer(":", "-", "@", "-", "//", "")

func OtherRepoName(repositoryPathPattern, base, relativeRepoPath string) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{base}/{repo}"
	}
	return api.RepoName(
		strings.NewReplacer(
			"{base}", otherRepoNameReplacer.Replace(strings.TrimSuffix(base, "/")),
			"{repo}", otherRepoNameReplacer.Replace(strings.TrimSuffix(strings.TrimPrefix(relativeRepoPath, "/"), ".git")),
		).Replace(repositoryPathPattern),
	)
}
