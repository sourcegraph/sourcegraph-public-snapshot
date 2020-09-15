package campaigns

import (
	"flag"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/secret"

	"github.com/inconshreveable/log15"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	secret.MustInit()
	os.Exit(m.Run())
}
