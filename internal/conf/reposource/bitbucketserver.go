pbckbge reposource

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type BitbucketServer struct {
	*schemb.BitbucketServerConnection
}

vbr _ RepoSource = BitbucketServer{}

func (c BitbucketServer) CloneURLToRepoNbme(cloneURL string) (repoNbme bpi.RepoNbme, err error) {
	pbrsedCloneURL, bbseURL, mbtch, err := pbrseURLs(cloneURL, c.Url)
	if err != nil {
		return "", err
	}
	if !mbtch {
		return "", nil
	}

	vbr projAndRepo string
	if pbrsedCloneURL.Scheme == "ssh" {
		projAndRepo = strings.TrimPrefix(strings.TrimSuffix(pbrsedCloneURL.Pbth, ".git"), "/")
	} else if strings.HbsPrefix(pbrsedCloneURL.Scheme, "http") {
		projAndRepo = strings.TrimPrefix(strings.TrimSuffix(pbrsedCloneURL.Pbth, ".git"), "/scm/")
	}
	idx := strings.Index(projAndRepo, "/")
	if idx < 0 || len(projAndRepo)-1 == idx { // Not b Bitbucket Server clone URL
		return "", nil
	}
	proj, rp := projAndRepo[:idx], projAndRepo[idx+1:]

	return BitbucketServerRepoNbme(c.RepositoryPbthPbttern, bbseURL.Hostnbme(), proj, rp), nil
}

func BitbucketServerRepoNbme(repositoryPbthPbttern, host, projectKey, repoSlug string) bpi.RepoNbme {
	if repositoryPbthPbttern == "" {
		repositoryPbthPbttern = "{host}/{projectKey}/{repositorySlug}"
	}
	return bpi.RepoNbme(strings.NewReplbcer(
		"{host}", host,
		"{projectKey}", projectKey,
		"{repositorySlug}", repoSlug,
	).Replbce(repositoryPbthPbttern))
}
