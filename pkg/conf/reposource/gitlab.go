package reposource

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GitLab struct {
	*schema.GitLabConnection
}

var _ RepoSource = GitLab{}

func (c GitLab) CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	parsedCloneURL, baseURL, match, err := parseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !match {
		return "", nil
	}

	pathWithNamespace := strings.TrimPrefix(strings.TrimSuffix(parsedCloneURL.Path, ".git"), "/")

	rps, err := CompileGitLabRegexReplacements(c.ReplaceAllInRepositoryName)
	if err != nil {
		return "", err
	}

	return GitLabRepoName(c.RepositoryPathPattern, baseURL.Hostname(), pathWithNamespace, rps), nil
}

func GitLabRepoName(repositoryPathPattern, host, pathWithNamespace string, rps []*RegexpReplacement) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{pathWithNamespace}"
	}

	name := strings.NewReplacer(
		"{host}", host,
		"{pathWithNamespace}", pathWithNamespace,
	).Replace(repositoryPathPattern)

	for _, rp := range rps {
		name = rp.Regexp.ReplaceAllString(name, rp.Replacement)
	}

	return api.RepoName(name)
}

// CompileGitLabRegexReplacements compiles a list of GitLabRegexReplacement into common RegexpReplacement,
// it halts and returns when any regex compile error occurred.
func CompileGitLabRegexReplacements(repls []*schema.GitLabRegexReplacement) ([]*RegexpReplacement, error) {
	rps := make([]*RegexpReplacement, len(repls))
	for i, rr := range repls {
		r, err := regexp.Compile(rr.Regex)
		if err != nil {
			return nil, fmt.Errorf("compile %q: %v", rr.Regex, err)
		}
		rps[i] = &RegexpReplacement{
			Regexp:      r,
			Replacement: rr.Replacement,
		}
	}
	return rps, nil
}
