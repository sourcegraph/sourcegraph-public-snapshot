package universe

import (
	"hash/crc32"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
)

// EnabledFile tells if universe should be used because file should use a universe backend.
func EnabledFile(file string) bool {
	return strings.HasSuffix(file, ".go")
}

// EnabledRepo tells if universe should be used because repo's language should use a universe backend.
func EnabledRepo(repo *sourcegraph.Repo) bool {
	return repo.Language == "Go"
}

// EnabledExcludingBeta is just like Enabled except it excludes users who are
// in the beta program. It should only be used for operations which would
// otherwise affect users not in the universe beta (e.g. data altering
// operations). Effectively it just checks if the given repo is a universe
// repo.
func EnabledExcludingBeta(repo string) bool {
	return repoChecker(feature.Features.Universe, os.Getenv("SG_UNIVERSE_REPO"), repo)
}

var (
	shadowRepoP = getenvPercentage("SG_UNIVERSE_SHADOW_REPO_P")
	shadowP     = getenvPercentage("SG_UNIVERSE_SHADOW_P")
)

// Shadow tells if universe should be sent shadow traffic. If true this means
// that the request is still served by srclib, but the request is also sent to
// universe. SG_UNIVERSE_SHADOW_REPO_P% of repos are considered, of that
// SG_UNIVERSE_SHADOW_P% requests will be shadowed. By default we shadow
// nothing.
func Shadow(repo string) bool {
	if !feature.Features.Universe {
		return false
	}
	if EnabledExcludingBeta(repo) {
		return true
	}
	h := crc32.ChecksumIEEE([]byte(repo))
	if h%100 >= shadowRepoP {
		return false
	}
	return rand.Uint32()%100 < shadowP
}

func getenvPercentage(key string) uint32 {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	p, err := strconv.Atoi(v)
	if err != nil || p < 0 || p > 100 {
		log.Printf("WARNING: env %s needs to be an int in [0, 100], got %s", key, v)
		return 0
	}
	return uint32(p)
}

func repoChecker(on bool, enabled, repo string) bool {
	if !on {
		return false
	}
	if enabled == "all" {
		return true
	}
	if enabled == "" {
		// Java testing repos.
		enabled += "github.com/slimsag/RxJava"
		enabled += ",github.com/slimsag/guava"
		enabled += ",github.com/slimsag/joda-time"

		// JavaScript testing repos.
		enabled += ",github.com/sgtest/javascript-nodejs-sample-0"
		enabled += ",github.com/sgtest/javascript-nodejs-xrefs-0"
		enabled += ",github.com/sgtest/minimal_nodejs_stdlib"
		enabled += ",github.com/sgtest/js-misc"
		enabled += ",github.com/sgtest/js-misc"
		enabled += ",github.com/sgtest/javascript-es6-tests"
	}
	for _, e := range strings.Split(enabled, ",") {
		if repo == e {
			return true
		}
	}
	return false
}
