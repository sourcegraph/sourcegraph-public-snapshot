package graphqlbackend

import "strings"

// repositoryRevisions specifies a repository and 0 or more revspecs.
// If no revspec is specified, then the repository's default branch is used.
type repositoryRevisions struct {
	repo     string
	revspecs []string
}

// parseRepositoryRevisions parses strings of the form "repo" or "repo@rev" into
// a repositoryRevisions.
func parseRepositoryRevisions(repoAndOptionalRev string) repositoryRevisions {
	i := strings.Index(repoAndOptionalRev, "@")
	if i == -1 {
		return repositoryRevisions{repo: repoAndOptionalRev}
	}
	rev := repoAndOptionalRev[i+1:]
	return repositoryRevisions{
		repo:     repoAndOptionalRev[:i],
		revspecs: []string{rev},
	}
}

func (r *repositoryRevisions) String() string {
	if r.hasSingleRevSpec() {
		return r.repo + "@" + r.revspecs[0]
	}
	return r.repo
}

func (r *repositoryRevisions) revSpecsOrDefaultBranch() []string {
	if len(r.revspecs) == 0 {
		return []string{""}
	}
	return r.revspecs
}

func (r *repositoryRevisions) hasSingleRevSpec() bool {
	return len(r.revspecs) == 1
}
