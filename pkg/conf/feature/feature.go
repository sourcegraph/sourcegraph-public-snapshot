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
}{
	Authors: true,
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
