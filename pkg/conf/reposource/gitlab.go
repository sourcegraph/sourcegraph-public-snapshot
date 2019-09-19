package reposource

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
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

func GitLabRepoName(repositoryPathPattern, host, pathWithNamespace string, rps RegexpReplacements) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{pathWithNamespace}"
	}

	name := strings.NewReplacer(
		"{host}", host,
		"{pathWithNamespace}", pathWithNamespace,
	).Replace(repositoryPathPattern)

	return api.RepoName(rps.Replace(name))
}

// CompileGitLabRegexReplacements compiles a list of GitLabRegexReplacement into common regexpReplacement,
// it halts and returns when any regex compile error occurred.
func CompileGitLabRegexReplacements(repls []*schema.GitLabRegexReplacement) (RegexpReplacements, error) {
	rps := make([]*regexpReplacement, len(repls))
	for i, rr := range repls {
		r, err := regexp.Compile(rr.Regex)
		if err != nil {
			return nil, errors.Errorf("regexp.Compile %q: %v", rr.Regex, err)
		}
		rps[i] = &regexpReplacement{
			regexp:      r,
			replacement: rr.Replacement,
		}
	}
	return rps, nil
}
