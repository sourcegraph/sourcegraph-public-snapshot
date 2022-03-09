package github

import (
	"flag"
	"os"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

// vcrToken is the OAuthBearerToken used for updating VCR fixtures used in tests in this
// package.
//
// Please use the token of the "sourcegraph-vcr" user for GITHUB_TOKEN, which can be found
// in 1Password.
var vcrToken = &auth.OAuthBearerToken{
	Token: os.Getenv("GITHUB_TOKEN"),
}

var gheToken = &auth.OAuthBearerToken{
	Token: os.Getenv("GHE_TOKEN"),
}

var newCache = func(key string, ttl int) Cache { return rcache.NewWithTTL(key, ttl) }

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

// update indicates whether this test's testdata should be updated.
func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}
