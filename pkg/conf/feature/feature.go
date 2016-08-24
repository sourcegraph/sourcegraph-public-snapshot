package feature

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Features is the source of truth for feature toggles. Use Features for
// querying whether a feature is toggled or not
var Features = struct {
	Authors bool // use git blame to show the authors of a def

	GodocRefs bool // redirect from /-/godoc/refs (links constructed by github.com/sourcegraph/gddo fork) to ref pages
	Universe  bool // Universe project
}{
	Authors: true,
}

// IsUniverseRepo returns true if the Universe feature has rolled out to repo.
func IsUniverseRepo(repo string) bool {
	return repoChecker(Features.Universe, os.Getenv("SG_UNIVERSE_REPO"), repo)
}

func init() {
	err := setFeatures(&Features, os.Environ())
	if err != nil {
		// We make this a fatal to prevent a user having a typo when
		// specifing feature toggles
		log.Fatal(err)
	}
}

const envPrefix = "SG_FEATURE_"

// setFeatures expects featureStruct to be a pointer to a simple struct like Features.
func setFeatures(featureStruct interface{}, environ []string) error {
	t := reflect.TypeOf(featureStruct).Elem()
	v := reflect.ValueOf(featureStruct).Elem()
	toggles := make(map[string]reflect.Value, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		toggles[envPrefix+strings.ToUpper(t.Field(i).Name)] = v.Field(i)
	}

	for _, e := range environ {
		pair := strings.SplitN(e, "=", 2)
		key, val := pair[0], pair[1]
		if !strings.HasPrefix(key, envPrefix) {
			continue
		}
		field, ok := toggles[key]
		if !ok {
			log.Printf("warning: Skipping unknown feature toggle %s", key)
			continue
		}
		on, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("Could not parse feature toggle %s", key)
		}
		field.SetBool(on)
	}
	return nil
}

func repoChecker(on bool, enabled, repo string) bool {
	if !on {
		return false
	}
	if enabled == "all" {
		return true
	}
	if enabled == "" {
		// Go testing repos.
		enabled = "github.com/sourcegraph/sourcegraph"
		enabled += ",github.com/slimsag/mux"
		enabled += ",github.com/slimsag/context"
		enabled += ",github.com/slimsag/docker"
		enabled += ",github.com/slimsag/kubernetes"
		enabled += ",github.com/slimsag/prometheus"

		// Java testing repos.
		enabled += ",github.com/slimsag/RxJava"
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
