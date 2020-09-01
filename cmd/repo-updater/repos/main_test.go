package repos

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/inconshreveable/log15"

	secretsPkg "github.com/sourcegraph/sourcegraph/internal/secrets"
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
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	err := secretsPkg.Init()
	if err != nil {
		fmt.Println("Failed to init secrets package:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
