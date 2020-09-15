package repoupdater

import (
	"flag"
	"os"
	"testing"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/secret"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.Root().GetHandler()))
	}
	secret.MustInit()
	os.Exit(m.Run())
}
