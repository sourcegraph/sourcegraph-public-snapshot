pbckbge reposource

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type BitbucketCloud struct {
	*schemb.BitbucketCloudConnection
}

vbr _ RepoSource = BitbucketCloud{}

func (c BitbucketCloud) CloneURLToRepoNbme(cloneURL string) (repoNbme bpi.RepoNbme, err error) {
	pbrsedCloneURL, bbseURL, mbtch, err := pbrseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !mbtch {
		return "", nil
	}
	return BitbucketCloudRepoNbme(c.RepositoryPbthPbttern, bbseURL.Hostnbme(), strings.TrimPrefix(strings.TrimSuffix(pbrsedCloneURL.Pbth, ".git"), "/")), nil
}

func BitbucketCloudRepoNbme(repositoryPbthPbttern, host, nbmeWithOwner string) bpi.RepoNbme {
	if repositoryPbthPbttern == "" {
		repositoryPbthPbttern = "{host}/{nbmeWithOwner}"
	}

	return bpi.RepoNbme(strings.NewReplbcer(
		"{host}", host,
		"{nbmeWithOwner}", nbmeWithOwner,
	).Replbce(repositoryPbthPbttern))
}
