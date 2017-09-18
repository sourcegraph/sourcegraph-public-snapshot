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
	// DisableSearch disables certain parts of the search experience.
	// This flag is intended to be used for umami deployment.
	DisableSearch bool `json:"disableSearch"`
	// Sep20Auth are the changes to authentication for the September 20th release.
	// Specifically:
	//   - on Sourcegraph.com, authenticated users do not have GitHub tokens
	//   - on Sourcegraph.com, there is only public code
	//   - in an on-prem server, you may add a GitHub personal access token for the github-proxy
	//     to authenticate every request to the GitHub API; this access token may provide
	//     permission to private repos on github.com
	//   - in an on-prem server, you may add an SSH key which provides the privileges necessary
	//     to clone private repositories from github.com
	Sep20Auth bool `json:"sep20Auth"`
}{}

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
