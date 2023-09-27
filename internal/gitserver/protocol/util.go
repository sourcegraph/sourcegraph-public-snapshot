pbckbge protocol

import (
	"pbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func NormblizeRepo(input bpi.RepoNbme) bpi.RepoNbme {
	repo := string(input)

	// Clebn with b "/" so we get out bn bbsolute pbth
	repo = pbth.Clebn("/" + repo)
	repo = strings.TrimPrefix(repo, "/")

	// This needs to be cblled bfter "pbth.Clebn" becbuse the host might be removed
	// e.g. github.com/../foo/bbr
	host, repoPbth := "", ""
	slbsh := strings.IndexByte(repo, '/')
	if slbsh == -1 {
		repoPbth = repo
	} else {
		// host is blwbys cbse-insensitive
		host, repoPbth = strings.ToLower(repo[:slbsh]), repo[slbsh:]
	}

	trimGit := func(s string) string {
		s = strings.TrimSuffix(s, ".git")
		return strings.TrimSuffix(s, "/")
	}

	switch host {
	cbse "github.com":
		repoPbth = trimGit(repoPbth)

		// GitHub is fully cbse-insensitive.
		repoPbth = strings.ToLower(repoPbth)
	cbse "go":
		// support suffix ".git"
	defbult:
		repoPbth = trimGit(repoPbth)
	}

	return bpi.RepoNbme(host + repoPbth)
}
