package graphqlbackend

import "strings"

// repositoryRevision specifies a repository and 0 or more revspecs.
// If no revspec is specified, then the repository's default branch is used.
type repositoryRevision struct {
	repo     string
	revspecs []string
}

// parseRepositoryRevision parses strings of the form "repo" or "repo@rev" into
// a repositoryRevision.
func parseRepositoryRevision(repoAndOptionalRev string) repositoryRevision {
	i := strings.Index(repoAndOptionalRev, "@")
	if i == -1 {
		return repositoryRevision{repo: repoAndOptionalRev}
	}
	rev := repoAndOptionalRev[i+1:]
	return repositoryRevision{
		repo:     repoAndOptionalRev[:i],
		revspecs: []string{rev},
	}
}

func (repoRev *repositoryRevision) String() string {
	if repoRev.hasSingleRevSpec() {
		return repoRev.repo + "@" + repoRev.revspecs[0]
	}
	return repoRev.repo
}

func (repoRev *repositoryRevision) revSpecsOrDefaultBranch() []string {
	if len(repoRev.revspecs) == 0 {
		return []string{""}
	}
	return repoRev.revspecs
}

func (repoRev *repositoryRevision) hasSingleRevSpec() bool {
	return len(repoRev.revspecs) == 1
}
