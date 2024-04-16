package batches

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

// TODO(mrnugget): Merge these two types (give them an "errorfmt" function,
// rename "Has*" methods to "NotEmpty" or something)

// UnsupportedRepoSet provides a set to manage repositories that are on
// unsupported code hosts. This type implements error to allow it to be
// returned directly as an error value if needed.
type UnsupportedRepoSet map[*graphql.Repository]struct{}

func (e UnsupportedRepoSet) Includes(r *graphql.Repository) bool {
	_, ok := e[r]
	return ok
}

func (e UnsupportedRepoSet) Error() string {
	repos := []string{}
	typeSet := map[string]struct{}{}
	for repo := range e {
		repos = append(repos, repo.Name)
		typeSet[repo.ExternalRepository.ServiceType] = struct{}{}
	}

	types := []string{}
	for t := range typeSet {
		types = append(types, t)
	}

	return fmt.Sprintf(
		"found repositories on unsupported code hosts: %s\nrepositories:\n\t%s",
		strings.Join(types, ", "),
		strings.Join(repos, "\n\t"),
	)
}

func (e UnsupportedRepoSet) Append(repo *graphql.Repository) {
	e[repo] = struct{}{}
}

func (e UnsupportedRepoSet) HasUnsupported() bool {
	return len(e) > 0
}

// IgnoredRepoSet provides a set to manage repositories that are on
// unsupported code hosts. This type implements error to allow it to be
// returned directly as an error value if needed.
type IgnoredRepoSet map[*graphql.Repository]struct{}

func (e IgnoredRepoSet) Includes(r *graphql.Repository) bool {
	_, ok := e[r]
	return ok
}

func (e IgnoredRepoSet) Error() string {
	repos := []string{}
	for repo := range e {
		repos = append(repos, repo.Name)
	}

	return fmt.Sprintf(
		"found repositories containing .batchignore files:\n\t%s",
		strings.Join(repos, "\n\t"),
	)
}

func (e IgnoredRepoSet) Append(repo *graphql.Repository) {
	e[repo] = struct{}{}
}

func (e IgnoredRepoSet) HasIgnored() bool {
	return len(e) > 0
}
