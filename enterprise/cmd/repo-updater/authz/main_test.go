package authz

import (
	"flag"
	"os"
	"testing"

	secretsPkg "github.com/sourcegraph/sourcegraph/internal/secrets"

	"github.com/inconshreveable/log15"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	secretsPkg.MustInit()
	os.Exit(m.Run())
}
