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

// CompileGitLabNameTransformations compiles a list of GitLabNameTransformation into common nameTransformation,
// it halts and returns when any compile error occurred.
func CompileGitLabNameTransformations(ts []*schema.GitLabNameTransformation) (NameTransformations, error) {
	nts := make([]nameTransformation, len(ts))
	for i, t := range ts {
		switch {
		case t.Regex != "":
			r, err := regexp.Compile(t.Regex)
			if err != nil {
				return nil, errors.Errorf("regexp.Compile %q: %v", t.Regex, err)
			}
			nts[i] = nameTransformation{
				regexp:      r,
				replacement: t.Replacement,
			}

		default:
			return nil, errors.Errorf("unrecognized transformation: %v", t)
		}
	}
	return nts, nil
}
