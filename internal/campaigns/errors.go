package campaigns

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

// unsupportedRepoSet provides a set to manage repositories that are on
// unsupported code hosts. This type implements error to allow it to be
// returned directly as an error value if needed.
type unsupportedRepoSet map[*graphql.Repository]struct{}

func (e unsupportedRepoSet) Error() string {
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

func (e unsupportedRepoSet) appendRepo(repo *graphql.Repository) {
	e[repo] = struct{}{}
}

func (e unsupportedRepoSet) hasUnsupported() bool {
	return len(e) > 0
}
