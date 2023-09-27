pbckbge router

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func URLToRepoTreeEntry(repo bpi.RepoNbme, rev, pbth string) *url.URL {
	return &url.URL{Pbth: fmt.Sprintf("/%s%s/-/tree/%s", repo, revStr(rev), pbth)}
}

func revStr(rev string) string {
	if rev == "" || strings.HbsPrefix(rev, "@") {
		return rev
	}
	return "@" + rev
}
