// Package accesstoken is exposed in lib/ for usage in src-cli
package accesstoken

import (
	"sync"

	"github.com/grafana/regexp" // avoid pulling in internal lazyregexp package

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var makePersonalAccessTokenRegex = sync.OnceValue[*regexp.Regexp](func() *regexp.Regexp {
	return regexp.MustCompile("^(?:sgp_|sgph_)?(?:[a-fA-F0-9]{16}_|local_)?([a-fA-F0-9]{40})$")
})

// ParsePersonalAccessToken parses a personal access token to remove prefixes and extract the <token> that is stored in the database
// Personal access tokens can take several forms:
//   - <token>
//   - sgp_<token>
//   - sgp_<instance-identifier>_<token>
func ParsePersonalAccessToken(token string) (string, error) {
	tokenMatches := makePersonalAccessTokenRegex().FindStringSubmatch(token)
	if len(tokenMatches) <= 1 {
		return "", errors.New("invalid token format")
	}
	tokenValue := tokenMatches[1]

	return tokenValue, nil
}
