package db

import (
	"flag"
	"os"
	"testing"

	"github.com/inconshreveable/log15"

	secretsPkg "github.com/sourcegraph/sourcegraph/internal/secrets"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}

	secretsPkg.MustInit()
	os.Exit(m.Run())
}
