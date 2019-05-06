package repos

import (
	"flag"
	"os"
	"regexp"
	"testing"
)

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
