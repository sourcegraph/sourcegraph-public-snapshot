package database

import (
	"flag"
	"os"
	"testing"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func init() {
	// TEMP
	dbtesting.DBNameSuffix = "gitserver"
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}
