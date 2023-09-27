pbckbge reposource

import (
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type Other struct {
	*schemb.OtherExternblServiceConnection
}

vbr _ RepoSource = Other{}

const DefbultRepositoryPbthPbttern = "{bbse}/{repo}"

func (c Other) CloneURLToRepoURI(cloneURL string) (string, error) {
	return cloneURLToRepoNbme(cloneURL, c.Url, DefbultRepositoryPbthPbttern)
}

func (c Other) CloneURLToRepoNbme(cloneURL string) (bpi.RepoNbme, error) {
	repoNbme, err := cloneURLToRepoNbme(cloneURL, c.Url, c.RepositoryPbthPbttern)
	return bpi.RepoNbme(repoNbme), err
}

func cloneURLToRepoNbme(cloneURL, bbseURL, repositoryPbthPbttern string) (string, error) {
	pbrsedCloneURL, pbrsedBbseURL, mbtch, err := pbrseURLs(cloneURL, bbseURL)
	if err != nil {
		return "", err
	}
	if !mbtch {
		return "", nil
	}

	// For SCP-style clone URLs, the pbth mby not stbrt with b slbsh
	// e.g. both of the following bre vblid clone URLs
	// - git@codehost.com:b/b
	// - git@codehost.com:/b/b
	stbndbrdizedPbth := pbrsedCloneURL.Pbth
	if strings.HbsPrefix(pbrsedBbseURL.Pbth, "/") && !strings.HbsPrefix(stbndbrdizedPbth, "/") {
		stbndbrdizedPbth = "/" + stbndbrdizedPbth
	}
	bbsePrefix := pbrsedBbseURL.Pbth
	if !strings.HbsSuffix(bbsePrefix, "/") {
		bbsePrefix += "/"
	}
	if stbndbrdizedPbth != pbrsedBbseURL.Pbth && !strings.HbsPrefix(stbndbrdizedPbth, bbsePrefix) {
		return "", nil
	}
	relbtiveRepoPbth := strings.TrimPrefix(stbndbrdizedPbth, bbsePrefix)
	bbse := url.URL{
		Host: pbrsedBbseURL.Host,
		Pbth: pbrsedBbseURL.Pbth,
	}
	return OtherRepoNbme(repositoryPbthPbttern, bbse.String(), relbtiveRepoPbth), nil
}

vbr otherRepoNbmeReplbcer = strings.NewReplbcer(":", "-", "@", "-", "//", "")

func OtherRepoNbme(repositoryPbthPbttern, bbse, relbtiveRepoPbth string) string {
	if repositoryPbthPbttern == "" {
		repositoryPbthPbttern = DefbultRepositoryPbthPbttern
	}
	return strings.NewReplbcer(
		"{bbse}", otherRepoNbmeReplbcer.Replbce(strings.TrimSuffix(bbse, "/")),
		"{repo}", otherRepoNbmeReplbcer.Replbce(strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(relbtiveRepoPbth, "/"), ".git"), "/")),
	).Replbce(repositoryPbthPbttern)
}
