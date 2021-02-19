package reposource

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
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

const DefaultRepositoryPathPattern = "{base}/{repo}"

func (c Other) CloneURLToRepoURI(cloneURL string) (string, error) {
	return cloneURLToRepoName(cloneURL, c.Url, DefaultRepositoryPathPattern)
}

func (c Other) CloneURLToRepoName(cloneURL string) (api.RepoName, error) {
	repoName, err := cloneURLToRepoName(cloneURL, c.Url, c.RepositoryPathPattern)
	return api.RepoName(repoName), err
}

func cloneURLToRepoName(cloneURL, baseURL, repositoryPathPattern string) (string, error) {
	parsedCloneURL, parsedBaseURL, match, err := parseURLs(cloneURL, baseURL)
	if err != nil {
		return "", err
	}
	if !match {
		return "", nil
	}

	basePrefix := parsedBaseURL.Path
	if !strings.HasSuffix(basePrefix, "/") {
		basePrefix = basePrefix + "/"
	}
	if parsedCloneURL.Path != parsedBaseURL.Path && !strings.HasPrefix(parsedCloneURL.Path, basePrefix) {
		return "", urlMismatchErr{cloneURL: cloneURL, hostURL: baseURL}
	}
	relativeRepoPath := strings.TrimPrefix(parsedCloneURL.Path, basePrefix)
	base := url.URL{
		Host: parsedBaseURL.Host,
		Path: parsedBaseURL.Path,
	}
	return OtherRepoName(repositoryPathPattern, base.String(), relativeRepoPath), nil
}

var otherRepoNameReplacer = strings.NewReplacer(":", "-", "@", "-", "//", "")

func OtherRepoName(repositoryPathPattern, base, relativeRepoPath string) string {
	if repositoryPathPattern == "" {
		repositoryPathPattern = DefaultRepositoryPathPattern
	}
	return strings.NewReplacer(
		"{base}", otherRepoNameReplacer.Replace(strings.TrimSuffix(base, "/")),
		"{repo}", otherRepoNameReplacer.Replace(strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(relativeRepoPath, "/"), ".git"), "/")),
	).Replace(repositoryPathPattern)
}
