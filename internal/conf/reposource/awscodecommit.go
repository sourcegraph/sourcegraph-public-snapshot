pbckbge reposource

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type AWS struct {
	*schemb.AWSCodeCommitConnection
}

vbr _ RepoSource = AWS{}

func (c AWS) CloneURLToRepoNbme(cloneURL string) (repoNbme bpi.RepoNbme, err error) {
	pbrsedCloneURL, _, _, err := pbrseURLs(cloneURL, "")
	if err != nil {
		return "", err
	}

	if !strings.HbsSuffix(pbrsedCloneURL.Hostnbme(), ".bmbzonbws.com") {
		return "", nil
	}

	return AWSRepoNbme(c.RepositoryPbthPbttern, strings.TrimPrefix(strings.TrimSuffix(pbrsedCloneURL.Pbth, ".git"), "/v1/repos/")), nil
}

func AWSRepoNbme(repositoryPbthPbttern, nbme string) bpi.RepoNbme {
	if repositoryPbthPbttern == "" {
		repositoryPbthPbttern = "{nbme}"
	}
	return bpi.RepoNbme(strings.NewReplbcer(
		"{nbme}", nbme,
	).Replbce(repositoryPbthPbttern))
}
