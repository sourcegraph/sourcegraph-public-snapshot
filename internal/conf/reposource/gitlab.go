package reposource

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
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

	nts, err := CompileGitLabNameTransformations(c.NameTransformations)
	if err != nil {
		return "", err
	}

	return GitLabRepoName(c.RepositoryPathPattern, baseURL.Hostname(), pathWithNamespace, nts), nil
}

func GitLabRepoName(repositoryPathPattern, host, pathWithNamespace string, nts NameTransformations) api.RepoName {
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{pathWithNamespace}"
	}

	name := strings.NewReplacer(
		"{host}", host,
		"{pathWithNamespace}", pathWithNamespace,
	).Replace(repositoryPathPattern)

	return api.RepoName(nts.Transform(name))
}

// CompileGitLabNameTransformations compiles a list of GitLabNameTransformation into common NameTransformation,
// it halts and returns when any compile error occurred.
func CompileGitLabNameTransformations(ts []*schema.GitLabNameTransformation) (NameTransformations, error) {
	nts := make([]NameTransformation, len(ts))
	for i, t := range ts {
		nt, err := NewNameTransformation(NameTransformationOptions{
			Regex:       t.Regex,
			Replacement: t.Replacement,
		})
		if err != nil {
			return nil, err
		}
		nts[i] = nt
	}
	return nts, nil
}
