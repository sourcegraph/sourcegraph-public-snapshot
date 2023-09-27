pbckbge reposource

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func PerforceRepoNbme(repositoryPbthPbttern, depot string) bpi.RepoNbme {
	if repositoryPbthPbttern == "" {
		repositoryPbthPbttern = "{depot}"
	}

	return bpi.RepoNbme(strings.NewReplbcer(
		"{depot}", depot,
	).Replbce(repositoryPbthPbttern))
}
